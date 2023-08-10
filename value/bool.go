package value

import (
	"fmt"
	"math"
	"strconv"
)

type Bool struct {
	value bool
}

func CreateBool(b bool) Value {
	return Bool{
		value: b,
	}
}

func (b Bool) True() bool {
	return b.value
}

func (b Bool) Add(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		if b.value {
			x.value += 1
		}
		return x, nil
	case Str:
		x.value = fmt.Sprintf("%t%s", b.value, x.value)
		return x, nil
	case Bool:
		var n float64
		if b.value {
			n += 1
		}
		if x.value {
			n += 1
		}
		return CreateFloat(n), nil
	case null:
		var n float64
		if b.value {
			n += 1
		}
		return CreateFloat(n), nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (b Bool) Sub(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		if b.value {
			x.value -= 1
		}
		return x, nil
	case Bool:
		var n float64
		if b.value {
			n += 1
		}
		if x.value {
			n -= 1
		}
		return CreateFloat(n), nil
	case null:
		var n float64
		if b.value {
			n += 1
		}
		return CreateFloat(n), nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (b Bool) Mul(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		if b.value {
			return x, nil
		}
		return CreateFloat(0), nil
	case Bool:
		var n float64
		if b.value {
			n += 1
		}
		if !x.value {
			n = 0
		}
		return CreateFloat(n), nil
	case null:
		return CreateFloat(0), nil
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (b Bool) Div(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		return CreateFloat(1 / x.value), nil
	case Bool:
		if !x.value {
			return nil, ErrZero
		}
		var n float64
		if b.value {
			n += 1
		}
		return CreateFloat(n), nil
	case null:
		return nil, ErrZero
	case undefined:
		return CreateFloat(math.NaN()), nil
	default:
		return nil, ErrIncompatible
	}
}

func (b Bool) Compare(other Value) (int, error) {
	x, ok := other.(Bool)
	if !ok {
		return 0, ErrIncompatible
	}
	var res int
	if b.value == x.value {
		// pass
	} else if b.value {
		res++
	} else {
		res--
	}
	return res, nil
}

func (b Bool) String() string {
	return strconv.FormatBool(b.value)
}

func (_ Bool) Type() string {
	return "boolean"
}
