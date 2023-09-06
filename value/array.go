package value

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/midbel/enjoy/env"
)

type Array struct {
	values []Value
}

func CreateArray(vs []Value) Value {
	return Array{
		values: vs,
	}
}

func (a Array) True() bool {
	return a.Len() > 0
}

func (a Array) Spread() []Value {
	return slices.Clone(a.values)
}

func (a Array) Len() int {
	return len(a.values)
}

func (a Array) At(ix Value) (Value, error) {
	x, ok := ix.(Float)
	if !ok {
		return nil, ErrOperation
	}
	i := int(x.value)
	if i < 0 || i >= a.Len() {
		return Undefined(), ErrIndex
	}
	return a.values[i], nil
}

func (a Array) Get(prop string) (Value, error) {
	if prop != "length" {
		return Undefined(), nil
	}
	return CreateFloat(float64(len(a.values))), nil
}

func (a Array) Call(fn string, args []Value) (Value, error) {
	call, ok := arrayPrototype[fn]
	if !ok {
		return nil, fmt.Errorf("%s not defined on array", fn)
	}
	return call(a, args)
}

func (a Array) String() string {
	var str strings.Builder
	str.WriteRune('[')
	for i, v := range a.values {
		if i > 0 {
			str.WriteRune(',')
			str.WriteRune(' ')
		}
		str.WriteString(v.String())
	}
	str.WriteRune(']')
	return str.String()
}

func (_ Array) Type() string {
	return "array"
}

var arrayPrototype = map[string]ValueFunc[Array]{
	"at":            CheckArity(1, arrayAt),
	"concat":        CheckArity(-1, arrayConcat),
	"entries":       CheckArity(0, arrayEntries),
	"every":         CheckArity(1, arrayEvery),
	"forEach":       CheckArity(1, arrayForEach),
	"fill":          CheckArity(1, arrayFill),
	"filter":        CheckArity(1, arrayFilter),
	"find":          CheckArity(1, arrayFind),
	"findIndex":     CheckArity(1, arrayFindIndex),
	"findLast":      CheckArity(1, arrayFindLast),
	"findLastIndex": CheckArity(1, arrayFindLastIndex),
	"flat":          CheckArity(0, arrayFlat),
	"flatMap":       CheckArity(1, arrayFlatMap),
	"includes":      CheckArity(1, arrayIncludes),
	"indexOf":       CheckArity(1, arrayIndexOf),
	"join":          CheckArity(0, arrayJoin),
	// "keys":          arrayKeys,
	"lastIndexOf": CheckArity(1, arrayLastIndexOf),
	"map":         CheckArity(1, arrayMap),
	"pop":         CheckArity(0, arrayPop),
	"push":        CheckArity(1, arrayPush),
	"reduce":      CheckArity(1, arrayReduce),
	"reduceRight": CheckArity(1, arrayReduceRight),
	"reverse":     CheckArity(0, arrayReverse),
	"shift":       CheckArity(0, arrayShift),
	"slice":       CheckArity(0, arraySlice),
	"some":        CheckArity(1, arraySome),
	"sort":        CheckArity(0, arraySort),
	"splice":      CheckArity(1, arraySplice),
	"unshift":     CheckArity(0, arrayUnshift),
	// "values":        arrayValues,
	"with": CheckArity(2, arrayWith),
}

func arrayAt(a Array, args []Value) (Value, error) {
	return a.At(args[0])
}

func arrayConcat(a Array, args []Value) (Value, error) {
	arr := slices.Clone(a.values)
	for i := range args {
		if x, ok := args[i].(Array); ok {
			arr = append(arr, x.values...)
		} else {
			arr = append(arr, args[i])
		}
	}
	return CreateArray(arr), nil
}

func arrayEntries(a Array, args []Value) (Value, error) {
	return nil, ErrImplemented
}

func arrayEvery(a Array, args []Value) (Value, error) {
	var (
		errFalse = errors.New("false")
		err      error
	)
	err = arrayApplyFunc(a, args, func(v Value, _ int, err error) error {
		if err == nil && !v.True() {
			return errFalse
		}
		return nil
	})
	if err != nil && !errors.Is(err, errFalse) {
		return Undefined(), err
	}
	if errors.Is(err, errFalse) {
		return CreateBool(false), nil
	}
	return CreateBool(true), nil
}

