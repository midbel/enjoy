package builtins

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/midbel/enjoy/value"
)

func Json() value.Value {
	obj := value.CreateGlobal("JSON")
	obj.RegisterFunc("parse", value.CheckArity(1, jsonParse))
	obj.RegisterFunc("stringify", value.CheckArity(1, jsonString))
	return obj
}

func Xml() value.Value {
	obj := value.CreateGlobal("XML")
	obj.RegisterFunc("parse", xmlParse)
	return obj
}

func xmlParse(_ value.Global, args []value.Value) (value.Value, error) {
	var (
		r = strings.NewReader(args[0].String())
		d interface{}
	)
	err := xml.NewDecoder(r).Decode(&d)
	if err != nil {
		return nil, err
	}
	return nativeToValues(d)
}

func jsonParse(_ value.Global, args []value.Value) (value.Value, error) {
	var (
		r = strings.NewReader(args[0].String())
		d interface{}
	)
	err := json.NewDecoder(r).Decode(&d)
	if err != nil {
		return nil, err
	}
	return nativeToValues(d)
}

func jsonString(_ value.Global, args []value.Value) (value.Value, error) {
	return nil, value.ErrImplemented
}

func nativeToValues(d interface{}) (value.Value, error) {
	switch v := d.(type) {
	case string:
		return value.CreateString(v), nil
	case float64:
		return value.CreateFloat(v), nil
	case bool:
		return value.CreateBool(v), nil
	case []interface{}:
		var list []value.Value
		for i := range v {
			d, err := nativeToValues(v[i])
			if err != nil {
				return nil, err
			}
			list = append(list, d)
		}
		return value.CreateArray(list), nil
	case map[string]interface{}:
		list := make(map[string]value.Value)
		for k, v := range v {
			a, err := nativeToValues(v)
			if err != nil {
				return nil, err
			}
			list[k] = a
		}
		return value.CreateObject(list), nil
	default:
		return nil, fmt.Errorf("%T unsupported json type")
	}
}
