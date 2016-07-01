// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package fastbezier

import (
	"fmt"
	"testing"

	"github.com/maruel/fastbezier/internal"
)

var curves = []struct {
	x0, y0, x1, y1 float32
}{
	{0.25, 0.1, 0.25, 1}, // TransitionEase
	{0.42, 0, 1, 1},      // TransitionEaseIn
	{0.42, 0, 0.58, 1},   // TransitionEaseInOut
	{0, 0, 0.58, 1},      // TransitionEaseOut
}

func TestLUT(t *testing.T) {
	for _, curve := range curves {
		var maxX uint16
		var maxDelta uint16
		l := MakeLUT(curve.x0, curve.y0, curve.x1, curve.y1, 16)
		for i := 0; i < 65536; i++ {
			expected := internal.CubicBezier16(curve.x0, curve.y0, curve.x1, curve.y1, uint16(i))
			actual := l.Eval16(uint16(i))
			var delta uint16
			if expected < actual {
				delta = actual - expected
			} else {
				delta = expected - actual
			}
			if delta > maxDelta {
				maxDelta = delta
				maxX = uint16(i)
			}
		}
		if maxDelta > 427 {
			t.Fatalf("curve=%4v: x=%d delta=%d\n", curve, maxX, maxDelta)
		}
		if l.Eval16(0) != 0 {
			t.Fatal("point 0 is not 0")
		}
		if l.Eval16(65535) != 65535 {
			t.Fatal("point 65535 is not 65535")
		}

		// Make sure fitting points are exact.
		// Use default number of steps.
		l = MakeLUT(curve.x0, curve.y0, curve.x1, curve.y1, 0)
		for i, expectedY := range l {
			if i == len(l)-1 {
				break
			}
			x := uint16(i * 65535 / (len(l) - 2))
			y := l.Eval16(x)
			t.Logf("x=%d expected y=%d y=%d", x, expectedY, y)
			if y != expectedY {
				t.Fatalf("At x=%d expected y=%d got %d\n%s", x, expectedY, y, l)
			} else {
				t.Logf("x=%d expected y=%d y=%d", x, expectedY, y)
			}
		}
	}
}

func ExampleMakeLUT() {
	l := MakeLUT(0, 0, 0.58, 1, 6)
	fmt.Printf("%s\n", l)
	// Each point is 16 bits.
	fmt.Printf("%d\n", len(l))
	fmt.Printf("%d\n", l.Eval16(1000))
	// Output:
	// LUT{(0, 0), (13107, 20209), (26214, 37413), (39321, 51454), (52428, 61453), (65535, 65535)}
	// 7
	// 1541
}

func ExampleLUT_Eval16() {
	const steps = 14
	l := MakeLUT(0.42, 0, 0.58, 1, 0)
	fmt.Println("  i    xf    xi   yfi    yf    yi delta   error")
	for i := 0; i < steps; i++ {
		xf := float32(i) / float32(steps-1)
		yf := internal.CubicBezier(0.42, 0, 0.58, 1, xf)
		xi := uint16(uint32(i) * 65535 / uint32(steps-1))
		yi := l.Eval16(xi)
		yfi := internal.FloatToUint16(yf * 65535.)
		delta := int(yfi) - int(yi)
		fmt.Printf("%3d %.3f %5d %.3f %5d %5d %5d %6.3f%%\n", i, xf, xi, yf, yfi, yi, delta, float32(delta)*100./65535.)
	}
	// Output:
	//   i    xf    xi   yfi    yf    yi delta   error
	//   0 0.000     0 0.000     0     0     0  0.000%
	//   1 0.077  5041 0.012   758   791   -33 -0.050%
	//   2 0.154 10082 0.048  3121  3147   -26 -0.040%
	//   3 0.231 15123 0.110  7182  7200   -18 -0.027%
	//   4 0.308 20164 0.197 12927 12960   -33 -0.050%
	//   5 0.385 25205 0.308 20159 20165    -6 -0.009%
	//   6 0.462 30246 0.434 28438 28443    -5 -0.008%
	//   7 0.538 35288 0.566 37097 37091     6  0.009%
	//   8 0.615 40329 0.692 45376 45369     7  0.011%
	//   9 0.692 45370 0.803 52608 52574    34  0.052%
	//  10 0.769 50411 0.890 58353 58334    19  0.029%
	//  11 0.846 55452 0.952 62414 62387    27  0.041%
	//  12 0.923 60493 0.988 64777 64743    34  0.052%
	//  13 1.000 65535 1.000 65535 65535     0  0.000%
}

var dummyL LUT
var dummyI uint16

func BenchmarkMakeLUT_8(b *testing.B) {
	var l LUT
	for n := 0; n < b.N; n++ {
		l = MakeLUT(0.42, 0, 0.58, 1, 8)
	}
	dummyL = l
}

func BenchmarkMakeLUT_32(b *testing.B) {
	var l LUT
	for n := 0; n < b.N; n++ {
		l = MakeLUT(0.42, 0, 0.58, 1, 32)
	}
	dummyL = l
}

func BenchmarkMakeLUT_130(b *testing.B) {
	var l LUT
	for n := 0; n < b.N; n++ {
		l = MakeLUT(0.42, 0, 0.58, 1, 128)
	}
	dummyL = l
}

func BenchmarkLUT_Eval16_1000(b *testing.B) {
	l := MakeLUT(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = l.Eval16(1000)
	}
	dummyI = r
}

func BenchmarkLUT_Eval16_32767(b *testing.B) {
	l := MakeLUT(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = l.Eval16(32767)
	}
	dummyI = r
}

func BenchmarkLUT_Eval16_65435(b *testing.B) {
	l := MakeLUT(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = l.Eval16(65435)
	}
	dummyI = r
}
