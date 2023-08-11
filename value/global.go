package value

import (
	"fmt"
)

type Global struct {
	name    string
	methods builtinMethodSet
	props   map[string]Value

	data interface{}
}

func CreateGlobal(name string) Global {
	return Global{
		name:    name,
		methods: make(builtinMethodSet),
		props:   make(map[string]Value),
	}
}

func (g Global) RegisterProp(ident string, val Value) {
	g.props[ident] = val
}

func (g Global) RegisterFunc(ident string, fn ValueFunc[Global]) {
	g.methods[ident] = fn
}

func (g Global) Get(prop string) (Value, error) {
	if v, ok := g.props[prop]; ok {
		return v, nil
	}
	if m, ok := g.methods[prop]; ok {
		return globalFunctoValue(prop, g, m), nil
	}
	return Undefined(), nil
}

func (g Global) Call(fn string, args []Value) (Value, error) {
	call, ok := g.methods[fn]
	if !ok {
		return nil, fmt.Errorf("%s not defined on %s", fn, g.name)
	}
	return call(g, args)
}

func (_ Global) True() bool {
	return true
}

func (g Global) String() string {
	return g.name
}

func (_ Global) Type() string {
	return "object"
}

type GlobalFunc ValueFunc[Global]

func globalFunctoValue(name string, g Global, fn ValueFunc[Global]) Value {
	call := func(args ...Value) (Value, error) {
		return fn(g, args)
	}
	return CreateBuiltin(name, call)
}

type builtinMethodSet map[string]ValueFunc[Global]

func ToNativeFloat(args []Value) ([]float64, error) {
	var list []float64
	for _, a := range args {
		n, err := Coerce(a)
		if err != nil {
			return nil, err
		}
		f, ok := n.(Float)
		if !ok {
			return nil, ErrOperation
		}
		list = append(list, f.value)
	}
	return list, nil
}
