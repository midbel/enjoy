package value

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/midbel/enjoy/env"
)

var (
	ErrIncompatible = errors.New("incompatible type")
	ErrOperation    = errors.New("unsupported operation")
	ErrZero         = errors.New("division by zero")
	ErrIndex        = errors.New("index out of range")
	ErrArgument     = errors.New("wrong number of arguments given")
	ErrImplemented  = errors.New("not yet implemented")
)

type Value interface {
	True() bool
	Type() string
	fmt.Stringer
}

type Evaluable interface {
	Eval(env.Environ[Value]) (Value, error)
}

type Spreadable interface {
	Spread() []Value
}

type Enumerable interface {
	Enumerate() []Value
}

func IsNull(v Value) bool {
	_, ok := v.(null)
	return ok || v == nil
}

func IsUndefined(v Value) bool {
	_, ok := v.(undefined)
	return ok || v == nil
}

type Comparable interface {
	Compare(Value) (int, error)
}

func Compare(fst, snd Value, do func(int) bool) (Value, error) {
	res, err := compare(fst, snd)
	if err != nil {
		if errors.Is(err, ErrIncompatible) {
			return CreateBool(false), nil
		}
		return nil, err
	}
	return CreateBool(do(res)), nil
}

func compare(fst, snd Value) (int, error) {
	cmp, ok := fst.(Comparable)
	if !ok {
		return -1, ErrOperation
	}
	return cmp.Compare(snd)
}

type Callable interface {
	Call(string, []Value) (Value, error)
}

type Indexable interface {
	At(Value) (Value, error)
}

func At(v, ix Value) (Value, error) {
	at, ok := v.(Indexable)
	if !ok {
		return nil, ErrOperation
	}
	return at.At(ix)
}

type Getter interface {
	Get(string) (Value, error)
}

func Get(v Value, prop string) (Value, error) {
	g, ok := v.(Getter)
	if !ok {
		return nil, ErrOperation
	}
	return g.Get(prop)
}

type Setter interface {
	Set(string, Value) error
}

func Set(v Value, prop string, val Value) error {
	s, ok := v.(Setter)
	if !ok {
		return ErrOperation
	}
	return s.Set(prop, val)
}

type ValueFunc[T any] func(T, []Value) (Value, error)

func CheckArity[T any](max int, fn ValueFunc[T]) ValueFunc[T] {
	return func(v T, args []Value) (Value, error) {
		if max >= 0 && len(args) < max {
			return nil, ErrArgument
		}
		return fn(v, args)
	}
}

func Coerce(v Value) (Value, error) {
	switch v := v.(type) {
	case Float:
		return v, nil
	case Str:
		n, err := strconv.ParseFloat(v.value, 64)
		if err != nil {
			return nil, err
		}
		return CreateFloat(n), nil
	case Bool:
		var n float64
		if v.value {
			n = 1
		}
		return CreateFloat(n), nil
	case null:
		return CreateFloat(0), nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrOperation
	}
}

func Reverse(v Value) (Value, error) {
	f, ok := v.(Float)
	if !ok {
		return nil, ErrOperation
	}
	return f.Rev(), nil
}

func Increment(v Value) (Value, error) {
	f, ok := v.(Float)
	if !ok {
		return nil, ErrOperation
	}
	return f.Incr(), nil
}

func Decrement(v Value) (Value, error) {
	f, ok := v.(Float)
	if !ok {
		return nil, ErrOperation
	}
	return f.Decr(), nil
}

type Spread struct {
	Value
}

func SpreadValue(v Value) (Value, error) {
	if _, ok := v.(Spreadable); !ok {
		return nil, ErrOperation
	}
	s := Spread{
		Value: v,
	}
	return s, nil
}

func (s Spread) Spread() []Value {
	switch s := s.Value.(type) {
	case *Array:
		return s.Spread()
	case Str:
		return s.Spread()
	default:
		return nil
	}
}

func toNativeInt(v Value) (int, error) {
	v, err := Coerce(v)
	if err != nil {
		return 0, err
	}
	f, ok := v.(Float)
	if !ok {
		return 0, nil
	}
	return int(f.value), nil
}
