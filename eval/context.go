package eval

import (
	"github.com/midbel/enjoy/env"
	"github.com/midbel/enjoy/value"
)

type context struct {
	env.Environ[value.Value]
	modules map[string]Module
}

func defaultContext(ev env.Environ[value.Value]) *context {
	return &context{
		Environ: ev,
		modules: make(map[string]Module),
	}
}

type Module struct {
}

func Load() Module {
	var mod Module
	return mod
}