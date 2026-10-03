// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package hash_to_curve

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/consensys/gnark-crypto/internal/generator/config"
)

// chainCode returns the body of a Go function
//
//	func f(dst *T, x *T[, y *T])
//
// that evaluates the polynomial described by the multiplication chain c at x and
// stores the result in dst (multiplied by y when mulBy == "y"). The code is straight
// line: no branches, no divisions, and every multiplication of the chain is a field
// Mul (or Square when both factors coincide). The small integer coefficients of the
// linear forms are realised with Double/Add/Sub. The result is accumulated in locals
// so dst may alias x or y, as it does in the isogeny map.
//
// constsVar names the array holding c.Constants and leadVar the leading coefficient
// (used only when the polynomial is not monic).
func chainCode(c *config.PolyChainInfo, coordType, constsVar, leadVar, mulBy string) string {
	e := &chainEmitter{constsVar: constsVar}

	for i, g := range c.Gates {
		out := chainVar{name: fmt.Sprintf("w[%d]", i), addr: fmt.Sprintf("&w[%d]", i)}
		if sameForm(g.Left, g.Right) {
			src := e.linear(g.Left, out)
			e.emit("%s.Square(%s)", out.name, src)
			continue
		}
		l := e.linear(g.Left, out)
		r := e.linear(g.Right, chainVar{name: "r", addr: "&r"})
		e.emit("%s.Mul(%s, %s)", out.name, l, r)
	}

	res := e.linear(c.Output, chainVar{name: "p", addr: "&p"})
	switch {
	case c.Leading == nil && mulBy == "":
		e.emit("dst.Set(%s)", res)
	case c.Leading == nil:
		e.emit("dst.Mul(%s, %s)", res, mulBy)
	case mulBy == "":
		e.emit("dst.Mul(%s, &%s)", res, leadVar)
	default:
		e.emit("p.Mul(%s, &%s)", res, leadVar)
		e.emit("dst.Mul(&p, %s)", mulBy)
		e.used["p"] = true
	}

	var decl []string
	decl = append(decl, fmt.Sprintf("var w [%d]%s // gate outputs", len(c.Gates), coordType))
	var tmps []string
	for _, v := range []string{"p", "r", "t"} {
		if e.used[v] {
			tmps = append(tmps, v)
		}
	}
	if len(tmps) > 0 {
		decl = append(decl, fmt.Sprintf("var %s %s", strings.Join(tmps, ", "), coordType))
	}
	return "\t" + strings.Join(append(decl, e.lines...), "\n\t")
}

type chainVar struct{ name, addr string }

type chainEmitter struct {
	constsVar string
	lines     []string
	used      map[string]bool
}

func (e *chainEmitter) emit(format string, a ...any) {
	e.lines = append(e.lines, fmt.Sprintf(format, a...))
}

func (e *chainEmitter) use(v chainVar) {
	if e.used == nil {
		e.used = make(map[string]bool)
	}
	e.used[v.name] = true
}

func wireRef(i int) string {
	if i == 0 {
		return "x"
	}
	return fmt.Sprintf("&w[%d]", i-1)
}

func sameForm(a, b config.LinearForm) bool {
	if a.Const != b.Const || len(a.Terms) != len(b.Terms) {
		return false
	}
	for i := range a.Terms {
		if a.Terms[i] != b.Terms[i] {
			return false
		}
	}
	return true
}

// smallMul emits target = k * src for |k| >= 2 by double-and-add.
func (e *chainEmitter) smallMul(target chainVar, src string, k int) {
	e.use(target)
	neg := k < 0
	if neg {
		k = -k
	}
	bits := strconv.FormatInt(int64(k), 2)
	e.emit("%s.Double(%s)", target.name, src)
	if bits[1] == '1' {
		e.emit("%s.Add(%s, %s)", target.name, target.addr, src)
	}
	for _, b := range bits[2:] {
		e.emit("%s.Double(%s)", target.name, target.addr)
		if b == '1' {
			e.emit("%s.Add(%s, %s)", target.name, target.addr, src)
		}
	}
	if neg {
		e.emit("%s.Neg(%s)", target.name, target.addr)
	}
}

// linear emits the evaluation of the linear form into target and returns the
// expression (a pointer) holding its value. A form that is a single wire with
// coefficient 1 is returned by reference without emitting anything.
func (e *chainEmitter) linear(lf config.LinearForm, target chainVar) string {
	type term struct {
		src string
		k   int
	}
	var ops []term
	for _, t := range lf.Terms {
		if t[1] == 0 || t[1] < -64 || t[1] > 64 {
			panic(fmt.Sprintf("unsupported wire coefficient %d", t[1]))
		}
		ops = append(ops, term{wireRef(t[0]), t[1]})
	}
	if lf.Const >= 0 {
		ops = append(ops, term{fmt.Sprintf("&%s[%d]", e.constsVar, lf.Const), 1})
	}
	rank := func(k int) int {
		switch k {
		case 1:
			return 0
		case -1:
			return 1
		}
		return 2
	}
	sort.SliceStable(ops, func(i, j int) bool { return rank(ops[i].k) < rank(ops[j].k) })
	if len(ops) == 0 {
		panic("empty linear form")
	}
	if len(ops) == 1 && ops[0].k == 1 {
		return ops[0].src
	}

	e.use(target)
	i := 1
	switch {
	case ops[0].k == 1 && len(ops) > 1 && ops[1].k == 1:
		e.emit("%s.Add(%s, %s)", target.name, ops[0].src, ops[1].src)
		i = 2
	case ops[0].k == 1 && len(ops) > 1 && ops[1].k == -1:
		e.emit("%s.Sub(%s, %s)", target.name, ops[0].src, ops[1].src)
		i = 2
	case ops[0].k == 1:
		e.emit("%s.Set(%s)", target.name, ops[0].src)
	case ops[0].k == -1:
		e.emit("%s.Neg(%s)", target.name, ops[0].src)
	default:
		e.smallMul(target, ops[0].src, ops[0].k)
	}
	tmp := chainVar{name: "t", addr: "&t"}
	for ; i < len(ops); i++ {
		switch {
		case ops[i].k == 1:
			e.emit("%s.Add(%s, %s)", target.name, target.addr, ops[i].src)
		case ops[i].k == -1:
			e.emit("%s.Sub(%s, %s)", target.name, target.addr, ops[i].src)
		case ops[i].k > 0:
			e.smallMul(tmp, ops[i].src, ops[i].k)
			e.emit("%s.Add(%s, %s)", target.name, target.addr, tmp.addr)
		default:
			e.smallMul(tmp, ops[i].src, -ops[i].k)
			e.emit("%s.Sub(%s, %s)", target.name, target.addr, tmp.addr)
		}
	}
	return target.addr
}