func arrayFill(a Array, args []Value) (Value, error) {
	var (
		val = args[0]
		beg = 0
		end = len(a.values)
		err error
	)
	if len(args) >= 2 {
		beg, err = toNativeInt(args[1])
		if err != nil {
			return Undefined(), err
		}
		beg = normalizeIndex(beg, len(a.values))
	}
	if len(args) >= 3 {
		end, err = toNativeInt(args[2])
		if err != nil {
			return Undefined(), err
		}
		end = normalizeIndex(end, len(a.values))
	}
	if end <= beg {
		return a, nil
	}
	for i := range a.values[beg:end] {
		a.values[beg+i] = val
	}
	return a, nil
}

func arrayFilter(a Array, args []Value) (Value, error) {
	var (
		list []Value
		err  error
	)
	err = arrayApplyFunc(a, args, func(v Value, _ int, err error) error {
		if err == nil && v.True() {
			list = append(list, v)
		}
		return err
	})
	return CreateArray(list), err
}

func arrayFind(a Array, args []Value) (Value, error) {
	var (
		val      Value
		err      error
		errFound = errors.New("found")
	)
	err = arrayApplyFunc(a, args, func(v Value, _ int, err error) error {
		if err == nil && v.True() {
			val = v
			return errFound
		}
		return err
	})
	if err != nil && !errors.Is(err, errFound) {
		return Undefined(), err
	}
	return val, err
}

func arrayFindIndex(a Array, args []Value) (Value, error) {
	var (
		val      Value = CreateFloat(-1)
		err      error
		errFound = errors.New("found")
	)
	err = arrayApplyFunc(a, args, func(v Value, i int, err error) error {
		if err == nil && v.True() {
			val = CreateFloat(float64(i))
			return errFound
		}
		return err
	})
	if err != nil && !errors.Is(err, errFound) {
		return Undefined(), err
	}
	return val, err
}

func arrayFindLast(a Array, args []Value) (Value, error) {
	var (
		val Value
		err error
	)
	err = arrayApplyFunc(a, args, func(v Value, _ int, err error) error {
		if err == nil && v.True() {
			val = v
		}
		return err
	})
	return val, err
}

func arrayFindLastIndex(a Array, args []Value) (Value, error) {
	var (
		val Value = CreateFloat(-1)
		err error
	)
	err = arrayApplyFunc(a, args, func(v Value, i int, err error) error {
		if err == nil && v.True() {
			val = CreateFloat(float64(i))
		}
		return err
	})
	return val, err
}

func arrayFlat(a Array, args []Value) (Value, error) {
	var (
		level   = -1
		err     error
		flatten func(Value, int) []Value
	)
	if len(args) >= 1 {
		level, err = toNativeInt(args[0])
		if err != nil {
			return nil, err
		}
	}
	flatten = func(v Value, lvl int) []Value {
		a, ok := v.(Array)
		if !ok || lvl == 0 {
			return []Value{v}
		}
		var list []Value
		for i := range a.values {
			xs := flatten(a.values[i], lvl-1)
			list = append(list, xs...)
		}
		return list
	}
	list := flatten(a, level)
	return CreateArray(list), nil
}

func arrayFlatMap(a Array, args []Value) (Value, error) {
	fn, ok := args[0].(Func)
	if !ok {
		return nil, ErrOperation
	}
	var (
		ident   string
		index   string
		array   string
		err     error
		flatten func(Value, Value, int) ([]Value, error)
	)

	if len(fn.Params) >= 1 {
		ident = fn.Params[0].Name
	}
	if len(fn.Params) >= 2 {
		index = fn.Params[1].Name
	}
	if len(fn.Params) >= 3 {
		array = fn.Params[2].Name
	}

	flatten = func(v, arr Value, i int) ([]Value, error) {
		a, ok := v.(Array)
		if !ok {
			tmp := env.EnclosedEnv[Value](fn.Env)
			if ident != "" {
				tmp.Define(ident, v, false)
			}
			if index != "" {
				tmp.Define(index, CreateFloat(float64(i)), false)
			}
			if array != "" {
				tmp.Define(array, arr, false)
			}
			v, err := fn.Body.Eval(tmp)
			return []Value{v}, err
		}
		var list []Value
		for i := range a.values {
			xs, err := flatten(a.values[i], a, i)
			if err != nil {
				return nil, err
			}
			list = append(list, xs...)
		}
		return list, nil
	}
	list, err := flatten(a, a, 0)
	if err != nil {
		return nil, err
	}
	return CreateArray(list), nil
}

