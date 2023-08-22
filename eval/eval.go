package eval

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/midbel/enjoy/ast"
	"github.com/midbel/enjoy/builtins"
	"github.com/midbel/enjoy/env"
	"github.com/midbel/enjoy/parser"
	"github.com/midbel/enjoy/token"
	"github.com/midbel/enjoy/value"
)

var (
	ErrBreak    = errors.New("break")
	ErrContinue = errors.New("continue")
	ErrReturn   = errors.New("return")
	ErrThrow    = errors.New("throw")
	ErrEval     = errors.New("node can not be evalualed in current context")
)

func Default() env.Environ[value.Value] {
	top := env.EmptyEnv[value.Value]()
	top.Define("console", builtins.Console(), true)
	top.Define("Math", builtins.Math(), true)
	top.Define("Object", builtins.Object(), true)
	top.Define("JSON", builtins.Json(), true)
	top.Define("XML", builtins.Xml(), true)

	top.Define("parseInt", builtins.ParseInt(), true)
	top.Define("parseFloat", builtins.ParseFloat(), true)
	top.Define("print", builtins.Print(), true)

	return env.Immutable(top)
}

type evaluableNode struct {
	ast.Node
}

func EvaluableNode(n ast.Node) value.Evaluable {
	return evaluableNode{
		Node: n,
	}
}

func (e evaluableNode) Eval(ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(e.Node, ev)
	if err == ErrReturn {
		err = nil
	}
	return v, err
}

func EvalDefault(r io.Reader) (value.Value, error) {
	return Eval(r, env.EnclosedEnv(Default()))
}

func Eval(r io.Reader, ev env.Environ[value.Value]) (value.Value, error) {
	n, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}
	return eval(n, ev)
}

func eval(node ast.Node, ev env.Environ[value.Value]) (value.Value, error) {
	switch n := node.(type) {
	case ast.NullNode:
		return value.Null(), nil
	case ast.UndefinedNode, ast.DiscardNode:
		return value.Undefined(), nil
	case ast.ValueNode[string]:
		return value.CreateString(n.Literal), nil
	case ast.ValueNode[float64]:
		return value.CreateFloat(n.Literal), nil
	case ast.ValueNode[bool]:
		return value.CreateBool(n.Literal), nil
	case ast.TemplateNode:
		return evalTemplate(n, ev)
	case ast.VarNode:
		return ev.Resolve(n.Ident)
	case ast.ObjectNode:
		return evalObject(n, ev)
	case ast.ArrayNode:
		return evalArray(n, ev)
	case ast.SpreadNode:
		return evalSpread(n, ev)
	case ast.TypeofNode:
		return evalTypeOf(n, ev)
	case ast.IndexNode:
		return evalIndex(n, ev)
	case ast.MemberNode:
		return evalMember(n, ev)
	case ast.SeqNode:
		return evalSeq(n, ev)
	case ast.BlockNode:
		return evalBlock(n, ev)
	case ast.LabelNode:
		return evalLabel(n, ev)
	case ast.BreakNode:
		return nil, ErrBreak
	case ast.ContinueNode:
		return nil, ErrContinue
	case ast.LetNode:
		return evalLet(n, ev)
	case ast.ConstNode:
		return evalConst(n, ev)
	case ast.AssignNode:
		return evalAssign(n, ev)
	case ast.UnaryNode:
		return evalUnary(n, ev)
	case ast.BinaryNode:
		return evalBinary(n, ev)
	case ast.TryNode:
	case ast.CatchNode:
	case ast.ThrowNode:
	case ast.IfNode:
		return evalIf(n, ev)
	case ast.SwitchNode:
	case ast.WhileNode:
		return evalWhile(n, ev)
	case ast.DoNode:
		return evalDo(n, ev)
	case ast.ForNode:
		return evalFor(n, ev)
	case ast.FuncNode:
		return evalFunc(n, ev)
	case ast.ArrowNode:
		return evalArrow(n, ev)
	case ast.CallNode:
		return evalCall(n, ev)
	case evaluableNode:
		return eval(n.Node, ev)
	case ast.ReturnNode:
		return evalReturn(n, ev)
	default:
		return nil, fmt.Errorf("node type %T not recognized", node)
	}
	return nil, nil
}

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

