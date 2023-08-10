package env

import (
	"errors"
	"fmt"
)

var (
	ErrDefined    = errors.New("variable already defined")
	ErrNotDefined = errors.New("variable not defined")
	ErrAssign     = errors.New("variable can be assigned")
)

type Environ[T any] interface {
	Define(string, T, bool) error
	Assign(string, T) error
	Resolve(string) (T, error)
}

type ImmutableEnv[T any] struct {
	Environ[T]
}

func Immutable[T any](env Environ[T]) Environ[T] {
	return &ImmutableEnv[T]{
		Environ: env,
	}
}

func (_ ImmutableEnv[T]) Define(_ string, _ T, _ bool) error {
	return ErrAssign
}

func (_ ImmutableEnv[T]) Assign(_ string, _ T) error {
	return ErrAssign
}

type Env[T any] struct {
	parent Environ[T]
	values map[string]value[T]
}

func EmptyEnv[T any]() Environ[T] {
	return EnclosedEnv[T](nil)
}

func EnclosedEnv[T any](parent Environ[T]) Environ[T] {
	e := Env[T]{
		parent: parent,
	}
	e.Clear()
	return &e
}

func (e *Env[T]) Clear() {
	e.values = make(map[string]value[T])
}

func (e *Env[T]) Delete(ident string) {
	delete(e.values, ident)
}

func (e *Env[T]) Define(ident string, val T, ro bool) error {
	if _, ok := e.values[ident]; ok {
		return fmt.Errorf("%s: %w", ident, ErrDefined)
	}
	e.values[ident] = value[T]{
		ro:    ro,
		value: val,
	}
	return nil
}

func (e *Env[T]) Assign(ident string, val T) error {
	v, ok := e.values[ident]
	if ok {
		if v.ro {
			return ErrAssign
		}
		v.value = val
		e.values[ident] = v
		return nil
	}
	if e.parent != nil {
		return e.parent.Assign(ident, val)
	}
	return fmt.Errorf("%s: %w", ident, ErrNotDefined)
}

func (e *Env[T]) Resolve(ident string) (T, error) {
	v, ok := e.values[ident]
	if ok {
		return v.value, nil
	}
	if e.parent != nil {
		return e.parent.Resolve(ident)
	}
	return v.value, fmt.Errorf("%s: %w", ident, ErrNotDefined)
}

type value[T any] struct {
	ro    bool
	value T
}
