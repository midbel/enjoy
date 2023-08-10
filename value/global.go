package value

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/midbel/enjoy/env"
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

func Default() env.Environ[Value] {
	top := env.EmptyEnv[Value]()
	top.Define("console", Console(), true)
	top.Define("Math", Math(), true)
	top.Define("Object", Super(), true)
	top.Define("JSON", Json(), true)

	top.Define("parseInt", CreateBuiltin("parseInt", builtinParseInt), true)
	top.Define("parseFloat", CreateBuiltin("parseFloat", builtinParseFloat), true)
	top.Define("print", CreateBuiltin("print", builtinPrint), true)

	return env.Immutable(top)
}

func Super() Value {
	obj := CreateGlobal("Object")
	obj.RegisterFunc("freeze", objectFreeze)
	obj.RegisterFunc("seal", objectSeal)
	obj.RegisterFunc("keys", objectKeys)
	return obj
}

func Console() Value {
	obj := CreateGlobal("console")
	obj.RegisterFunc("log", CheckArity(-1, consoleLog))
	obj.RegisterFunc("error", CheckArity(-1, consoleErr))
	return obj
}

func Json() Value {
	obj := CreateGlobal("JSON")
	obj.RegisterFunc("parse", CheckArity(1, jsonParse))
	obj.RegisterFunc("stringify", CheckArity(1, jsonString))
	return obj
}

func Math() Value {
	obj := CreateGlobal("Math")
	obj.RegisterProp("PI", CreateFloat(math.Pi))
	obj.RegisterProp("E", CreateFloat(math.E))

	one := func(fn func(float64) float64) ValueFunc[Global] {
		return func(_ Global, args []Value) (Value, error) {
			return doMath(args[0], fn)
		}
	}

	obj.RegisterFunc("sin", CheckArity(1, one(math.Sin)))
	obj.RegisterFunc("cos", CheckArity(1, one(math.Cos)))
	obj.RegisterFunc("tan", CheckArity(1, one(math.Tan)))
	obj.RegisterFunc("abs", CheckArity(1, one(math.Abs)))
	obj.RegisterFunc("ceil", CheckArity(1, one(math.Ceil)))
	obj.RegisterFunc("floor", CheckArity(1, one(math.Floor)))
	obj.RegisterFunc("round", CheckArity(1, one(math.Round)))
	obj.RegisterFunc("trunc", CheckArity(1, one(math.Trunc)))

	multi := func(fn func(float64, float64) float64) ValueFunc[Global] {
		return func(_ Global, args []Value) (Value, error) {
			return doMathN(args, fn)
		}
	}

	obj.RegisterFunc("min", CheckArity(-1, multi(math.Min)))
	obj.RegisterFunc("max", CheckArity(-1, multi(math.Max)))

	return obj
}

func consoleLog(_ Global, args []Value) (Value, error) {
	printValues(os.Stdout, args)
	return nil, nil
}

func consoleErr(_ Global, args []Value) (Value, error) {
	printValues(os.Stderr, args)
	return nil, nil
}

func printValues(w io.Writer, args []Value) {
	for i := range args {
		if i > 0 {
			fmt.Fprint(w, " ")
		}
		fmt.Fprint(w, args[i].String())
	}
	fmt.Fprintln(w)
}

func doMathN(vs []Value, do func(float64, float64) float64) (Value, error) {
	if len(vs) == 0 {
		return Undefined(), nil
	}
	list, err := toNativeFloat(vs)
	if err != nil {
		return nil, err
	}
	res := list[0]
	if len(list) == 1 {
		return CreateFloat(res), nil
	}
	for i := 1; i < len(list); i++ {
		res = do(res, list[i])
	}
	return CreateFloat(res), nil
}

func doMath(v Value, do func(float64) float64) (Value, error) {
	f, ok := v.(Float)
	if !ok {
		return nil, ErrOperation
	}
	f.value = do(f.value)
	return f, nil
}

func toNativeFloat(args []Value) ([]float64, error) {
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

func objectKeys(_ Global, args []Value) (Value, error) {
	obj, ok := args[0].(Object)
	if !ok {
		return nil, ErrOperation
	}
	var list []Value
	for k := range obj.values {
		list = append(list, CreateString(k))
	}
	slices.SortFunc(list, func(s1, s2 Value) int {
		return strings.Compare(s1.String(), s2.String())
	})
	return CreateArray(list), nil
}

func objectFreeze(_ Global, args []Value) (Value, error) {
	obj, ok := args[0].(Object)
	if !ok {
		return nil, ErrOperation
	}
	obj.frozen = true
	return obj, nil
}

func objectSeal(_ Global, args []Value) (Value, error) {
	obj, ok := args[0].(Object)
	if !ok {
		return nil, ErrOperation
	}
	obj.sealed = true
	return obj, nil
}

func jsonParse(_ Global, args []Value) (Value, error) {
	var (
		r = strings.NewReader(args[0].String())
		d interface{}
	)
	err := json.NewDecoder(r).Decode(&d)
	if err != nil {
		return nil, err
	}
	return jsonValues(d)
}

func jsonString(_ Global, args []Value) (Value, error) {
	return nil, ErrImplemented
}

func jsonValues(d interface{}) (Value, error) {
	switch v := d.(type) {
	case string:
		return CreateString(v), nil
	case float64:
		return CreateFloat(v), nil
	case bool:
		return CreateBool(v), nil
	case []interface{}:
		var list []Value
		for i := range v {
			d, err := jsonValues(v[i])
			if err != nil {
				return nil, err
			}
			list = append(list, d)
		}
		return CreateArray(list), nil
	case map[string]interface{}:
		list := make(map[string]Value)
		for k, v := range v {
			a, err := jsonValues(v)
			if err != nil {
				return nil, err
			}
			list[k] = a
		}
		return CreateObject(list), nil
	default:
		return nil, fmt.Errorf("%T unsupported json type")
	}
}
