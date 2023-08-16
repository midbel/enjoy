package value

import (
	"fmt"
	"strings"
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
	"toUpperCase": CheckArity(0, strUpperCall),
	"toLowerCase": CheckArity(0, strLowerCall),
	"replaceAll":  CheckArity(2, strReplaceAllCall),
}

func strReplaceAllCall(s Str, args []Value) (Value, error) {
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

func strUpperCall(s Str, _ []Value) (Value, error) {
	s.value = strings.ToUpper(s.value)
	return s, nil
}

func strLowerCall(s Str, _ []Value) (Value, error) {
	s.value = strings.ToUpper(s.value)
	return s, nil
}