func argObject(prm value.Parameter, obj value.Object, ev env.Environ[value.Value]) error {
	switch a := prm.Value.(type) {
	case ast.AssignNode:
		prm.Value = a.Ident
		return argObject(prm, obj, ev)
	case ast.BindingObjectNode:
		for k, n := range a.List {
			v, err := obj.Get(k)
			if err != nil || value.IsUndefined(v) || value.IsNull(v) {
				a, ok := n.(ast.AssignNode)
				if !ok {
					return ErrEval
				}
				v, err = eval(a.Expr, ev)
				if err != nil {
					return err
				}
			}
			if err = ev.Define(k, v, false); err != nil {
				return err
			}
		}
	default:
		if prm.Name == "" {
			return ErrEval
		}
		return ev.Define(prm.Name, obj, false)
	}
	return nil
}

func argArray(prm value.Parameter, arr value.Array, ev env.Environ[value.Value]) error {
	return nil
}

func prepareArgs2(fn value.Func, args []value.Value, ev env.Environ[value.Value]) (env.Environ[value.Value], error) {
	var (
		tmp = env.EnclosedEnv[value.Value](fn.Env)
		arg value.Value
		err error
	)
	for i := 0; i < len(fn.Params); i++ {
		p := fn.Params[i]
		if i >= len(args) {
			arg = value.Undefined()
			if p.Value != nil {
				arg, err = eval(p.Value, ev)
			}
			if err != nil {
				return nil, err
			}
			if err = tmp.Define(p.Name, arg, false); err != nil {
				return nil, err
			}
			continue
		}
		arg = args[i]
		if s, ok := arg.(value.Spread); ok {
			for _, a := range s.Spread() {
				if i >= len(fn.Params) {
					break
				}
				if err := tmp.Define(fn.Params[i].Name, a, false); err != nil {
					return nil, err
				}
				i++
			}
			continue
		}
		if obj, ok := arg.(value.Object); ok && p.Name == "" {
			n, ok := p.Value.(ast.ObjectNode)
			if !ok {
				return nil, ErrEval
			}
			for k := range n.List {
				arg, _ = obj.Get(k)
				if err := tmp.Define(k, arg, false); err != nil {
					return nil, err
				}
			}
			continue
		}
		if p.Name == "" {
			return nil, ErrEval
		}
		if err = tmp.Define(p.Name, arg, false); err != nil {
			return nil, err
		}
	}
	return tmp, nil
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

func evalFor(n ast.ForNode, ev env.Environ[value.Value]) (value.Value, error) {
	return nil, nil
}

func evalDo(n ast.DoNode, ev env.Environ[value.Value]) (value.Value, error) {
	var (
		res value.Value
		err error
	)
	for {
		res, err = eval(n.Body, env.EnclosedEnv(ev))
		if err != nil && !errors.Is(err, ErrContinue) {
			break
		}
		v, err := eval(n.Cdt, ev)
		if err != nil {
			return nil, err
		}
		if !v.True() {
			break
		}
	}
	if errors.Is(err, ErrBreak) {
		err = nil
	}
	return res, err
}

func evalWhile(n ast.WhileNode, ev env.Environ[value.Value]) (value.Value, error) {
	var (
		res value.Value
		err error
	)
	for {
		v, err := eval(n.Cdt, ev)
		if err != nil {
			return nil, err
		}
		if !v.True() {
			break
		}
		res, err = eval(n.Body, env.EnclosedEnv(ev))
		if err != nil && !errors.Is(err, ErrContinue) {
			break
		}
	}
	if errors.Is(err, ErrBreak) {
		err = nil
	}
	return res, err
}

func evalIf(n ast.IfNode, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Cdt, ev)
	if err != nil {
		return nil, err
	}
	if v.True() {
		return eval(n.Csq, env.EnclosedEnv(ev))
	}
	if n.Alt != nil {
		return eval(n.Alt, env.EnclosedEnv(ev))
	}
	return nil, nil
}

func evalThrow(n ast.ThrowNode, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Node, ev)
	if err == nil {
		err = ErrThrow
	}
	return v, err
}

func evalCatch(n ast.CatchNode, ev env.Environ[value.Value]) (value.Value, error) {
	return eval(n.Body, ev)
}

