package builtins

import (
	"fmt"

	"github.com/midbel/enjoy/value"
)

func ParseInt() value.Value {
	return value.CreateBuiltin("parseInt", parseInt)
}

func ParseFloat() value.Value {
	return value.CreateBuiltin("parseFloat", parseFloat)
}

func Print() value.Value {
	return value.CreateBuiltin("print", print)
}

func Fetch() value.Value {
	return nil
}

func parseInt(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, value.ErrArgument
	}
	return value.Coerce(args[0])
}

func parseFloat(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, value.ErrArgument
	}
	return value.Coerce(args[0])
}

func print(args ...value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, value.ErrArgument
	}
	fmt.Println(args[0])
	return nil, nil
}
