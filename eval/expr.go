package eval

import (
	"github.com/midbel/enjoy/ast"
	"github.com/midbel/enjoy/env"
	"github.com/midbel/enjoy/token"
	"github.com/midbel/enjoy/value"
)

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
