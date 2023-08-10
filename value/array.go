package value

import (
	"fmt"
	"strings"

	"github.com/midbel/enjoy/env"
)

type Array struct {
	values []Value
}

func CreateArray(vs []Value) Value {
	return Array{
		values: vs,
	}
}

func (a Array) True() bool {
	return a.Len() > 0
}

func (a Array) Len() int {
	return len(a.values)
}

func (a Array) At(ix Value) (Value, error) {
	x, ok := ix.(Float)
	if !ok {
		return nil, ErrOperation
	}
	i := int(x.value)
	if i < 0 || i >= a.Len() {
		return Undefined(), ErrIndex
	}
	return a.values[i], nil
}

func (a Array) Get(prop string) (Value, error) {
	if prop != "length" {
		return Undefined(), nil
	}
	return CreateFloat(float64(len(a.values))), nil
}

func (a Array) Call(fn string, args []Value) (Value, error) {
	call, ok := arrayPrototype[fn]
	if !ok {
		return nil, fmt.Errorf("%s not defined on array", fn)
	}
	return call(a, args)
}

func (a Array) String() string {
	var str strings.Builder
	str.WriteRune('[')
	for i, v := range a.values {
		if i > 0 {
			str.WriteRune(',')
			str.WriteRune(' ')
		}
		str.WriteString(v.String())
	}
	str.WriteRune(']')
	return str.String()
}

func (b Array) Type() string {
	return "array"
}

var arrayPrototype = map[string]ValueFunc[Array]{
	"map":     arrayMap,
	"forEach": arrayForEach,
}

func arrayMap(a Array, args []Value) (Value, error) {
	fn, ok := args[0].(Func)
	if !ok {
		return nil, ErrOperation
	}
	var (
		list  []Value
		ident string
		index string
	)
	if len(fn.Params) >= 1 {
		ident = fn.Params[0].Name
	}
	for i := range a.values {
		tmp := env.EnclosedEnv[Value](fn.Env)
		if ident != "" {
			tmp.Define(ident, a.values[i], false)
		}
		if index != "" {
			tmp.Define(index, CreateFloat(float64(i)), false)
		}
		v, err := fn.Body.Eval(tmp)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return CreateArray(list), nil
}

func arrayForEach(a Array, args []Value) (Value, error) {
	fn, ok := args[0].(Func)
	if !ok {
		return nil, ErrOperation
	}
	var (
		ident string
		index string
	)
	if len(fn.Params) >= 1 {
		ident = fn.Params[0].Name
	}
	if len(fn.Params) >= 2 {
		index = fn.Params[1].Name
	}
	for i := range a.values {
		tmp := env.EnclosedEnv[Value](fn.Env)
		if ident != "" {
			tmp.Define(ident, a.values[i], false)
		}
		if index != "" {
			tmp.Define(index, CreateFloat(float64(i)), false)
		}
		_, err := fn.Body.Eval(tmp)
		if err != nil {
			return nil, err
		}
	}
	return null{}, nil
}
