## "sage sis.sage" will generate test_cases.json
## tested with a fresh sage install on macOS (Feb 2023)

import json
import random

# mamabear FR
# 2^49 - 2^34 + 1
R = 2**49-2**34+1
FR_BYTE_SIZE = 8
FR_BIT_SIZE = FR_BYTE_SIZE*8
GFR = GF(R)
FR.<x> = GFR[]
Z = IntegerRing()

# Montgomery constant.
# NOTE: this is NOT 2^(FR_BYTE_SIZE*8) as it is for the other fields. MamaBear
# uses a Montgomery radix of 2^52 - chosen to match the AVX-512IFMA VPMADD52
# operand width - which is narrower than the 64-bit word an element is stored
# in. Using 2^64 here silently produces the wrong reference vectors.
RR = GFR(2**52)

# utils


def build_poly(a):
    """ Builds a poly from the array a

    Args:
        a an array

    Returns:
        a[0]+a[1]*X + .. + a[n]*X**n
    """

    res = GFR(0)
    for i, v in enumerate(a):
        res += GFR(v)*x**i
    return res


def get_ith_bit(i, b):
    """
    Args:
        i: index of the bit to retrieve
        b: array of bytes

    Returns:
        the i-th bit of b, when it is written b[0] || b[1] || ...
    """
    k = i//8
    if k >= len(b):
        return 0
    j = i % 8
    return (b[k] >> (7-j)) & 1


def to_bytes(m, s):
    """

    Args:
        m: an integer
        s: the expected number of bytes of the result. If s is bigger than the
        number of bytes in m, the remaining bytes are set to zero.

    Returns:
        the byte representation of m as a byte array, as
        in gnark-crypto.
    """
    _m = Z(m)
    res = s*[0]
    mask = 255
    for i in range(s):
        res[s-1-i] = _m & 255
        _m = _m >> 8
    return res


def split_coeffs(b, logTwoBound):
    """
    Args:
        b: an array of bytes
        logTwoBound: number of bits of the bound

    Returns:
        an array of coeffs, each coeff being the i-th chunk of logTwoBounds bits of b.
        The coeffs are formed as follow. The input byte string is implicitly parsed as
        a slice of field elements of FR_BYTE_SIZE bytes each in bigendian-natural form. the outputs
        are in a little-endian form. That is, each chunk of size FR_BIT_SIZE / logTwoBounds of the
        output can be seen as a polynomial, such that, when evaluated at 2 we get the original
        field element.
    """
    nbBits = len(b)*8
    res = [] 
    i = 0

    if len(b) % FR_BYTE_SIZE != 0:
        exit("the length of b should divide the field size")

    # The number of fields that we are parsing. In case we have that
    # logTwoBound does not divide the number of bits to represent a
    # field element, we do not merge them.
    nbField = len(b) / FR_BYTE_SIZE
    nbBitsInField = int(FR_BYTE_SIZE * 8)
    
    for fieldID in range(nbField):
        fieldStart = fieldID * FR_BIT_SIZE
        e = 0
        for bitInField in range(nbBitsInField):
            j = bitInField % logTwoBound
            at = fieldStart + nbBitsInField - 1 - bitInField
            e |= get_ith_bit(at, b) << j 
            # Switch to a new limb
            if j == logTwoBound - 1 or bitInField == FR_BYTE_SIZE * 8 - 1:
                res.append(e)
                e = 0

    # careful Montgomery constant...
    return [GFR(e)*RR**-1 for e in res]


def poly_pseudo_rand(seed, n):
    """ Generates a pseudo random polynomial of size n from seed.

    Args:
        seed: seed for the pseudo random gen
        n: degree of the polynomial
    """
    seed = GFR(seed)
    a = n*[0]
    for i in range(n):
        a[i] = seed**2
        seed = a[i]
    return build_poly(a)


# SIS
class SIS:
    def __init__(self, seed, logTwoDegree, logTwoBound, maxNbElementsToHash):
        """
            Args:
                seed
                logTwoDegree: 
                logTwoBound: bound of SIS
                maxNbElementsToHash
        """
        capacity = maxNbElementsToHash * FR_BYTE_SIZE
        degree = 1 << logTwoDegree

        n = capacity * 8 / logTwoBound  # number of coefficients
        if n % degree == 0:  # check how sage / python rounds the int div.
            n = n / degree
        else:
            n = n / degree
            n = n + 1

        n = int(n)

        self.logTwoBound = logTwoBound
        self.degree = degree
        self.size = n
        self.key = n * [0]
        for i in range(n):
            self.key[i] = poly_pseudo_rand(seed, self.degree)
            seed += 1

    def hash(self, inputs):
        """ 
        Args:
           inputs is a vector of FR elements

        Returns:
            the sis hash of m.
        """
        b = []
        for i in inputs:
            b.extend(to_bytes(i, FR_BYTE_SIZE))

        return self.hash_bytes(b)

    def hash_bytes(self, b):
        """ 
        Args:
            b is a list of bytes to hash

        Returns:
            the sis hash of m.
        """
        # step 1: build the polynomials from m
        c = split_coeffs(b, self.logTwoBound)
        mp = [build_poly(c[self.degree*i:self.degree*(i+1)])
              for i in range(self.size)]

        # step 2: compute sum_i mp[i]*key[i] mod X^n+1
        modulo = x**self.degree+1
        res = 0
        for i in range(self.size):
            res += self.key[i]*mp[i]
        res = res % modulo
        return res


def vectorToString(v):
    # v is a vector of field elements
    # we return a list of strings in base10
    r = []
    for e in v:
        r.append("{}".format(Z(e)))
        # r.append("0x" + Z(e).hex())
    return r
    

def SISParams(seed, logTwoDegree, logTwoBound, maxNbElementsToHash):
    p = {}
    p['seed'] = int(seed)
    p['logTwoDegree'] = int(logTwoDegree)
    p['logTwoBound'] = int(logTwoBound)
    p['maxNbElementsToHash'] = int(maxNbElementsToHash)
    return p

PARAMS = [
]

bounds = [8, 16]
degrees = [5,6,7,8,9]

for bound in bounds:
    for degree in degrees:
        PARAMS.append(SISParams(5, degree, bound, 260))

def random_inputs(size, modulus):
    return [GFR(random.randint(0, modulus - 1)) for _ in range(size)]

INPUTS = random_inputs(260, R)

TEST_CASES = {}

TEST_CASES['inputs'] = vectorToString(INPUTS)
TEST_CASES['entries'] = []

for p in PARAMS:

    entry = {}
    entry['params'] = p 
    # print("generating test cases with SIS params " + json.dumps(p))
    instance = SIS(p['seed'], p['logTwoDegree'], p['logTwoBound'], p['maxNbElementsToHash'])

    # hash the vector
    hResult = instance.hash(INPUTS)
    entry['expected'] = vectorToString(hResult)
    
    TEST_CASES['entries'].append(entry)


TEST_CASES_json = json.dumps(TEST_CASES, indent=4)
with open("test_cases.json", "w") as outfile:
    outfile.write(TEST_CASES_json)
