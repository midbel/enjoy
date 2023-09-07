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
	case ast.LoopNode:
		return evalLoop(n, ev)
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
	case ast.ImportNode:
		return evalImport(n, ev)
	case ast.ExportNode:
		return evalExport(n, ev)
	case ast.ExportFromNode:
		return evalExportFrom(n, ev)
	default:
		return nil, fmt.Errorf("node type %T not recognized", node)
	}
	return nil, nil
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
	obj, ok := v.(*value.Object)
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
	arr, ok := v.(*value.Array)
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
