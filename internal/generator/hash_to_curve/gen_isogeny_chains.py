#!/usr/bin/env python3
"""Offline preprocessing of the hash-to-curve isogeny polynomials into short
multiplication chains.

A monic polynomial of degree n over a field of characteristic 0 or p > n can be
evaluated with floor(n/2)+1 field multiplications (instead of Horner's n-1) after a
one-time *rational preprocessing* of its coefficients:

    T. D. Ahle, "Fast Evaluation of Polynomials with Rational Preprocessing",
    https://arxiv.org/abs/2609.06022, https://thomasahle.com/fast-polynomials/

This script reads the isogeny coefficient tables of one curve from
internal/generator/config/<curve>.go, runs the paper's decoder (tools/polychain.py
from https://github.com/thomasahle/fast-polynomials) on every map of degree > 6,
verifies each chain against the original polynomial at random points, and writes
the chain data as Go literals to internal/generator/config/<curve>_isogeny_chains.go.
The code generator (template pkg_sswu.go.tmpl) turns that data into straight-line,
branch-free Go.

Usage:
    python3 gen_isogeny_chains.py --polychain-dir <fast-polynomials>/tools \
        --config internal/generator/config/bw6-761.go --suite HashE2 --var bw6761G2IsogenyChains

Prime fields and quadratic extensions Fp[u]/(u^2 - beta) (beta = CoordExtRoot in
the curve config) are supported.  Non-monic numerators are handled by dividing by
the leading coefficient before decoding and multiplying the chain's result by it
(one extra multiplication).  Small integer wire coefficients in the chains
(e.g. -12 .. 2) are realised by doublings and additions, never by a field
multiplication.
"""
import argparse
import json
import os
import random
import re
import subprocess
import sys


def parse_args():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument('--polychain-dir', required=True, help='directory containing polychain.py and poly_schedule.py')
    ap.add_argument('--config', required=True, help='path to internal/generator/config/<curve>.go')
    ap.add_argument('--suite', required=True, choices=['HashE1', 'HashE2'])
    ap.add_argument('--var', required=True, help='Go variable name for the emitted IsogenyChains value')
    ap.add_argument('--out', help='output .go file (default: <config dir>/<curve>_isogeny_chains.go)')
    ap.add_argument('--min-degree', type=int, default=7, help='maps of lower degree keep Horner (default 7)')
    ap.add_argument('--checks', type=int, default=200, help='random evaluation points per chain (default 200)')
    ap.add_argument('--seed', type=int, default=1)
    return ap.parse_args()


args = parse_args()
sys.path.insert(0, args.polychain_dir)
import poly_schedule as ps  # noqa: E402
import polychain as pc  # noqa: E402

random.seed(args.seed)
src = open(args.config).read()


# --------------------------------------------------------------------------- config parsing

def go_int(s):
    return int(s, 0)


P = go_int(re.search(r'FpModulus:\s*"([^"]+)"', src).group(1))


def block(text, start):
    """text[start:] must begin with '{'; return the brace-balanced block (inclusive)."""
    assert text[start] == '{', text[start:start + 20]
    depth = 0
    for i in range(start, len(text)):
        if text[i] == '{':
            depth += 1
        elif text[i] == '}':
            depth -= 1
            if depth == 0:
                return text[start:i + 1]
    raise SystemExit('unbalanced braces in config')


def go_string_matrix(text):
    """the elements of a [][]string literal, as lists of Python ints"""
    inner = block(text, text.index('{'))[1:-1]
    rows = []
    for m in re.finditer(r'\{([^{}]*)\}', inner):
        rows.append([go_int(v) for v in re.findall(r'"([^"]+)"', m.group(1))])
    return rows


suite_start = src.index(args.suite + ':')
suite = block(src, src.index('{', suite_start))

