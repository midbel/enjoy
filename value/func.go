package value

import (
	"fmt"
	"strings"

	"github.com/midbel/enjoy/ast"
	"github.com/midbel/enjoy/env"
)

type Func struct {
	Ident  string
	Params []Parameter
	Body   Evaluable
	Env    env.Environ[Value]
}

func (_ Func) True() bool {
	return true
}

func (f Func) String() string {
	name := f.Ident
	if name == "" {
		name = "anonymous"
	}
	var params []string
	for _, p := range f.Params {
		params = append(params, p.Name)
	}
	return fmt.Sprintf("f %s (%s)", name, strings.Join(params, ", "))
}

func (_ Func) Type() string {
	return "function"
}

type Parameter struct {
	Name  string
	Value ast.Node
}

type BuiltinFunc func(...Value) (Value, error)

type Builtin struct {
	name string
	call BuiltinFunc
}

func CreateBuiltin(name string, fn BuiltinFunc) Builtin {
	return Builtin{
		name: name,
		call: fn,
	}
}

func (_ Builtin) True() bool {
	return true
}

func (b Builtin) String() string {
	return fmt.Sprintf("f %s() { [native code] }", b.name)
}

func (b Builtin) Type() string {
	return "builtin"
}

func (b Builtin) Apply(args []Value) (Value, error) {
	v, err := b.call(args...)
	if err != nil {
		err = fmt.Errorf("%s: %w", b.name, err)
	}
	return v, err
}

func builtinParseInt(args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, ErrArgument
	}
	return Coerce(args[0])
}

func builtinParseFloat(args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, ErrArgument
	}
	return Coerce(args[0])
}

func builtinPrint(args ...Value) (Value, error) {
	if len(args) != 1 {
		return nil, ErrArgument
	}
	fmt.Println(args[0])
	return nil, nil
}
