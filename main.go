// Copyright (c) 2023 by Gilbert Ramirez <gram@alumni.rice.edu>

package main

import (
	"fmt"
	"os"

	"github.com/gilramir/argparse/v2"
)

type Options struct {
	Filename string
}

type Program struct {
	opts Options
}

func (s *Program) run() error {
	ap := s.build_args()
	ap.Parse()

	return s.convert()
}

func (s *Program) build_args() *argparse.ArgumentParser {
	ap := argparse.New(&argparse.Command{
		Description: "Create quizlet CSV files",
		Values:      &s.opts,
	})

	/*
		ap.Add(&argparse.Argument{
			Name: "filename",
			Help: "The input filename",
		})
	*/

	return ap
}

func main() {
	var prog Program
	err := prog.run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