# point definition of this suite (G1 for HashE1, G2 for HashE2)
point_name = 'g1' if args.suite == 'HashE1' else 'g2'
point_blocks = [block(src, m.start()) for m in re.finditer(r'\{\s*\n\s*CoordType:', src)]
point = next(b for b in point_blocks if re.search(r'PointName:\s*"%s"' % point_name, b))
ext_degree = int(re.search(r'CoordExtDegree:\s*(\d+)', point).group(1))
m = re.search(r'CoordExtRoot:\s*(-?\d+)', point)
beta = int(m.group(1)) if m else None
assert ext_degree in (1, 2), 'only Fp and Fp^2 coordinates are supported'
if ext_degree == 2:
    assert beta is not None

iso_start = suite.index('Isogeny:')
iso = block(suite, suite.index('{', iso_start))


def poly_table(map_name, part):
    m_start = iso.index(map_name + ':')
    mp = block(iso, iso.index('{', m_start))
    p_start = mp.index(part + ':')
    return go_string_matrix(mp[p_start:])


# --------------------------------------------------------------------------- fields

class Fp(ps.Field):
    """GF(p) with elements represented as 1-tuples so both field degrees share code."""
    d = 1

    def __init__(self):
        super().__init__(modulus=P, use_fractions=False)

    def coerce(self, x):
        if isinstance(x, tuple):
            return (x[0] % P,)
        return (int(x) % P,)

    def zero(self): return (0,)
    def one(self): return (1,)
    def add(self, a, b): a, b = self.coerce(a), self.coerce(b); return ((a[0] + b[0]) % P,)
    def sub(self, a, b): a, b = self.coerce(a), self.coerce(b); return ((a[0] - b[0]) % P,)
    def neg(self, a): a = self.coerce(a); return ((-a[0]) % P,)
    def mul(self, a, b): a, b = self.coerce(a), self.coerce(b); return ((a[0] * b[0]) % P,)

    def inv(self, a):
        a = self.coerce(a)
        if a[0] == 0:
            raise ZeroDivisionError
        return (pow(a[0], -1, P),)

    def div(self, a, b): return self.mul(a, self.inv(b))
    def sqrt(self, a): raise NotImplementedError
    def is_zero(self, a): return self.coerce(a) == (0,)
    def random(self): return (random.randrange(P),)


class Fp2(ps.Field):
    """GF(p^2) = Fp[u]/(u^2 - beta), elements as pairs (a0, a1) = a0 + a1 u."""
    d = 2

    def __init__(self, beta):
        super().__init__(modulus=P, use_fractions=False)
        self.beta = beta % P

    def coerce(self, x):
        if isinstance(x, tuple):
            return (x[0] % P, x[1] % P)
        return (int(x) % P, 0)

    def zero(self): return (0, 0)
    def one(self): return (1, 0)
    def add(self, a, b): a, b = self.coerce(a), self.coerce(b); return ((a[0] + b[0]) % P, (a[1] + b[1]) % P)
    def sub(self, a, b): a, b = self.coerce(a), self.coerce(b); return ((a[0] - b[0]) % P, (a[1] - b[1]) % P)
    def neg(self, a): a = self.coerce(a); return ((-a[0]) % P, (-a[1]) % P)

    def mul(self, a, b):
        a, b = self.coerce(a), self.coerce(b)
        return ((a[0] * b[0] + self.beta * a[1] * b[1]) % P, (a[0] * b[1] + a[1] * b[0]) % P)

    def inv(self, a):
        a = self.coerce(a)
        norm = (a[0] * a[0] - self.beta * a[1] * a[1]) % P
        if norm == 0:
            raise ZeroDivisionError
        ni = pow(norm, -1, P)
        return ((a[0] * ni) % P, (-a[1] * ni) % P)

    def div(self, a, b): return self.mul(a, self.inv(b))
    def sqrt(self, a): raise NotImplementedError
    def is_zero(self, a): return self.coerce(a) == (0, 0)
    def random(self): return (random.randrange(P), random.randrange(P))


F = Fp() if ext_degree == 1 else Fp2(beta)

# polychain's rational constants come back as bare ints; route them through coerce
_orig_f2f = pc._fraction_to_field
pc._fraction_to_field = lambda fr, field: field.coerce(_orig_f2f(fr, field))


def elem(coords):
    """config coordinate list -> field element"""
    assert len(coords) == F.d, coords
    return F.coerce(tuple(coords))


