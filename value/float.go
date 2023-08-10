package value

import (
	"fmt"
	"math"
	"strconv"
)

type Float struct {
	value float64
}

func CreateFloat(f float64) Value {
	return Float{
		value: f,
	}
}

func (f Float) True() bool {
	return f.value != 0
}

func (f Float) Rev() Value {
	f.value = -f.value
	return f
}

func (f Float) Incr() Value {
	i := int64(f.value)
	i++
	f.value = float64(i)
	return f
}

func (f Float) Decr() Value {
	i := int64(f.value)
	i--
	f.value = float64(i)
	return f
}

func (f Float) Add(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		f.value += x.value
		return f, nil
	case Str:
		x.value = fmt.Sprintf("%f%s", f.value, x.value)
		return x, nil
	case Bool:
		if x.value {
			f.value += 1
		}
		return f, nil
	case null:
		return f, nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (f Float) Sub(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		f.value -= x.value
		return f, nil
	case Bool:
		if x.value {
			f.value -= 1
		}
		return f, nil
	case null:
		return f, nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (f Float) Mul(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		f.value *= x.value
		return f, nil
	case Bool:
		if !x.value {
			return CreateFloat(0), nil
		}
		return f, nil
	case null:
		return CreateFloat(0), nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (f Float) Pow(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		f.value = math.Pow(f.value, x.value)
		return f, nil
	case Bool:
		if !x.value {
			return CreateFloat(1), nil
		}
		return f, nil
	case null:
		return CreateFloat(0), nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (f Float) Div(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value /= x.value
		return f, nil
	case Bool:
		if !x.value {
			return nil, ErrZero
		}
		return f, nil
	case null:
		return nil, ErrZero
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (f Float) Mod(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value = math.Mod(f.value, x.value)
		return f, nil
	case Bool:
		if !x.value {
			return CreateFloat(math.NaN()), nil
		}
		return CreateFloat(0), nil
	case undefined, null:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (f Float) Compare(other Value) (int, error) {
	x, ok := other.(Float)
	if !ok {
		return 0, ErrIncompatible
	}
	var res int
	if f.value == x.value {
		// pass
	} else if f.value < x.value {
		res--
	} else {
		res++
	}
	return res, nil
}

func (f Float) Call(fn string, args []string) (Value, error) {
	return nil, nil
}

func (f Float) String() string {
	return strconv.FormatFloat(f.value, 'f', -1, 64)
}

func (_ Float) Type() string {
	return "float"
}
