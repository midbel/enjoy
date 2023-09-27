package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/midbel/enjoy/eval"
)

func main() {
	var (
		trace = flag.Bool("t", false, "trace")
	)
	flag.Parse()
	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	now := time.Now()
	v, err := eval.EvalDefault(r)
	if err == nil && v != nil {
		fmt.Println(v)
	}
	if *trace {
		fmt.Printf("execution time: %s", time.Since(now))
		fmt.Println()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
