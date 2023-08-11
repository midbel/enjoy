package builtins

import (
	"strings"

	"github.com/midbel/enjoy/value"
	"github.com/midbel/sax"
)

func Xml() value.Value {
	obj := value.CreateGlobal("XML")
	obj.RegisterFunc("parse", xmlParse)
	return obj
}

func xmlParse(_ value.Global, args []value.Value) (value.Value, error) {
	rs := sax.New(strings.NewReader(args[0].String()), nil)
	_ = rs
	return nil, nil
}