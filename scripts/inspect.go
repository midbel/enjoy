package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/midbel/enjoy/ast"
	"github.com/midbel/enjoy/eval"
	"github.com/midbel/enjoy/parser"
	"github.com/midbel/enjoy/scanner"
	"github.com/midbel/enjoy/token"
)

func main() {
	var (
		scanning = flag.Bool("s", false, "scan")
		parsing  = flag.Bool("p", false, "parse")
		trace    = flag.Bool("t", false, "trace")
	)
	flag.Parse()
	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	switch {
	case *scanning:
		err = scanFile(r)
	case *parsing:
		err = parseFile(r)
	default:
		err = evalFile(r, *trace)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func evalFile(r io.Reader, trace bool) error {
	now := time.Now()
	v, err := eval.EvalDefault(r)
	if err == nil && v != nil {
		fmt.Println(v)
	}
	if trace {
		fmt.Printf("execution time: %s", time.Since(now))
		fmt.Println()
	}
	return err
}

func parseFile(r io.Reader) error {
	node, err := parser.NewParser(r).Parse()
	if err != nil {
		return err
	}
	b, ok := node.(ast.BlockNode)
	if !ok {
		fmt.Printf("%#v", node)
		fmt.Println()
		return nil
	}
	for _, n := range b.Nodes {
		fmt.Printf("%#v", n)
		fmt.Println()
	}
	return nil
}

func scanFile(r io.Reader) error {
	scan := scanner.Scan(r)
	for {
		tok := scan.Scan()
		if tok.Type == token.EOF {
			break
		}
		fmt.Printf("%d,%d: %s", tok.Line, tok.Column, tok)
		fmt.Println()
	}
	return nil
}
