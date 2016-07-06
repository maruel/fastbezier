// Copyright 2016 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/maruel/fastbezier"
	"github.com/maruel/fastbezier/internal/rejected"
)

func compare() {
	evaluators := []rejected.Evaluator{
		rejected.MakePrecise(0.42, 0, 0.58, 1),
		fastbezier.Make(0.42, 0, 0.58, 1, 0),
		rejected.MakePointsTrimmed(0.42, 0, 0.58, 1, 0),
		rejected.MakePointsFull(0.42, 0, 0.58, 1, 0),
		rejected.MakeTableTrimmed(0.42, 0, 0.58, 1, 0),
		rejected.MakeTableFull(0.42, 0, 0.58, 1, 0),
	}
	for _, e := range evaluators {
		fmt.Printf("%s\n", e)
	}
	fmt.Printf("     x   Slow   LUT  Pnts PtsFl Table TblFl     LUT  Pnts PtsFl Table TblFl        LUT   Points  PtsFull    Table  TablFll\n")
	r := make([]uint16, len(evaluators))
	delta := make([]int, len(evaluators)-1)
	relDelta := make([]string, len(evaluators)-1)
	const steps = 50
	for i := 0; i <= steps; i++ {
		x := i * 65535 / steps
		for j, e := range evaluators {
			r[j] = e.Eval(uint16(x))
			if j != 0 {
				delta[j-1] = int(r[0]) - int(r[j])
				if delta[j-1] == 0 {
					relDelta[j-1] = "0"
				} else {
					relDelta[j-1] = fmt.Sprintf("%2.2f%%", 100.*float32(delta[j-1])/65535.)
				}
			}
		}
		fmt.Printf("%6d %5v %5v %8v\n", x, r, delta, relDelta)
	}
}

func mainImpl() error {
	compareF := flag.Bool("compare", false, "compare evaluators")
	flag.Parse()

	if *compareF {
		if flag.NArg() != 0 {
			return errors.New("do not supply values with -compare")
		}
		compare()
		return nil
	}
	if flag.NArg() != 5 {
		return errors.New("supply 5 values")
	}
	x0, err := strconv.ParseFloat(flag.Arg(0), 64)
	if err != nil {
		return err
	}
	y0, err := strconv.ParseFloat(flag.Arg(1), 64)
	if err != nil {
		return err
	}
	x1, err := strconv.ParseFloat(flag.Arg(2), 64)
	if err != nil {
		return err
	}
	y1, err := strconv.ParseFloat(flag.Arg(3), 64)
	if err != nil {
		return err
	}
	steps, err := strconv.Atoi(flag.Arg(4))
	if err != nil {
		return err
	}
	f := fastbezier.Make(float32(x0), float32(y0), float32(x1), float32(y1), uint16(steps))
	_, err = fmt.Printf("%s\n", f)
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "usage: makebezier <x0> <y0> <x1> <y1> <steps>\nmakebezier: %s.\n", err)
		os.Exit(1)
	}
}
