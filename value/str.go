package value

import (
	"fmt"
	"strings"
	"unicode"
)

type Str struct {
	value string
}

func CreateString(s string) Value {
	return Str{
		value: s,
	}
}

func (s Str) True() bool {
	return s.value != ""
}

func (s Str) Spread() []Value {
	var list []Value
	for _, c := range strings.Split(s.value, "") {
		list = append(list, CreateString(c))
	}
	return list
}

func (s Str) Len() int {
	return len(s.value)
}

func (s Str) At(ix Value) (Value, error) {
	x, ok := ix.(Float)
	if !ok {
		return nil, ErrOperation
	}
	i := int(x.value)
	if i < 0 || i >= len(s.value) {
		return nil, ErrIndex
	}
	char := s.value[i]
	return CreateString(string(char)), nil
}

func (s Str) Get(prop string) (Value, error) {
	if prop != "length" {
		return Undefined(), nil
	}
	return CreateFloat(float64(len(s.value))), nil
}

func (s Str) Call(fn string, args []Value) (Value, error) {
	call, ok := stringPrototype[fn]
	if !ok {
		return nil, fmt.Errorf("%s not defined on string", fn)
	}
	return call(s, args)
}

func (s Str) Add(other Value) (Value, error) {
	switch x := other.(type) {
	case Float:
		s.value = fmt.Sprintf("%s%f", s.value, x.value)
	case Str:
		s.value += x.value
	case Bool:
		s.value = fmt.Sprintf("%s%t", s.value, x.value)
	case null:
		s.value += "null"
	case undefined:
		s.value += "undefined"
	default:
		return nil, ErrIncompatible
	}
	return s, nil
}

func (s Str) Compare(other Value) (int, error) {
	x, ok := other.(Str)
	if !ok {
		return 0, ErrIncompatible
	}
	return strings.Compare(s.value, x.value), nil
}

func (s Str) String() string {
	return s.value
}

func (b Str) Type() string {
	return "string"
}

var stringPrototype = map[string]ValueFunc[Str]{
	"at":          CheckArity(1, strAt),
	"concat":      CheckArity(-1, strConcat),
	"endsWith":    CheckArity(1, strEndsWith),
	"includes":    CheckArity(1, strIncludes),
	"indexOf":     CheckArity(1, strIndexOf),
	"padEnd":      CheckArity(1, strPadEnd),
	"padStart":    CheckArity(1, strPadStart),
	"repeat":      CheckArity(1, strRepeat),
	"replace":     CheckArity(2, strReplace),
	"replaceAll":  CheckArity(2, strReplaceAll),
	"slice":       CheckArity(1, strSlice),
	"split":       CheckArity(0, strSplit),
	"startsWith":  CheckArity(1, strStartsWith),
	"substring":   CheckArity(1, strSubstr),
	"toUpperCase": CheckArity(0, strUpper),
	"toLowerCase": CheckArity(0, strLower),
	"trim":        CheckArity(0, strTrim),
	"trimEnd":     CheckArity(0, strTrimRight),
	"trimStart":   CheckArity(0, strTrimLeft),
	"trimLeft":    CheckArity(0, strTrimLeft),
	"trimRight":   CheckArity(0, strTrimRight),
}

func strAt(s Str, args []Value) (Value, error) {
	return s.At(args[0])
}

func strConcat(s Str, args []Value) (Value, error) {
	var list []string
	list = append(list, s.String())
	for _, a := range args {
		list = append(list, a.String())
	}
	return CreateString(strings.Join(list, "")), nil
}

func strEndsWith(s Str, args []Value) (Value, error) {
	var (
		offset int
		err    error
	)
	if len(args) >= 2 {
		offset, err = toNativeInt(args[1])
		if err != nil {
			return nil, err
		}
		if offset > len(s.value) || offset < 0 {
			return Undefined(), ErrIndex
		}
	}
	ok := strings.HasSuffix(s.value[offset:], args[0].String())
	return CreateBool(ok), nil
}

func strIncludes(s Str, args []Value) (Value, error) {
	var (
		offset int
		err    error
	)
	if len(args) >= 2 {
		offset, err = toNativeInt(args[1])
		if err != nil {
			return nil, err
		}
		if offset > len(s.value) || offset < 0 {
			return Undefined(), ErrIndex
		}
	}
	ok := strings.Contains(s.value[offset:], args[0].String())
	return CreateBool(ok), nil
}

func strIndexOf(s Str, args []Value) (Value, error) {
	var (
		pat = args[0].String()
		pos int
		err error
	)
	if len(args) >= 2 {
		pos, err = toNativeInt(args[1])
		if err != nil {
			return nil, err
		}
		if pos < 0 {
			pos = len(s.value) + pos
		}
	}
	x := strings.Index(s.value[pos:], pat)
	return CreateFloat(float64(x)), nil
}

