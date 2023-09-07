package builtins

import (
	"github.com/midbel/enjoy/value"
)

func Object() value.Value {
	obj := value.CreateGlobal("Object")
	obj.RegisterFunc("freeze", objectFreeze)
	obj.RegisterFunc("seal", objectSeal)
	obj.RegisterFunc("keys", objectKeys)
	return obj
}

func objectKeys(_ value.Global, args []value.Value) (value.Value, error) {
	obj, ok := args[0].(*value.Object)
	if !ok {
		return nil, value.ErrOperation
	}
	return obj.Keys(), nil
}

func objectFreeze(_ value.Global, args []value.Value) (value.Value, error) {
	obj, ok := args[0].(*value.Object)
	if !ok {
		return nil, value.ErrOperation
	}
	obj.Freeze()
	return obj, nil
}

func objectSeal(_ value.Global, args []value.Value) (value.Value, error) {
	obj, ok := args[0].(*value.Object)
	if !ok {
		return nil, value.ErrOperation
	}
	obj.Seal()
	return obj, nil
}