func evalTry(n ast.TryNode, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Try, env.EnclosedEnv(ev))
	if err == nil {
		return v, err
	}
	if n.Catch != nil {
		if _, ok := n.Catch.(ast.CatchNode); ok {
			// define error in env
		}
		v, err = eval(n.Catch, env.EnclosedEnv(ev))
	}
	if n.Finally != nil {
		_, err1 := eval(n.Finally, env.EnclosedEnv(ev))
		if err == nil {
			err = err1
		}
	}
	return v, err
}

func evalBinary(n ast.BinaryNode, ev env.Environ[value.Value]) (value.Value, error) {
	left, err := eval(n.Left, ev)
	if err != nil {
		return nil, err
	}
	right, err := eval(n.Right, ev)
	if err != nil {
		return nil, err
	}
	switch n.Op {
	case token.Nullish:
		if value.IsNull(left) || value.IsUndefined(left) {
			return right, nil
		}
		return left, nil
	case token.Add:
		return addValues(left, right)
	case token.Sub:
		return subValues(left, right)
	case token.Mul:
		return mulValues(left, right)
	case token.Div:
		return divValues(left, right)
	case token.Mod:
		return divValues(left, right)
	case token.Pow:
		return powValues(left, right)
	case token.Lshift:
	case token.Rshift:
	case token.Eq:
		return cmpEq(left, right)
	case token.Seq:
	case token.Ne:
		return cmpNe(left, right)
	case token.Sne:
	case token.Lt:
		return cmpLt(left, right)
	case token.Le:
		return cmpLe(left, right)
	case token.Gt:
		return cmpGt(left, right)
	case token.Ge:
		return cmpGe(left, right)
	case token.Band:
	case token.Bor:
	case token.And:
		return value.CreateBool(left.True() && right.True()), nil
	case token.Or:
		return value.CreateBool(left.True() || right.True()), nil
	default:
		return nil, value.ErrOperation
	}
	return nil, nil
}

func evalUnary(n ast.UnaryNode, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Expr, ev)
	if err != nil {
		return nil, err
	}
	switch n.Op {
	case token.Add:
		return value.Coerce(v)
	case token.Sub:
		return value.Reverse(v)
	case token.Not:
		return value.CreateBool(!v.True()), nil
	case token.Increment:
		v, err := value.Increment(v)
		if err == nil {
			i, ok := n.Expr.(ast.VarNode)
			if !ok {
				return nil, ErrEval
			}
			if err := ev.Assign(i.Ident, v); err != nil {
				return nil, err
			}
		}
		return v, err
	case token.Decrement:
		v, err := value.Decrement(v)
		if err == nil {
			i, ok := n.Expr.(ast.VarNode)
			if !ok {
				return nil, ErrEval
			}
			if err := ev.Assign(i.Ident, v); err != nil {
				return nil, err
			}
		}
		return v, err
	default:
		return nil, value.ErrOperation
	}
}

func evalAssign(n ast.AssignNode, ev env.Environ[value.Value]) (value.Value, error) {
	ident, ok := n.Ident.(ast.VarNode)
	if !ok {
		return nil, ErrEval
	}
	v, err := eval(n.Expr, ev)
	if err == nil {
		err = ev.Assign(ident.Ident, v)
	}
	return v, err
}

func evalConst(n ast.ConstNode, ev env.Environ[value.Value]) (value.Value, error) {
	switch x := n.Ident.(type) {
	case ast.VarNode:
		return setVar(x, n.Expr, ev, true)
	case ast.BindingArrayNode:
		return evalBindArray(x, n.Expr, ev, true)
	case ast.BindingObjectNode:
		return evalBindObject(x, n.Expr, ev, true)
	default:
		return nil, ErrEval
	}
}

func evalLet(n ast.LetNode, ev env.Environ[value.Value]) (value.Value, error) {
	switch x := n.Ident.(type) {
	case ast.VarNode:
		return setVar(x, n.Expr, ev, false)
	case ast.BindingArrayNode:
		return evalBindArray(x, n.Expr, ev, false)
	case ast.BindingObjectNode:
		return evalBindObject(x, n.Expr, ev, false)
	default:
		return nil, ErrEval
	}
}

