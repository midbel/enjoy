package builtins

import (
	"math"

	"github.com/midbel/enjoy/value"
)

func Math() value.Value {
	obj := value.CreateGlobal("Math")
	obj.RegisterProp("PI", value.CreateFloat(math.Pi))
	obj.RegisterProp("E", value.CreateFloat(math.E))

	one := func(fn func(float64) float64) value.ValueFunc[value.Global] {
		return func(_ value.Global, args []value.Value) (value.Value, error) {
			return doMath(args[0], fn)
		}
	}

	obj.RegisterFunc("sin", value.CheckArity(1, one(math.Sin)))
	obj.RegisterFunc("cos", value.CheckArity(1, one(math.Cos)))
	obj.RegisterFunc("tan", value.CheckArity(1, one(math.Tan)))
	obj.RegisterFunc("abs", value.CheckArity(1, one(math.Abs)))
	obj.RegisterFunc("ceil", value.CheckArity(1, one(math.Ceil)))
	obj.RegisterFunc("floor", value.CheckArity(1, one(math.Floor)))
	obj.RegisterFunc("round", value.CheckArity(1, one(math.Round)))
	obj.RegisterFunc("trunc", value.CheckArity(1, one(math.Trunc)))

	multi := func(fn func(float64, float64) float64) value.ValueFunc[value.Global] {
		return func(_ value.Global, args []value.Value) (value.Value, error) {
			return doMathN(args, fn)
		}
	}

	obj.RegisterFunc("min", value.CheckArity(-1, multi(math.Min)))
	obj.RegisterFunc("max", value.CheckArity(-1, multi(math.Max)))

	return obj
}

func doMathN(vs []value.Value, do func(float64, float64) float64) (value.Value, error) {
	if len(vs) == 0 {
		return value.Undefined(), nil
	}
	list, err := value.ToNativeFloat(vs)
	if err != nil {
		return nil, err
	}
	res := list[0]
	if len(list) == 1 {
		return value.CreateFloat(res), nil
	}
	for i := 1; i < len(list); i++ {
		res = do(res, list[i])
	}
	return value.CreateFloat(res), nil
}

func doMath(v value.Value, do func(float64) float64) (value.Value, error) {
	f, ok := v.(value.Float)
	if !ok {
		return nil, value.ErrOperation
	}
	res := do(f.Native())
	return value.CreateFloat(res), nil
}
