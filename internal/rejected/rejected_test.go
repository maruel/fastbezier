// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package rejected

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

func TestPointsTrimmedError(t *testing.T) {
	testError(t, func(x0, y0, x1, y1 float32, steps uint16) Evaluator16 {
		return MakePointsTrimmed(x0, y0, x1, y1, steps)
	})
	// Make sure fitting points are exact.
	for _, curve := range curves {
		// Use default number of steps.
		p := MakePointsTrimmed(curve.x0, curve.y0, curve.x1, curve.y1, 0)
		for _, point := range p {
			y := p.Eval16(point.x)
			if y != point.y {
				t.Fatalf("At x=%d expected y=%d got %d", point.x, point.y, y)
			}
		}
	}
}

func TestPointsFullError(t *testing.T) {
	testError(t, func(x0, y0, x1, y1 float32, steps uint16) Evaluator16 {
		return MakePointsFull(x0, y0, x1, y1, steps)
	})
	// Make sure fitting points are exact.
	for _, curve := range curves {
		// Use default number of steps.
		p := MakePointsFull(curve.x0, curve.y0, curve.x1, curve.y1, 0)
		for _, point := range p {
			y := p.Eval16(point.x)
			if y != point.y {
				t.Fatalf("At x=%d expected y=%d got %d", point.x, point.y, y)
			}
		}
	}
}

func TestTableTrimmedError(t *testing.T) {
	/*
		testError(t, func(x0, y0, x1, y1 float32, steps uint16) Evaluator16 {
			return MakeTableTrimmed(x0, y0, x1, y1, steps)
		})
	*/
	// Make sure fitting points are exact.
	for _, curve := range curves {
		// Use default number of steps.
		e := MakeTableTrimmed(curve.x0, curve.y0, curve.x1, curve.y1, 0)
		for i, expectedY := range e {
			x := uint16((i + 1) * 65535 / (len(e) + 2))
			y := e.Eval16(x)
			if y != expectedY {
				t.Fatalf("At x=%d expected y=%d got %d", x, expectedY, y)
			}
		}
	}
}

func TestTableFullError(t *testing.T) {
	testError(t, func(x0, y0, x1, y1 float32, steps uint16) Evaluator16 {
		return MakeTableFull(x0, y0, x1, y1, steps)
	})
	// Make sure fitting points are exact.
	for _, curve := range curves {
		// Use default number of steps.
		e := MakeTableFull(curve.x0, curve.y0, curve.x1, curve.y1, 0)
		for i, expectedY := range e {
			if i == len(e)-1 {
				break
			}
			x := uint16(i * 65535 / (len(e) - 2))
			y := e.Eval16(x)
			if y != expectedY {
				t.Fatalf("At x=%d expected y=%d got %d", x, expectedY, y)
			}
		}
	}
}

func testError(t *testing.T, maker func(x0, y0, x1, y1 float32, steps uint16) Evaluator16) {
	for _, curve := range curves {
		var maxX uint16
		var maxDelta uint16
		// This results in 16 points.
		e := maker(curve.x0, curve.y0, curve.x1, curve.y1, 18)
		for i := 0; i < 65536; i++ {
			x := uint16(i)
			expected := internal.CubicBezier16(curve.x0, curve.y0, curve.x1, curve.y1, x)
			actual := e.Eval16(x)
			var delta uint16
			if expected < actual {
				delta = actual - expected
			} else {
				delta = expected - actual
			}
			if delta > maxDelta {
				maxDelta = delta
				maxX = x
			}
		}
		if maxDelta > 352 {
			t.Fatalf("curve=%4v: x=%d delta=%d\n", curve, maxX, maxDelta)
		}
		if e.Eval16(0) != 0 {
			t.Fatal("point 0 is not 0")
		}
		if e.Eval16(65535) != 65535 {
			t.Fatal("point 65535 is not 65535")
		}
	}
}

func ExampleMakePointsTrimmed() {
	p := MakePointsTrimmed(0, 0, 0.58, 1, 6)
	fmt.Printf("%s\n", p)
	// Each point is 32 bits.
	fmt.Printf("%d\n", len(p))
	fmt.Printf("%d\n", p.Eval16(1000))
	// Output:
	// PointsTrimmed{(0, 0), (4173, 6816), (15141, 23068), (30576, 42467), (48150, 58719), (65535, 65535)}
	// 4
	// 1633
}

func ExampleMakeTableTrimmed() {
	t := MakeTableTrimmed(0, 0, 0.58, 1, 6)
	fmt.Printf("%s\n", t)
	// Each point is 16 bits.
	fmt.Printf("%d\n", len(t))
	fmt.Printf("%d\n", t.Eval16(1000))
	// Output:
	// TableTrimmed{(0, 0), (13107, 17061), (26214, 32004), (39321, 44868), (52428, 55301), (65535, 65535)}
	// 4
	// 1562
}

