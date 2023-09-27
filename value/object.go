package value

import (
	"slices"
	"strings"
)

type Descriptor struct {
	Value
	Writable     bool
	Configurable bool
	Enumerable   bool
}

func createDescriptor(val Value) Descriptor {
	return Descriptor{
		Value:        val,
		Writable:     true,
		Configurable: true,
		Enumerable:   true,
	}
}

type Object struct {
	frozen bool
	sealed bool
	values map[string]Descriptor
}

func CreateObject(list map[string]Value) Value {
	obj := Object{
		values: make(map[string]Descriptor),
	}
	for k, v := range list {
		obj.values[k] = createDescriptor(v)
	}
	return &obj
}

func (o *Object) Keys() Value {
	var list []Value
	for k, v := range o.values {
		if !v.Enumerable {
			continue
		}
		list = append(list, CreateString(k))
	}
	slices.SortFunc(list, func(s1, s2 Value) int {
		return strings.Compare(s1.String(), s2.String())
	})
	return CreateArray(list)
}

func (o *Object) Freeze() {
	o.frozen = true
}

func (o *Object) Seal() {
	o.sealed = true
}

func (o *Object) At(ix Value) (Value, error) {
	v, ok := o.values[ix.String()]
	if !ok {
		return undefined{}, nil
	}
	return v, nil
}

func (o *Object) Get(prop string) (Value, error) {
	if prop == "length" {
		return CreateFloat(float64(len(o.values))), nil
	}
	v, ok := o.values[prop]
	if !ok {
		return undefined{}, nil
	}
	return v, nil
}

func (o *Object) Set(prop string, val Value) (Value, error) {
	if o.frozen {
		return nil, ErrOperation
	}
	return nil, nil
}

func (o *Object) Call(fn string, args []Value) (Value, error) {
	return nil, nil
}

func (o *Object) True() bool {
	return len(o.values) > 0
}

func (o *Object) String() string {
	var str strings.Builder
	str.WriteRune('{')

	var i int
	for k, v := range o.values {
		if i > 0 {
			str.WriteRune(',')
			str.WriteRune(' ')
		}
		i++
		str.WriteString(k)
		str.WriteRune(':')
		str.WriteString(v.String())
	}
	str.WriteRune('}')
	return str.String()
}

func (o *Object) Type() string {
	return "object"
}
