// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package fastbezier

import (
	"bytes"
	"fmt"
	"io"

	"github.com/maruel/fastbezier/internal"
)

// LUT is a fast cubic bezier curve evaluator over uint16 that uses a lookup
// table.
//
// Values are constrained in the range [0, 65535] for both x and y. It forces
// points [0, 0] and [65535, 65535].
type LUT []uint16

// Make returns a LUT object.
//
// Memory allocation is 2*(steps+1) bytes.
func Make(x0, y0, x1, y1 float32, steps uint16) LUT {
	if steps < 3 {
		// Make invalid `steps` value silently work instead of crashing or inducing
		// unnecessary error handling.
		steps = 32
	}
	stepsm1 := 1. / float32(steps-1)
	l := make(LUT, steps, steps+1)
	for i := range l {
		l[i] = internal.FloatToUint16(internal.CubicBezier(x0, y0, x1, y1, float32(i)*stepsm1) * 65535.)
	}
	// Adds a second 65535 to speed up Eval(); otherwise x==65535 has to be
	// special cased which slows it down.
	l = append(l, 65535)
	return l
}

func (l LUT) String() string {
	b := bytes.NewBufferString("LUT{")
	steps := len(l) - 2
	for i, y := range l {
		x := i * 65535 / steps
		fmt.Fprintf(b, "(%d, %d)", x, y)
		if i == steps {
			break
		}
		io.WriteString(b, ", ")
	}
	io.WriteString(b, "}")
	return b.String()
}

func (l LUT) Eval(x uint16) uint16 {
	steps := uint32(len(l) - 2)
	x32 := uint32(x)
	index := x32 * steps / 65535
	nextX := (index + 1) * 65535 / steps
	baseX := index * 65535 / steps
	a := uint32(l[index]) * (nextX - x32)
	b := uint32(l[index+1]) * (x32 - baseX)
	return uint16((a + b) / (nextX - baseX))
}