func ExamplePointsTrimmed_Eval16() {
	curve := curves[3]
	const steps = 14
	b := MakePointsTrimmed(curve.x0, curve.y0, curve.x1, curve.y1, 0)
	fmt.Println("  i    xf    xi   yfi    yf    yi delta  %")
	for i := 0; i < steps; i++ {
		xf := float32(i) / float32(steps-1)
		yf := internal.CubicBezier(curve.x0, curve.y0, curve.x1, curve.y1, xf)
		xi := uint16(uint32(i) * 65535 / uint32(steps-1))
		yi := b.Eval16(xi)
		yfi := internal.FloatToUint16(yf * 65535.)
		delta := int(yfi) - int(yi)
		fmt.Printf("%3d %.3f %5d %.3f %5d %5d %5d %.3f%%\n", i, xf, xi, yf, yfi, yi, delta, float32(delta)*100./65535.)
	}
	// Output:
	//   i    xf    xi   yfi    yf    yi delta  %
	//   0 0.000     0 0.000     0     0     0 0.000%
	//   1 0.077  5041 0.125  8180  8177     3 0.005%
	//   2 0.154 10082 0.242 15830 15827     3 0.005%
	//   3 0.231 15123 0.352 23044 23033    11 0.017%
	//   4 0.308 20164 0.455 29835 29822    13 0.020%
	//   5 0.385 25205 0.552 36195 36178    17 0.026%
	//   6 0.462 30246 0.642 42098 42079    19 0.029%
	//   7 0.538 35288 0.725 47510 47492    18 0.027%
	//   8 0.615 40329 0.799 52383 52375     8 0.012%
	//   9 0.692 45370 0.864 56652 56636    16 0.024%
	//  10 0.769 50411 0.919 60236 60203    33 0.050%
	//  11 0.846 55452 0.962 63022 62987    35 0.053%
	//  12 0.923 60493 0.990 64860 64837    23 0.035%
	//  13 1.000 65535 1.000 65535 65535     0 0.000%
}

func ExampleTableTrimmed_Eval16() {
	curve := curves[3]
	const steps = 14
	b := MakeTableTrimmed(curve.x0, curve.y0, curve.x1, curve.y1, 0)
	fmt.Println("  i    xf    xi   yfi    yf    yi delta  %")
	for i := 0; i < steps; i++ {
		xf := float32(i) / float32(steps-1)
		yf := internal.CubicBezier(curve.x0, curve.y0, curve.x1, curve.y1, xf)
		xi := uint16(uint32(i) * 65535 / uint32(steps-1))
		yi := b.Eval16(xi)
		yfi := internal.FloatToUint16(yf * 65535.)
		delta := int(yfi) - int(yi)
		fmt.Printf("%3d %.3f %5d %.3f %5d %5d %5d %.3f%%\n", i, xf, xi, yf, yfi, yi, delta, float32(delta)*100./65535.)
	}
	// Output:
	//   i    xf    xi   yfi    yf    yi delta  %
	//   0 0.000     0 0.000     0     0     0 0.000%
	//   1 0.077  5041 0.125  8180  8171     9 0.014%
	//   2 0.154 10082 0.242 15830 15827     3 0.005%
	//   3 0.231 15123 0.352 23044 23035     9 0.014%
	//   4 0.308 20164 0.455 29835 29831     4 0.006%
	//   5 0.385 25205 0.552 36195 36186     9 0.014%
	//   6 0.462 30246 0.642 42098 42090     8 0.012%
	//   7 0.538 35288 0.725 47510 47502     8 0.012%
	//   8 0.615 40329 0.799 52383 52372    11 0.017%
	//   9 0.692 45370 0.864 56652 56645     7 0.011%
	//  10 0.769 50411 0.919 60236 60220    16 0.024%
	//  11 0.846 55452 0.962 63022 63015     7 0.011%
	//  12 0.923 60493 0.990 64860 64835    25 0.038%
	//  13 1.000 65535 1.000 65535 65535     0 0.000%
}

var dummyE Evaluator16
var dummyI uint16

func BenchmarkMakePointsTrimmed_8(b *testing.B) {
	var p PointsTrimmed
	for n := 0; n < b.N; n++ {
		p = MakePointsTrimmed(0.42, 0, 0.58, 1, 8)
	}
	dummyE = p
}

func BenchmarkMakePointsTrimmed_32(b *testing.B) {
	var p PointsTrimmed
	for n := 0; n < b.N; n++ {
		p = MakePointsTrimmed(0.42, 0, 0.58, 1, 32)
	}
	dummyE = p
}