func evalBindObject(o ast.BindingObjectNode, n ast.Node, ev env.Environ[value.Value], ro bool) (value.Value, error) {
	if ro && n == nil {
		return nil, ErrEval
	}
	var (
		res = value.Undefined()
		err error
	)
	if n != nil {
		res, err = eval(n, ev)
		if err != nil {
			return nil, err
		}
	}
	if value.IsUndefined(res) || value.IsNull(res) {
		return nil, ErrEval
	}
	return res, bindObject(o, res, ev, ro)
}

func bindObject(o ast.BindingObjectNode, v value.Value, ev env.Environ[value.Value], ro bool) error {
	obj, ok := v.(value.Object)
	if !ok {
		return nil
	}
	var err error
	for k, n := range o.List {
		v, _ := obj.Get(k)
		if v == nil {
			v = value.Undefined()
		}
		a, ok := n.(ast.AssignNode)
		if !ok {
			return ErrEval
		}
		switch i := a.Ident.(type) {
		case ast.VarNode:
			if value.IsUndefined(v) && a.Expr != nil {
				v, err = eval(a.Expr, ev)
				if err != nil {
					break
				}
			}
			err = ev.Define(i.Ident, v, ro)
		case ast.BindingObjectNode:
			err = bindObject(i, v, ev, ro)
		default:
			return ErrEval
		}
		if err != nil {
			break
		}
	}
	return err
}

func evalBindArray(a ast.BindingArrayNode, n ast.Node, ev env.Environ[value.Value], ro bool) (value.Value, error) {
	if ro && n == nil {
		return nil, ErrEval
	}
	var (
		res = value.Undefined()
		err error
	)
	if n != nil {
		res, err = eval(n, ev)
		if err != nil {
			return nil, err
		}
	}
	if value.IsUndefined(res) || value.IsNull(res) {
		return nil, ErrEval
	}
	return res, bindArray(a, res, ev, ro)
}

func bindArray(a ast.BindingArrayNode, v value.Value, ev env.Environ[value.Value], ro bool) error {
	arr, ok := v.(value.Array)
	if !ok {
		return ErrEval
	}
	var (
		nodes = slices.Clone(a.List)
		err   error
	)
	for i := 0; i < len(nodes); i++ {
		val, _ := arr.At(value.CreateFloat(float64(i)))
		switch n := nodes[i].(type) {
		case ast.DiscardNode:
		case ast.VarNode:
			err = ev.Define(n.Ident, val, ro)
		case ast.AssignNode:
			id, ok := n.Ident.(ast.VarNode)
			if !ok {
				return ErrEval
			}
			if value.IsUndefined(val) || value.IsNull(val) {
				val, err = eval(n.Expr, ev)
			}
			if err == nil {
				err = ev.Define(id.Ident, val, ro)
			}
		case ast.SpreadNode:
			b, ok := n.Node.(ast.BindingArrayNode)
			if !ok {
				return ErrEval
			}
			tmp := slices.Clone(b.List)
			nodes = append(nodes[:i], append(tmp, nodes[i+1:]...)...)
			i--
		case ast.BindingObjectNode:
			err = bindObject(n, val, ev, ro)
		default:
			err = ErrEval
		}
		if err != nil {
			break
		}
	}
	return err
}

func setVar(v ast.VarNode, n ast.Node, ev env.Environ[value.Value], ro bool) (value.Value, error) {
	if ro && n == nil {
		return nil, ErrEval
	}
	var (
		res = value.Undefined()
		err error
	)
	if n != nil {
		res, err = eval(n, ev)
	}
	if err == nil {
		err = ev.Define(v.Ident, res, ro)
	}
	return res, err
}

func evalLabel(n ast.LabelNode, ev env.Environ[value.Value]) (value.Value, error) {
	return nil, nil
}

func evalBlock(n ast.BlockNode, ev env.Environ[value.Value]) (value.Value, error) {
	var (
		res value.Value
		err error
	)
	for _, n := range n.Nodes {
		res, err = eval(n, ev)
		if err != nil {
			break
		}
	}
	return res, err
}

func evalSeq(n ast.SeqNode, ev env.Environ[value.Value]) (value.Value, error) {
	var (
		res value.Value
		err error
	)
	for _, n := range n.Nodes {
		res, err = eval(n, ev)
		if err != nil {
			return nil, err
		}
	}
	return res, err
}