func strPadEnd(s Str, args []Value) (Value, error) {
	var (
		size int
		err  error
		char = " "
	)
	if size, err = toNativeInt(args[0]); err != nil {
		return nil, err
	}
	if len(s.value) >= size {
		return s, nil
	}
	if len(args) >= 2 {
		char = args[1].String()
	}
	str := strings.Repeat(char, size-len(s.value))
	return CreateString(s.value + str), nil
}

func strPadStart(s Str, args []Value) (Value, error) {
	var (
		size int
		err  error
		char = " "
	)
	if size, err = toNativeInt(args[0]); err != nil {
		return nil, err
	}
	if len(s.value) >= size {
		return s, nil
	}
	if len(args) >= 2 {
		char = args[1].String()
	}
	str := strings.Repeat(char, size-len(s.value))
	return CreateString(str + s.value), nil
}

func strRepeat(s Str, args []Value) (Value, error) {
	r, err := toNativeInt(args[0])
	if err != nil {
		return nil, err
	}
	if r <= 0 {
		return nil, ErrArgument
	}
	s.value = strings.Repeat(s.value, r)
	return s, nil
}

func strReplace(s Str, args []Value) (Value, error) {
	pat := args[0].String()
	rep := args[1].String()

	s.value = strings.Replace(s.value, pat, rep, 1)
	return s, nil
}

func strSlice(s Str, args []Value) (Value, error) {
	var (
		beg = 0
		end = len(s.value)
		err error
	)
	switch len(args) {
	case 1:
		beg, err = toNativeInt(args[0])
		if err != nil {
			return nil, err
		}
	case 2:
		beg, err = toNativeInt(args[0])
		if err != nil {
			return nil, err
		}
		end, err = toNativeInt(args[1])
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrArgument
	}
	if beg < 0 {
		beg = len(s.value) + beg
	}
	if end < 0 {
		end = len(s.value) + end
	}
	if beg > len(s.value) || beg >= end {
		return CreateString(""), nil
	}
	return CreateString(s.value[beg:end]), nil
}

func strSplit(s Str, args []Value) (Value, error) {
	var (
		limit int
		err   error
		sep   string
	)
	switch len(args) {
	case 0:
		return CreateArray([]Value{s}), nil
	case 1:
		limit--
		sep = args[0].String()
	case 2:
		sep = args[0].String()
		limit, err = toNativeInt(args[1])
		if err != nil {
			return nil, err
		}
		if limit == 0 {
			return CreateArray(nil), nil
		}
	default:
		return nil, ErrArgument
	}
	var (
		parts = strings.Split(s.value, sep)
		list  []Value
	)
	for i := 0; i < limit && i < len(parts); i++ {
		list = append(list, CreateString(parts[i]))
	}
	return CreateArray(list), nil
}

func strStartsWith(s Str, args []Value) (Value, error) {
	var (
		offset int
		err    error
	)
	if len(args) >= 2 {
		offset, err = toNativeInt(args[1])
		if err != nil {
			return nil, err
		}
		if offset > len(s.value) || offset < 0 {
			return Undefined(), ErrIndex
		}
	}
	ok := strings.HasPrefix(s.value[offset:], args[0].String())
	return CreateBool(ok), nil
}

func strSubstr(s Str, args []Value) (Value, error) {
	start, err := toNativeInt(args[0])
	if err != nil {
		return nil, err
	}
	end := len(s.value)
	if len(args) >= 2 {
		end, err = toNativeInt(args[1])
		if err != nil {
			return nil, err
		}
	}
	if start >= end {
		return Undefined(), ErrIndex
	}
	return CreateString(s.value[start:end]), nil
}

func strTrim(s Str, _ []Value) (Value, error) {
	str := strings.TrimSpace(s.value)
	return CreateString(str), nil
}

func strTrimLeft(s Str, _ []Value) (Value, error) {
	str := strings.TrimLeftFunc(s.value, unicode.IsSpace)
	return CreateString(str), nil
}

func strTrimRight(s Str, _ []Value) (Value, error) {
	str := strings.TrimRightFunc(s.value, unicode.IsSpace)
	return CreateString(str), nil
}

func strReplaceAll(s Str, args []Value) (Value, error) {
	s1, ok := args[0].(Str)
	if !ok {
		return nil, ErrIncompatible
	}
	s2, ok := args[1].(Str)
	if !ok {
		return nil, ErrIncompatible
	}
	s.value = strings.ReplaceAll(s.value, s1.value, s2.value)
	return s, nil
}

func strUpper(s Str, _ []Value) (Value, error) {
	s.value = strings.ToUpper(s.value)
	return s, nil
}

func strLower(s Str, _ []Value) (Value, error) {
	s.value = strings.ToUpper(s.value)
	return s, nil
}
