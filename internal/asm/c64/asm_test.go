// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package c64_test

import (
	"testing"

	"golang.org/x/exp/rand"

	"gonum.org/v1/gonum/cmplxs"
	"gonum.org/v1/gonum/floats/scalar"
	"gonum.org/v1/gonum/internal/cmplx64"
	"gonum.org/v1/gonum/internal/math32"
)

const (
	msgVal   = "%v: unexpected value at %v Got: %v Expected: %v"
	msgGuard = "%v: Guard violated in %s vector %v %v"
)

var (
	nan = math32.NaN()

	cnan = cmplx64.NaN()
	cinf = cmplx64.Inf()
)

// TODO(kortschak): Harmonise the situation in asm/{f32,f64} and their sinks.
const testLen = 1e5

var x = make([]complex64, testLen)

// guardVector copies the source vector (vec) into a new slice with guards.
// Guards guarded[:gdLn] and guarded[len-gdLn:] will be filled with sigil value gdVal.
func guardVector(vec []complex64, gdVal complex64, gdLn int) (guarded []complex64) {
	guarded = make([]complex64, len(vec)+gdLn*2)
	copy(guarded[gdLn:], vec)
	for i := 0; i < gdLn; i++ {
		guarded[i] = gdVal
		guarded[len(guarded)-1-i] = gdVal
	}
	return guarded
}

// isValidGuard will test for violated guards, generated by guardVector.
func isValidGuard(vec []complex64, gdVal complex64, gdLn int) bool {
	for i := 0; i < gdLn; i++ {
		if !sameCmplx(vec[i], gdVal) || !sameCmplx(vec[len(vec)-1-i], gdVal) {
			return false
		}
	}
	return true
}

// guardIncVector copies the source vector (vec) into a new incremented slice with guards.
// End guards will be length gdLen.
// Internal and end guards will be filled with sigil value gdVal.
func guardIncVector(vec []complex64, gdVal complex64, inc, gdLen int) (guarded []complex64) {
	if inc < 0 {
		inc = -inc
	}
	inrLen := len(vec) * inc
	guarded = make([]complex64, inrLen+gdLen*2)
	for i := range guarded {
		guarded[i] = gdVal
	}
	for i, v := range vec {
		guarded[gdLen+i*inc] = v
	}
	return guarded
}

// checkValidIncGuard will test for violated guards, generated by guardIncVector
func checkValidIncGuard(t *testing.T, vec []complex64, gdVal complex64, inc, gdLen int) {
	srcLn := len(vec) - 2*gdLen
	for i := range vec {
		switch {
		case sameCmplx(vec[i], gdVal):
			// Correct value
		case (i-gdLen)%inc == 0 && (i-gdLen)/inc < len(vec):
			// Ignore input values
		case i < gdLen:
			t.Errorf("Front guard violated at %d %v", i, vec[:gdLen])
		case i > gdLen+srcLn:
			t.Errorf("Back guard violated at %d %v", i-gdLen-srcLn, vec[gdLen+srcLn:])
		default:
			t.Errorf("Internal guard violated at %d %v", i-gdLen, vec[gdLen:gdLen+srcLn])
		}
	}
}

// same tests for nan-aware equality.
func same(a, b float32) bool {
	return a == b || (math32.IsNaN(a) && math32.IsNaN(b))
}

// sameApprox tests for nan-aware equality within tolerance.
func sameApprox(a, b, tol float32) bool {
	return same(a, b) || scalar.EqualWithinAbsOrRel(float64(a), float64(b), float64(tol), float64(tol))
}

// sameCmplx tests for nan-aware equality.
func sameCmplx(a, b complex64) bool {
	return a == b || (cmplx64.IsNaN(a) && cmplx64.IsNaN(b))
}

// sameCmplxApprox tests for nan-aware equality within tolerance.
func sameCmplxApprox(a, b complex64, tol float32) bool {
	return sameCmplx(a, b) || cmplxs.EqualWithinAbsOrRel(complex128(a), complex128(b), float64(tol), float64(tol))
}

var ( // Offset sets for testing alignment handling in Unitary assembly functions.
	align1 = []int{0, 1}
)

func randomSlice(n, inc int) []complex64 {
	if inc < 0 {
		inc = -inc
	}
	x := make([]complex64, (n-1)*inc+1)
	for i := range x {
		x[i] = complex(float32(rand.Float64()), float32(rand.Float64()))
	}
	return x
}
