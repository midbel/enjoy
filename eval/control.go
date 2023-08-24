package eval

import (
	"errors"

	"github.com/midbel/enjoy/ast"
	"github.com/midbel/enjoy/env"
	"github.com/midbel/enjoy/value"
)

func evalLabel(n ast.LabelNode, ev env.Environ[value.Value]) (value.Value, error) {
	return nil, nil
}

func evalLoop(n ast.LoopNode, ev env.Environ[value.Value]) (value.Value, error) {
		switch n.Ident.(type) {
		case ast.IterInNode:
			return evalForIn(n, ev)
		case ast.IterOfNode:
			return evalForOf(n, ev)
		default:
			return nil, ErrEval
		}
}

func evalForIn(n ast.LoopNode, ev env.Environ[value.Value]) (value.Value, error) {
		return nil, nil
}
}

func evalForOf(n ast.LoopNode, ev env.Environ[value.Value]) (value.Value, error) {
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
