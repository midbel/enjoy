package eval

import (
	"errors"
	"fmt"

	"github.com/midbel/enjoy/ast"
	"github.com/midbel/enjoy/env"
	"github.com/midbel/enjoy/value"
)

func evalReturn(n ast.ReturnNode, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Node, ev)
	if err == nil {
		err = ErrReturn
	}
	return v, err
}

func evalCall(n ast.CallNode, ev env.Environ[value.Value]) (value.Value, error) {
	var (
		res value.Value
		err error
	)
	switch m := n.Ident.(type) {
	case ast.MemberNode:
		res, err = callMember(m, n.Args, ev)
	default:
		res, err = callDefault(n, ev)
	}
	return res, err
}

func callMember(n ast.MemberNode, args ast.Node, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Curr, ev)
	if err != nil {
		return nil, err
	}
	id, ok := n.Next.(ast.VarNode)
	if !ok {
		return nil, ErrEval
	}
	call, ok := v.(value.Callable)
	if !ok {
		return nil, value.ErrOperation
	}
	values, err := callArgs(args, ev)
	if err != nil {
		return nil, err
	}
	v, err = call.Call(id.Ident, values)
	if err != nil {
		err = fmt.Errorf("%s: %w", id.Ident, err)
	}
	return v, err
}

func callDefault(n ast.CallNode, ev env.Environ[value.Value]) (value.Value, error) {
	call, err := eval(n.Ident, ev)
	if err != nil {
		return nil, err
	}
	args, err := callArgs(n.Args, ev)
	if err != nil {
		return nil, err
	}
	switch call := call.(type) {
	case value.Func:
		return execUserFunc(call, args, ev)
	case value.Builtin:
		return execBuiltinFunc(call, args)
	default:
		return nil, ErrEval
	}
}

func callArgs(n ast.Node, ev env.Environ[value.Value]) ([]value.Value, error) {
	seq, ok := n.(ast.SeqNode)
	if !ok {
		return nil, ErrEval
	}
	var args []value.Value
	for _, a := range seq.Nodes {
		g, err := eval(a, ev)
		if err != nil {
			return nil, err
		}
		args = append(args, g)
	}
	return args, nil
}

func execBuiltinFunc(fn value.Builtin, args []value.Value) (value.Value, error) {
	return fn.Apply(args)
}

func prepareArgs(fn value.Func, args []value.Value, ev env.Environ[value.Value]) (env.Environ[value.Value], error) {
	var (
		tmp = env.EnclosedEnv[value.Value](fn.Env)
		arg value.Value
		err error
	)
	for i := 0; i < len(fn.Params); i++ {
		p := fn.Params[i]
		if i < len(args) {
			arg = args[i]
		} else {
			arg = value.Undefined()
		}
		arg, err = argValue(p, arg, tmp)
		if err != nil {
			return nil, err
		}
		switch arg := arg.(type) {
		case value.Object:
			err = argObject(p, arg, tmp)
		case value.Array:
			err = argArray(p, arg, tmp)
		case value.Spread:
			for _, a := range arg.Spread() {
				if i >= len(fn.Params) {
					break
				}
				if err := tmp.Define(fn.Params[i].Name, a, false); err != nil {
					return nil, err
				}
				i++
			}
		default:
			err = tmp.Define(p.Name, arg, false)
		}
		if err != nil {
			return nil, err
		}
	}
	return tmp, err
}

func argValue(prm value.Parameter, arg value.Value, ev env.Environ[value.Value]) (value.Value, error) {
	if !value.IsUndefined(arg) && !value.IsNull(arg) {
		return arg, nil
	}
	if prm.Value == nil {
		return arg, nil
	}
	switch a := prm.Value.(type) {
	case ast.AssignNode:
		return eval(a.Expr, ev)
	case ast.BindingArrayNode:
	case ast.BindingObjectNode:
	default:
		return eval(prm.Value, ev)
	}
	return arg, nil
}

func argArray(prm value.Parameter, arr value.Array, ev env.Environ[value.Value]) error {
	switch a := prm.Value.(type) {
	case ast.AssignNode:
		prm.Value = a.Ident
		return argArray(prm, arr, ev)
	case ast.BindingArrayNode:
		return bindArray(a, arr, ev, false)
	default:
		if prm.Name == "" {
			return ErrEval
		}
		return ev.Define(prm.Name, arr, false)
	}
	return nil
}

func argObject(prm value.Parameter, obj value.Object, ev env.Environ[value.Value]) error {
	switch a := prm.Value.(type) {
	case ast.AssignNode:
		prm.Value = a.Ident
		return argObject(prm, obj, ev)
	case ast.BindingObjectNode:
		return bindObject(a, obj, ev, false)
	default:
		if prm.Name == "" {
			return ErrEval
		}
		return ev.Define(prm.Name, obj, false)
	}
	return nil
}

func execUserFunc(fn value.Func, args []value.Value, ev env.Environ[value.Value]) (value.Value, error) {
	tmp, err := prepareArgs(fn, args, ev)
	if err != nil {
		return nil, err
	}
	res, err := eval(fn.Body, tmp)
	if errors.Is(err, ErrReturn) {
		err = nil
	}
	return res, err
}

func evalArrow(n ast.ArrowNode, ev env.Environ[value.Value]) (value.Value, error) {
	fn := value.Func{
		Body: EvaluableNode(n.Body),
		Env:  ev,
	}
	switch n := n.Args.(type) {
	case ast.VarNode:
		p := value.Parameter{
			Name: n.Ident,
		}
		fn.Params = append(fn.Params, p)
	case ast.SeqNode:
		for _, a := range n.Nodes {
			var p value.Parameter
			g, ok := a.(ast.VarNode)
			if !ok {
				return nil, ErrEval
			}
			p.Name = g.Ident
			fn.Params = append(fn.Params, p)
		}
	default:
		return nil, ErrEval
	}
	return fn, nil
}

func evalFunc(n ast.FuncNode, ev env.Environ[value.Value]) (value.Value, error) {
	fn := value.Func{
		Ident: n.Ident,
		Body:  EvaluableNode(n.Body),
		Env:   ev,
	}
	seq, ok := n.Args.(ast.SeqNode)
	if !ok {
		return nil, ErrEval
	}
	for _, a := range seq.Nodes {
		var p value.Parameter
		switch g := a.(type) {
		case ast.AssignNode:
			if i, ok := g.Ident.(ast.VarNode); ok {
				p.Name = i.Ident
				p.Value = g.Expr
			} else {
				p.Value = a
			}
		case ast.VarNode:
			p.Name = g.Ident
		case ast.BindingArrayNode, ast.BindingObjectNode:
			p.Value = a
		default:
			return nil, ErrEval
		}
		fn.Params = append(fn.Params, p)
	}
	if fn.Ident != "" {
		if err := ev.Define(fn.Ident, fn, false); err != nil {
			return nil, err
		}
	}
	return fn, nil
}