func evalIndex(n ast.IndexNode, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Expr, ev)
	if err != nil {
		return nil, err
	}
	i, err := eval(n.Index, ev)
	if err != nil {
		return nil, err
	}
	return value.At(v, i)
}

func evalMember(n ast.MemberNode, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Curr, ev)
	if err != nil {
		return nil, err
	}
	id, ok := n.Next.(ast.VarNode)
	if !ok {
		return nil, ErrEval
	}
	return value.Get(v, id.Ident)
}

func evalTypeOf(n ast.TypeofNode, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Node, ev)
	if err == nil {
		v = value.CreateString(v.Type())
	}
	return v, err
}

func evalSpread(n ast.SpreadNode, ev env.Environ[value.Value]) (value.Value, error) {
	v, err := eval(n.Node, ev)
	if err != nil {
		return nil, err
	}
	return value.SpreadValue(v)
}

func evalArray(n ast.ArrayNode, ev env.Environ[value.Value]) (value.Value, error) {
	var list []value.Value
	for _, a := range n.List {
		v, err := eval(a, ev)
		if err != nil {
			return nil, err
		}
		if s, ok := v.(value.Spread); ok {
			list = append(list, s.Spread()...)
		} else {
			list = append(list, v)
		}
	}
	return value.CreateArray(list), nil
}

func evalObject(n ast.ObjectNode, ev env.Environ[value.Value]) (value.Value, error) {
	list := make(map[string]value.Value)
	for k, a := range n.List {
		v, err := eval(a, ev)
		if err != nil {
			return nil, err
		}
		list[k] = v
	}
	return value.CreateObject(list), nil
}

func evalTemplate(n ast.TemplateNode, ev env.Environ[value.Value]) (value.Value, error) {
	var list []string
	for _, n := range n.Nodes {
		v, err := eval(n, ev)
		if err != nil {
			return nil, err
		}
		list = append(list, v.String())
	}
	str := strings.Join(list, "")
	return value.CreateString(str), nil
}

func cmpEq(fst, snd value.Value) (value.Value, error) {
	return value.Compare(fst, snd, func(res int) bool { return res == 0 })
}

func cmpNe(fst, snd value.Value) (value.Value, error) {
	return value.Compare(fst, snd, func(res int) bool { return res != 0 })
}

func cmpLt(fst, snd value.Value) (value.Value, error) {
	return value.Compare(fst, snd, func(res int) bool { return res < 0 })
}

func cmpLe(fst, snd value.Value) (value.Value, error) {
	return value.Compare(fst, snd, func(res int) bool { return res <= 0 })
}

func cmpGt(fst, snd value.Value) (value.Value, error) {
	return value.Compare(fst, snd, func(res int) bool { return res > 0 })
}

func cmpGe(fst, snd value.Value) (value.Value, error) {
	return value.Compare(fst, snd, func(res int) bool { return res >= 0 })
}

func addValues(fst, snd value.Value) (value.Value, error) {
	a, ok := fst.(interface {
		Add(value.Value) (value.Value, error)
	})
	if !ok {
		return nil, value.ErrOperation
	}
	return a.Add(snd)
}

func subValues(fst, snd value.Value) (value.Value, error) {
	a, ok := fst.(interface {
		Sub(value.Value) (value.Value, error)
	})
	if !ok {
		return nil, value.ErrOperation
	}
	return a.Sub(snd)
}

func mulValues(fst, snd value.Value) (value.Value, error) {
	a, ok := fst.(interface {
		Mul(value.Value) (value.Value, error)
	})
	if !ok {
		return nil, value.ErrOperation
	}
	return a.Mul(snd)
}

func modValues(fst, snd value.Value) (value.Value, error) {
	a, ok := fst.(interface {
		Mod(value.Value) (value.Value, error)
	})
	if !ok {
		return nil, value.ErrOperation
	}
	return a.Mod(snd)
}

func powValues(fst, snd value.Value) (value.Value, error) {
	a, ok := fst.(interface {
		Pow(value.Value) (value.Value, error)
	})
	if !ok {
		return nil, value.ErrOperation
	}
	return a.Pow(snd)
}

func divValues(fst, snd value.Value) (value.Value, error) {
	a, ok := fst.(interface {
		Div(value.Value) (value.Value, error)
	})
	if !ok {
		return nil, value.ErrOperation
	}
	return a.Div(snd)
}