func arrayForEach(a Array, args []Value) (Value, error) {
	return Null(), arrayApplyFunc(a, args, func(_ Value, _ int, err error) error {
		return err
	})
}

func arrayIncludes(a Array, args []Value) (Value, error) {
	var (
		val = args[0]
		beg int
		err error
	)
	if len(args) >= 2 {
		if beg, err = toNativeInt(args[1]); err != nil {
			return nil, err
		}
		beg = normalizeIndex(beg, len(a.values))
	}
	for i := range a.values[beg:] {
		v, err := Compare(a.values[i], val, func(res int) bool {
			return res == 0
		})
		if err != nil {
			return nil, err
		}
		if v.True() {
			return CreateBool(true), nil
		}
	}
	return CreateBool(false), nil
}

func arrayIndexOf(a Array, args []Value) (Value, error) {
	var (
		val = args[0]
		beg int
		err error
	)
	if len(args) >= 2 {
		if beg, err = toNativeInt(args[1]); err != nil {
			return nil, err
		}
		beg = normalizeIndex(beg, len(a.values))
	}
	for i := range a.values[beg:] {
		v, err := Compare(a.values[i], val, func(res int) bool {
			return res == 0
		})
		if err != nil {
			return nil, err
		}
		if v.True() {
			return CreateFloat(float64(i)), nil
		}
	}
	return CreateFloat(-1), nil
}

func arrayJoin(a Array, args []Value) (Value, error) {
	var (
		list []string
		sep  = ","
	)
	if len(args) >= 1 {
		sep = args[0].String()
	}
	for i := range a.values {
		list = append(list, a.values[i].String())
	}
	res := strings.Join(list, sep)
	return CreateString(res), nil
}

func arrayKeys(a Array, args []Value) (Value, error) {
	return nil, ErrImplemented
}

func arrayLastIndexOf(a Array, args []Value) (Value, error) {
	var (
		val = args[0]
		beg int
		err error
	)
	if len(args) >= 2 {
		if beg, err = toNativeInt(args[1]); err != nil {
			return nil, err
		}
		beg = normalizeIndex(beg, len(a.values))
	}
	for i := beg; i >= 0; i-- {
		v, err := Compare(a.values[i], val, func(res int) bool {
			return res == 0
		})
		if err != nil {
			return nil, err
		}
		if v.True() {
			return CreateFloat(float64(i)), nil
		}
	}
	return CreateFloat(-1), nil
}

func arrayMap(a Array, args []Value) (Value, error) {
	var (
		list []Value
		err  error
	)
	err = arrayApplyFunc(a, args, func(v Value, _ int, err error) error {
		if err == nil {
			list = append(list, v)
		}
		return err
	})
	return CreateArray(list), err
}

func arrayPop(a Array, args []Value) (Value, error) {
	n := len(a.values)
	if n == 0 {
		return Undefined(), nil
	}
	n--
	x := a.values[n]
	a.values = a.values[:n]
	return x, nil
}

func arrayPush(a Array, args []Value) (Value, error) {
	for i := range args {
		a.values = append(a.values, args[i])
	}
	return a, nil
}

func arrayReduce(a Array, args []Value) (Value, error) {
	return arrayApplyReduce(a, args)
}

func arrayReduceRight(a Array, args []Value) (Value, error) {
	values := slices.Clone(a.values)
	slices.Reverse(values)
	return arrayApplyReduce(CreateArray(values).(Array), args)
}

func arrayReverse(a Array, args []Value) (Value, error) {
	slices.Reverse(a.values)
	return a, nil
}

func arrayShift(a Array, args []Value) (Value, error) {
	n := len(a.values)
	if n == 0 {
		return Undefined(), nil
	}
	x := a.values[0]
	a.values = a.values[1:]
	return x, nil
}

func arraySlice(a Array, args []Value) (Value, error) {
	var (
		beg = 0
		end = len(a.values)
		err error
	)
	if len(args) >= 1 {
		beg, err = toNativeInt(args[0])
		if err != nil {
			return nil, err
		}
		beg = normalizeIndex(beg, len(a.values))
	}
	if len(args) >= 2 {
		end, err = toNativeInt(args[1])
		if err != nil {
			return nil, err
		}
		end = normalizeIndex(end, len(a.values))
	}
	arr := slices.Clone(a.values[beg:end])
	return CreateArray(arr), nil
}

