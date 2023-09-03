package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/midbel/enjoy/ast"
	"github.com/midbel/enjoy/parser"
)

func main() {
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	defer r.Close()

	n, err := parser.Parse(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := ast.Debug(n, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
