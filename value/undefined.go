package value

import (
	"math"
)

type undefined struct{}

func Undefined() Value {
	return undefined{}
}

func (_ undefined) True() bool {
	return false
}

func (_ undefined) Add(other Value) (Value, error) {
	if s, ok := other.(Str); ok {
		s.value = "undefined" + s.value
		return s, nil
	}
	return CreateFloat(math.NaN()), nil
}

func (_ undefined) Sub(_ Value) (Value, error) {
	return CreateFloat(math.NaN()), nil
}

func (_ undefined) Mul(_ Value) (Value, error) {
	return CreateFloat(math.NaN()), nil
}

func (_ undefined) Div(_ Value) (Value, error) {
	return CreateFloat(math.NaN()), nil
}

func (_ undefined) Lshift(_ Value) (Value, error) {
	return CreateFloat(math.NaN()), nil
}

func (_ undefined) Rshift(_ Value) (Value, error) {
	return CreateFloat(math.NaN()), nil
}

func (_ undefined) Band(_ Value) (Value, error) {
	return CreateFloat(math.NaN()), nil
}

func (_ undefined) Bor(_ Value) (Value, error) {
	return CreateFloat(math.NaN()), nil
}

func (_ undefined) Compare(other Value) (int, error) {
	if _, ok := other.(undefined); ok {
		return 0, nil
	}
	return 0, ErrIncompatible
}

func (_ undefined) String() string {
	return "undefined"
}

func (u undefined) Type() string {
	return u.String()
}

type null struct{}

func Null() Value {
	return null{}
}

func (_ null) True() bool {
	return false
}

func (_ null) Add(other Value) (Value, error) {
	if s, ok := other.(Str); ok {
		s.value = "null" + s.value
		return s, nil
	}
	return other, nil
}

func (_ null) Sub(other Value) (Value, error) {
	switch other.(type) {
	case Float:
		return Reverse(other)
	case Bool:
		return CreateFloat(-1), nil
	case null:
		return CreateFloat(0), nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (_ null) Mul(other Value) (Value, error) {
	switch other.(type) {
	case Float, Bool, null:
		return CreateFloat(0), nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (_ null) Div(other Value) (Value, error) {
	switch other.(type) {
	case Float, Bool, null:
		return CreateFloat(0), nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (_ null) Compare(other Value) (int, error) {
	if _, ok := other.(null); ok {
		return 0, nil
	}
	return 0, ErrIncompatible
}

func (_ null) String() string {
	return "null"
}

func (n null) Type() string {
	return n.String()
}