func arraySome(a Array, args []Value) (Value, error) {
	var (
		errTrue = errors.New("true")
		err     error
	)
	err = arrayApplyFunc(a, args, func(v Value, _ int, err error) error {
		if err == nil && v.True() {
			return errTrue
		}
		return err
	})
	if err != nil && !errors.Is(err, errTrue) {
		return Undefined(), err
	}
	if errors.Is(err, errTrue) {
		return CreateBool(true), nil
	}
	return CreateBool(false), nil
}

func arraySort(a Array, args []Value) (Value, error) {
	slices.SortFunc(a.values, func(v1, v2 Value) int {
		res, _ := compare(v1, v2)
		return res
	})
	return a, nil
}

func arraySplice(a Array, args []Value) (Value, error) {
	var (
		start int
		size  int
		err   error
		list  []Value
		rest  []Value
	)
	if start, err = toNativeInt(args[0]); err != nil {
		return nil, err
	}
	start = normalizeIndex(start, len(a.values))
	if len(args) >= 2 {
		size, err = toNativeInt(args[1])
		if err != nil {
			return nil, err
		}
		if size < 0 {
			return nil, fmt.Errorf("negative count")
		}
		if start+size >= len(a.values) {
			a.values, list = a.values[:start], a.values[start:]
			size = 0
		} else {
			list = a.values[start:start+size]
			a.values = append(a.values[:start], a.values[start+size:]...)
		}
	} else {
		a.values, list = a.values[:start], a.values[start:]
		return CreateArray(list), nil
	}
	if len(args) >= 3 {
		a.values = append(a.values[:start], append(rest, a.values[start+size:]...)...)
		fmt.Println(a.values)
	}
	return CreateArray(list), nil
}

func arrayUnshift(a Array, args []Value) (Value, error) {
	a.values = append(slices.Clone(args), a.values...)
	return a, nil
}

func arrayValues(a Array, args []Value) (Value, error) {
	return nil, ErrImplemented
}

func arrayWith(a Array, args []Value) (Value, error) {
	var (
		beg int
		err error
	)
	if beg, err = toNativeInt(args[0]); err != nil {
		return nil, err
	}
	beg = normalizeIndex(beg, len(a.values))

	arr := slices.Clone(a.values)
	arr[beg] = args[1]
	return CreateArray(arr), nil
}

func arrayApplyReduce(arr Array, args []Value) (Value, error) {
	fn, ok := args[0].(Func)
	if !ok {
		return nil, ErrOperation
	}
	var (
		accum   string
		current string
		index   string
		array   string
		res     = Undefined()
		err     error
	)
	if len(fn.Params) >= 1 {
		accum = fn.Params[0].Name
	}
	if len(fn.Params) >= 2 {
		current = fn.Params[1].Name
	}
	if len(fn.Params) >= 3 {
		index = fn.Params[2].Name
	}
	if len(fn.Params) >= 4 {
		index = fn.Params[3].Name
	}
	if len(args) >= 2 {
		res = args[1]
	}
	for i := range arr.values {
		tmp := env.EnclosedEnv[Value](fn.Env)
		if accum != "" {
			tmp.Define(accum, res, false)
		}
		if current != "" {
			tmp.Define(current, arr.values[i], false)
		}
		if index != "" {
			tmp.Define(index, CreateFloat(float64(i)), false)
		}
		if array != "" {
			tmp.Define(array, arr, false)
		}
		res, err = fn.Body.Eval(tmp)
		if err != nil {
			break
		}
	}
	return res, err
}

func arrayApplyFunc(a Array, args []Value, apply func(v Value, i int, err error) error) error {
	fn, ok := args[0].(Func)
	if !ok {
		return ErrOperation
	}
	var (
		ident string
		index string
		array string
	)
	if len(fn.Params) >= 1 {
		ident = fn.Params[0].Name
	}
	if len(fn.Params) >= 2 {
		index = fn.Params[1].Name
	}
	if len(fn.Params) >= 3 {
		array = fn.Params[2].Name
	}
	for i := range a.values {
		tmp := env.EnclosedEnv[Value](fn.Env)
		if ident != "" {
			tmp.Define(ident, a.values[i], false)
		}
		if index != "" {
			tmp.Define(index, CreateFloat(float64(i)), false)
		}
		if array != "" {
			tmp.Define(array, a, false)
		}
		v, err := fn.Body.Eval(tmp)
		if err := apply(v, i, err); err != nil {
			return err
		}
	}
	return nil
}

func normalizeIndex(x, size int) int {
	if x < 0 {
		return x + size
	}
	if x < -size {
		return 0
	}
	if x >= size {
		return size
	}
	return x
}