func BenchmarkMakePointsTrimmed_130(b *testing.B) {
	var p PointsTrimmed
	for n := 0; n < b.N; n++ {
		p = MakePointsTrimmed(0.42, 0, 0.58, 1, 130)
	}
	dummyE = p
}

func BenchmarkMakePointsFull_8(b *testing.B) {
	var p PointsFull
	for n := 0; n < b.N; n++ {
		p = MakePointsFull(0.42, 0, 0.58, 1, 8)
	}
	dummyE = p
}

func BenchmarkMakePointsFull_32(b *testing.B) {
	var p PointsFull
	for n := 0; n < b.N; n++ {
		p = MakePointsFull(0.42, 0, 0.58, 1, 32)
	}
	dummyE = p
}

func BenchmarkMakePointsFull_130(b *testing.B) {
	var p PointsFull
	for n := 0; n < b.N; n++ {
		p = MakePointsFull(0.42, 0, 0.58, 1, 130)
	}
	dummyE = p
}

func BenchmarkMakeTableTrimmed_8(b *testing.B) {
	var t TableTrimmed
	for n := 0; n < b.N; n++ {
		t = MakeTableTrimmed(0.42, 0, 0.58, 1, 8)
	}
	dummyE = t
}

func BenchmarkMakeTableTrimmed_32(b *testing.B) {
	var t TableTrimmed
	for n := 0; n < b.N; n++ {
		t = MakeTableTrimmed(0.42, 0, 0.58, 1, 32)
	}
	dummyE = t
}

func BenchmarkMakeTableTrimmed_130(b *testing.B) {
	var t TableTrimmed
	for n := 0; n < b.N; n++ {
		t = MakeTableTrimmed(0.42, 0, 0.58, 1, 130)
	}
	dummyE = t
}

func BenchmarkMakeTableFull_8(b *testing.B) {
	var t TableFull
	for n := 0; n < b.N; n++ {
		t = MakeTableFull(0.42, 0, 0.58, 1, 8)
	}
	dummyE = t
}

func BenchmarkMakeTableFull_32(b *testing.B) {
	var t TableFull
	for n := 0; n < b.N; n++ {
		t = MakeTableFull(0.42, 0, 0.58, 1, 32)
	}
	dummyE = t
}

func BenchmarkMakeTableFull_130(b *testing.B) {
	var t TableFull
	for n := 0; n < b.N; n++ {
		t = MakeTableFull(0.42, 0, 0.58, 1, 130)
	}
	dummyE = t
}

//

func BenchmarkPointsTrimmed_Eval16_1000(b *testing.B) {
	p := MakePointsTrimmed(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = p.Eval16(1000)
	}
	dummyI = r
}

func BenchmarkPointsTrimmed_Eval16_32767(b *testing.B) {
	p := MakePointsTrimmed(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = p.Eval16(32767)
	}
	dummyI = r
}

func BenchmarkPointsTrimmed_Eval16_65435(b *testing.B) {
	p := MakePointsTrimmed(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = p.Eval16(65435)
	}
	dummyI = r
}

func BenchmarkPointsFull_Eval16_1000(b *testing.B) {
	p := MakePointsFull(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = p.Eval16(1000)
	}
	dummyI = r
}

func BenchmarkPointsFull_Eval16_32767(b *testing.B) {
	p := MakePointsFull(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = p.Eval16(32767)
	}
	dummyI = r
}

func BenchmarkPointsFull_Eval16_65435(b *testing.B) {
	p := MakePointsFull(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = p.Eval16(65435)
	}
	dummyI = r
}

func BenchmarkTableTrimmed_Eval16_1000(b *testing.B) {
	t := MakeTableTrimmed(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = t.Eval16(1000)
	}
	dummyI = r
}

func BenchmarkTableTrimmed_Eval16_32767(b *testing.B) {
	t := MakeTableTrimmed(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = t.Eval16(32767)
	}
	dummyI = r
}

func BenchmarkTableTrimmed_Eval16_65435(b *testing.B) {
	t := MakeTableTrimmed(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = t.Eval16(65435)
	}
	dummyI = r
}

func BenchmarkTableFull_Eval16_1000(b *testing.B) {
	t := MakeTableFull(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = t.Eval16(1000)
	}
	dummyI = r
}

func BenchmarkTableFull_Eval16_32767(b *testing.B) {
	t := MakeTableFull(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = t.Eval16(32767)
	}
	dummyI = r
}

func BenchmarkTableFull_Eval16_65435(b *testing.B) {
	t := MakeTableFull(0.42, 0, 0.58, 1, 0)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = t.Eval16(65435)
	}
	dummyI = r
}

func BenchmarkPrecise(b *testing.B) {
	p := MakePrecise(0.42, 0, 0.58, 1)
	r := uint16(0)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r = p.Eval16(1000)
	}
	dummyI = r
}