def poly_eval(c, x):
    acc = F.zero()
    for cv in reversed(c):
        acc = F.add(F.mul(acc, x), cv)
    return acc


# --------------------------------------------------------------------------- chains

TERM = re.compile(r'^(-?)(\d+)?\*?(a\d+)?$')


def eval_const(expr, keys):
    """linear expression in the keys a_i with integer coefficients -> field element"""
    s = expr.replace(' ', '').replace('-', '+-')
    tot = F.zero()
    for t in s.split('+'):
        if not t:
            continue
        m = TERM.match(t)
        assert m, (expr, t)
        sign = -1 if m.group(1) else 1
        coef = int(m.group(2)) if m.group(2) else 1
        val = keys[int(m.group(3)[1:])] if m.group(3) else F.one()
        tot = F.add(tot, F.mul(F.coerce(sign * coef), val))
    return tot


def build_chain(name, coeffs):
    """coeffs: c_0..c_n (c_n = leading coefficient).  Returns the Go literal data
    (dict) or None when the degree is below the threshold."""
    n = len(coeffs) - 1
    if n < args.min_degree:
        return None, n, None
    lead = coeffs[n]
    monic = lead == F.one()
    inv = F.inv(lead)
    mon = [F.mul(v, inv) for v in coeffs[:n]]
    keys = pc.decode(n, mon, F)
    back = [F.coerce(v) for v in pc.encode(n, keys, F)]
    assert back == mon, 'decode/encode round trip failed for ' + name
    r = subprocess.run([sys.executable, os.path.join(args.polychain_dir, 'polychain.py'), 'chain', str(n), '--json'],
                       capture_output=True, text=True)
    if r.returncode:
        raise SystemExit(r.stderr)
    ch = json.loads(r.stdout)
    assert ch['n'] == n

    wires = [w for w in ch['wires'] if w != '1']
    assert wires[0] == 'x'
    widx = {w: i for i, w in enumerate(wires)}
    consts, cidx = [], {}

    def const_index(val):
        if F.is_zero(val):
            return -1
        if val not in cidx:
            cidx[val] = len(consts)
            consts.append(val)
        return cidx[val]

    def linform(lf):
        assert '1' not in lf['terms'], lf
        terms = [[widx[w], int(k)] for w, k in lf['terms'].items()]
        for _, k in terms:
            assert 1 <= abs(k) <= 64, 'wire coefficient %d out of the supported range' % k
        return {'Terms': terms, 'Const': const_index(eval_const(lf['const'], keys))}

    gates = []
    for i, g in enumerate(ch['gates']):
        assert widx[g['out']] == i + 1, 'gate outputs must be numbered in order'
        gates.append({'Left': linform(g['left']), 'Right': linform(g['right'])})
    output = linform(ch['output'])
    data = {
        'Degree': n,
        'Leading': None if monic else lead,
        'Constants': consts,
        'Gates': gates,
        'Output': output,
    }

    # self-check: evaluate the *emitted* data structure against the polynomial
    def lin(lf, w):
        tot = consts[lf['Const']] if lf['Const'] >= 0 else F.zero()
        for wi, k in lf['Terms']:
            tot = F.add(tot, F.mul(F.coerce(k), w[wi]))
        return tot

    def evaluate(x):
        w = [x]
        for g in gates:
            w.append(F.mul(lin(g['Left'], w), lin(g['Right'], w)))
        out = lin(output, w)
        return out if monic else F.mul(out, lead)

    pts = [F.zero(), F.one(), F.neg(F.one())] + [F.random() for _ in range(args.checks)]
    for x in pts:
        assert evaluate(x) == poly_eval(coeffs, x), 'chain self-check failed for ' + name
    nmul = len(gates) + (0 if monic else 1)
    horner = n - 1 if monic else n
    print('  %-4s degree %2d %-10s Horner %3d mults -> chain %3d mults (%d random points + 0, 1, -1 verified)'
          % (name, n, '' if monic else '(non-monic)', horner, nmul, args.checks))
    return data, n, (horner, nmul)


