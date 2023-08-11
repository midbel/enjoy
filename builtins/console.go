package builtins

import (
	"fmt"
	"io"
	"os"

	"github.com/midbel/enjoy/value"
)

func Console() value.Value {
	obj := value.CreateGlobal("console")
	obj.RegisterFunc("log", value.CheckArity(-1, consoleLog))
	obj.RegisterFunc("error", value.CheckArity(-1, consoleErr))
	return obj
}

func consoleLog(_ value.Global, args []value.Value) (value.Value, error) {
	printValues(os.Stdout, args)
	return nil, nil
}

func consoleErr(_ value.Global, args []value.Value) (value.Value, error) {
	printValues(os.Stderr, args)
	return nil, nil
}

func printValues(w io.Writer, args []value.Value) {
	for i := range args {
		if i > 0 {
			fmt.Fprint(w, " ")
		}
		fmt.Fprint(w, args[i].String())
	}
	fmt.Fprintln(w)
}
