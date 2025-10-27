package idgen

import "fmt"

// feistel implements a k-bit Feistel network as an Obfuscator.
// It uses simple ARX-style round functions with fixed, non-secret constants.
type feistel struct {
	k      uint // domain bits
	rounds int
	lBits  uint // left half bits (initial)
	rBits  uint // right half bits (initial)
	mask   uint64
}

// NewFeistel creates a k-bit Feistel obfuscator with the given number of rounds.
// k must be in [1, 63], rounds >= 2.
func NewFeistel(k uint, rounds int) (Obfuscator, error) {
	if k == 0 || k > 63 {
		return nil, fmt.Errorf("feistel: k out of range: %d", k)
	}
	if rounds < 2 {
		return nil, fmt.Errorf("feistel: rounds must be >= 2")
	}
	l := (k + 1) / 2 // favor left = ceil(k/2)
	r := k - l
	return &feistel{
		k:      k,
		rounds: rounds,
		lBits:  l,
		rBits:  r,
		mask:   (uint64(1) << k) - 1,
	}, nil
}

func (f *feistel) DomainBits() uint { return f.k }

// simpleFbits maps an input of inBits to an output of outBits using ARX ops.
func (f *feistel) simpleFbits(input uint64, inBits, outBits uint, round int) uint64 {
	constC := [...]uint64{0x9E3779B97F4A7C15, 0xBF58476D1CE4E5B9, 0x94D049BB133111EB, 0xD6E8FEB86659FD93, 0xA24BAED4963EE407, 0x9FB21C651E98DF25}
	c := constC[round%len(constC)]
	inMask := (uint64(1) << inBits) - 1
	outMask := (uint64(1) << outBits) - 1
	x := (input & inMask) ^ (c & inMask)
	x = (x*0x5bd1e995 + 0x27d4eb2d) & outMask
	if outBits > 1 {
		rot := (5 + uint(round)) % outBits
		x = ((x << rot) | (x >> (outBits - rot))) & outMask
	}
	x ^= (x >> 3) & outMask
	return x & outMask
}

func (f *feistel) Obfuscate(x uint64) uint64 {
	x &= f.mask
	lMask := (uint64(1) << f.lBits) - 1
	rMask := (uint64(1) << f.rBits) - 1
	// Initial halves: L has lBits (low), R has rBits (high)
	L := x & lMask
	R := (x >> f.lBits) & rMask
	for i := 0; i < f.rounds; i++ {
		if i%2 == 0 {
			// Map R (rBits) -> lBits
			fn := f.simpleFbits(R, f.rBits, f.lBits, i)
			L, R = R, (L^fn)&lMask // new L has rBits, new R has lBits
		} else {
			// Map R (lBits) -> rBits (sizes swapped after previous round)
			fn := f.simpleFbits(R, f.lBits, f.rBits, i)
			L, R = R, (L^fn)&rMask // new L has lBits, new R has rBits
		}
	}
	// Pack depending on parity: after even rounds, halves return to original sizes.
	if f.rounds%2 == 0 {
		// L has lBits, R has rBits
		return (R << f.lBits) | (L & lMask)
	}
	// After odd rounds: L has rBits, R has lBits
	return (R << f.rBits) | (L & rMask)
}

func (f *feistel) Deobfuscate(y uint64) uint64 {
	y &= f.mask
	lMask := (uint64(1) << f.lBits) - 1
	rMask := (uint64(1) << f.rBits) - 1
	// Unpack current halves depending on parity of rounds
	var L, R uint64
	if f.rounds%2 == 0 {
		// L has lBits (low), R has rBits (high)
		L = y & lMask
		R = (y >> f.lBits) & rMask
	} else {
		// L has rBits (low), R has lBits (high)
		L = y & rMask
		R = (y >> f.rBits) & lMask
	}
	for i := f.rounds - 1; i >= 0; i-- {
		if i%2 == 0 {
			// Inverse of even forward round: previous Rp=rBits, Lp=lBits
			Rp := L // rBits
			fn := f.simpleFbits(Rp, f.rBits, f.lBits, i)
			Lp := (R ^ fn) & lMask
			L, R = Lp, Rp
		} else {
			// Inverse of odd forward round: previous Rp=lBits, Lp=rBits
			Rp := L // lBits
			fn := f.simpleFbits(Rp, f.lBits, f.rBits, i)
			Lp := (R ^ fn) & rMask
			L, R = Lp, Rp
		}
	}
	// Pack original arrangement: L=lBits (low), R=rBits (high)
	return (R << f.lBits) | (L & lMask)
}