x_num = [elem(c) for c in poly_table('XMap', 'Num')]
x_den = [elem(c) for c in poly_table('XMap', 'Den')] + [F.one()]   # Den omits the leading 1
y_num = [elem(c) for c in poly_table('YMap', 'Num')]
y_den = [elem(c) for c in poly_table('YMap', 'Den')] + [F.one()]

print('%s %s over %s (p has %d bits)' % (os.path.basename(args.config), args.suite,
                                          'Fp' if F.d == 1 else 'Fp^2 (u^2 = %d)' % beta, P.bit_length()))
results = {}
tot_h = tot_c = 0
for field_name, coeffs in (('XNum', x_num), ('XDen', x_den), ('YNum', y_num), ('YDen', y_den)):
    data, n, counts = build_chain(field_name, coeffs)
    results[field_name] = data
    if counts:
        tot_h += counts[0]
        tot_c += counts[1]
    else:
        print('  %-4s degree %2d: kept Horner (degree < %d)' % (field_name, n, args.min_degree))
if not any(results.values()):
    raise SystemExit('no map of degree >= %d; nothing to emit' % args.min_degree)
print('  total multiplications on chained maps: Horner %d -> chains %d' % (tot_h, tot_c))


# --------------------------------------------------------------------------- Go output

def go_elem(v):
    return '[]string{' + ', '.join('"%d"' % c for c in v) + '}'


def go_linform(lf, indent):
    terms = ', '.join('{%d, %d}' % (w, k) for w, k in lf['Terms'])
    return '%sLinearForm{Terms: [][2]int{%s}, Const: %d}' % (indent, terms, lf['Const'])


out = []
out.append('// Copyright 2020-2026 Consensys Software Inc.')
out.append('// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.')
out.append('')
out.append('// Generated by internal/generator/hash_to_curve/gen_isogeny_chains.py from %s (%s);' % (os.path.basename(args.config), args.suite))
out.append('// do not edit by hand. Regenerate with')
out.append('//   python3 internal/generator/hash_to_curve/gen_isogeny_chains.py --polychain-dir <fast-polynomials>/tools \\')
out.append('//     --config %s --suite %s --var %s' % (os.path.relpath(args.config), args.suite, args.var))
out.append('//')
out.append('// Multiplication chains for the isogeny polynomials, preprocessed offline with the')
out.append('// decoder of "Fast Evaluation of Polynomials with Rational Preprocessing"')
out.append('// (https://arxiv.org/abs/2609.06022). Each chain evaluates the polynomial with')
out.append('// floor(n/2)+1 multiplications (+1 for a non-monic leading coefficient).')
out.append('')
out.append('package config')
out.append('')
out.append('var %s = IsogenyChains{' % args.var)
for field_name in ('XNum', 'XDen', 'YNum', 'YDen'):
    data = results[field_name]
    if data is None:
        out.append('\t%s: nil, // degree < %d: Horner' % (field_name, args.min_degree))
        continue
    out.append('\t%s: &PolyChain{' % field_name)
    out.append('\t\tDegree: %d,' % data['Degree'])
    if data['Leading'] is not None:
        out.append('\t\tLeading: %s,' % go_elem(data['Leading']))
    out.append('\t\tConstants: [][]string{')
    for c in data['Constants']:
        out.append('\t\t\t%s,' % go_elem(c))
    out.append('\t\t},')
    out.append('\t\tGates: []ChainGate{')
    for g in data['Gates']:
        out.append('\t\t\t{')
        out.append('\t\t\t\tLeft:  %s,' % go_linform(g['Left'], ''))
        out.append('\t\t\t\tRight: %s,' % go_linform(g['Right'], ''))
        out.append('\t\t\t},')
    out.append('\t\t},')
    out.append('\t\tOutput: %s,' % go_linform(data['Output'], ''))
    out.append('\t},')
out.append('}')
out.append('')

out_path = args.out or os.path.join(os.path.dirname(args.config),
                                    os.path.basename(args.config)[:-3] + '_%s_isogeny_chains.go' % point_name)
open(out_path, 'w').write('\n'.join(out))
subprocess.run(['gofmt', '-w', out_path], check=True)
print('  wrote', os.path.relpath(out_path))
