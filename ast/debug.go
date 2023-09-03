package ast

import (
	"fmt"
	"io"
	"strings"

	"github.com/midbel/enjoy/token"
)

func Debug(n Node, w io.Writer) error {
	return debug(n, 0, w)
}

func debug(n Node, level int, w io.Writer) error {
	prefix := strings.Repeat(" ", level)
	switch n := n.(type) {
	case ValueNode[float64]:
		fmt.Fprint(w, prefix)
		fmt.Fprintf(w, "number(%f)", n.Literal)
		fmt.Fprintln(w)
	case ValueNode[string]:
		fmt.Fprint(w, prefix)
		fmt.Fprintf(w, "string(%s)", n.Literal)
		fmt.Fprintln(w)
	case ValueNode[bool]:
		fmt.Fprint(w, prefix)
		fmt.Fprintf(w, "boolean(%t)", n.Literal)
		fmt.Fprintln(w)
	case VarNode:
		fmt.Fprint(w, prefix)
		fmt.Fprintf(w, "variable(%s)", n.Ident)
		fmt.Fprintln(w)
	case DiscardNode:
		fmt.Fprint(w, prefix)
		fmt.Fprint(w, "discard")
		fmt.Fprintln(w)
	case TemplateNode:
		return debugNode(w, "template", prefix, func() error {
			return debugList(n.Nodes, level+1, w)
		})
	case NullNode:
		fmt.Fprint(w, prefix)
		fmt.Fprint(w, "null")
		fmt.Fprintln(w)
	case UndefinedNode:
		fmt.Fprint(w, prefix)
		fmt.Fprint(w, "undefined")
		fmt.Fprintln(w)
	case SpreadNode:
		return debugNode(w, "spread", prefix, func() error {
			return debug(n.Node, level+1, w)
		})
	case ArrayNode:
		return debugNode(w, "array", prefix, func() error {
			return debugList(n.List, level+1, w)
		})
	case ObjectNode:
		return debugNode(w, "object", prefix, func() error {
			return nil
		})
	case ExportNode:
		return debugNode(w, "export", prefix, func() error {
			return nil
		})
	case ExportFromNode:
	case ImportNode:
		return debugNode(w, "import", prefix, func() error {
			return nil
		})
	case LetNode:
		return debugNode(w, "let", prefix, func() error {
			if err := debug(n.Ident, level+1, w); err != nil {
				return err
			}
			return debug(n.Expr, level+1, w)
		})
	case ConstNode:
		return debugNode(w, "const", prefix, func() error {
			if err := debug(n.Ident, level+1, w); err != nil {
				return err
			}
			return debug(n.Expr, level+1, w)
		})
	case AssignNode:
		return debugNode(w, "assignment", prefix, func() error {
			if err := debug(n.Ident, level+1, w); err != nil {
				return err
			}
			return debug(n.Expr, level+1, w)
		})
	case UnaryNode:
		return debugUnary(w, n, level)
	case BinaryNode:
		return debugBinary(w, n, level)
	case FuncNode:
		return debugFunc(w, n, level)
	case CallNode:
		return debugCall(w, n, level)
	case ReturnNode:
		return debugNode(w, "return", prefix, func() error {
			return debug(n.Node, level+1, w)
		})
	case TryNode:
	case CatchNode:
	case ThrowNode:
		return debugNode(w, "throw", prefix, func() error {
			return debug(n.Node, level+1, w)
		})
	case BlockNode:
		return debugNode(w, "block", prefix, func() error {
			return debugList(n.Nodes, level+1, w)
		})
	case IfNode:
		return debugNode(w, "if", prefix, func() error {
			return debugIf(w, n, level+1)
		})
	case SwitchNode:
	case CaseNode:
	case WhileNode:
		return debugNode(w, "while", prefix, func() error {
			return nil
		})
	case DoNode:
	case ForNode:
	case LabelNode:
		fmt.Fprint(w, prefix)
		fmt.Fprintf(w, "label(%s)", n.Ident)
		fmt.Fprintln(w)
	case BreakNode:
		fmt.Fprint(w, prefix)
		if n.Label != "" {
			fmt.Fprintf(w, "break(%s)", n.Label)
		} else {
			fmt.Fprint(w, "break")
		}
		fmt.Fprintln(w)
	case ContinueNode:
		fmt.Fprint(w, prefix)
		if n.Label != "" {
			fmt.Fprintf(w, "continue(%s)", n.Label)
		} else {
			fmt.Fprint(w, "continue")
		}
		fmt.Fprintln(w)
	default:
	}
	return nil
}

func debugNode(w io.Writer, name, prefix string, do func() error) error {
	fmt.Fprint(w, prefix)
	fmt.Fprintf(w, "%s {", name)
	fmt.Fprintln(w)
	if do == nil {
		do = func() error { return nil }
	}
	if err := do(); err != nil {
		return err
	}
	fmt.Fprint(w, prefix, "}")
	fmt.Fprintln(w)
	return nil
}

func debugBinary(w io.Writer, n BinaryNode, level int) error {
	var (
		name   string
		prefix = strings.Repeat(" ", level)
	)
	switch n.Op {
	default:
		name = "unknown"
	case token.AddAssign:
		name = "add-assign"
	case token.Add:
		name = "add"
	case token.Sub:
		name = "subtract"
	case token.Mul:
		name = "multiply"
	case token.Div:
		name = "divide"
	case token.Mod:
		name = "modulo"
	case token.Pow:
		name = "power"
	case token.Eq:
		name = "equal"
	case token.Ne:
		name = "not-equal"
	case token.Lt:
		name = "less-than"
	case token.Le:
		name = "less-equal"
	case token.Gt:
		name = "great-than"
	case token.Ge:
		name = "great-equal"
	case token.And:
		name = "and"
	case token.Or:
		name = "or"
	}
	return debugNode(w, name, prefix, func() error {
		if err := debug(n.Left, level+1, w); err != nil {
			return err
		}
		return debug(n.Right, level+1, w)
	})
}

func debugUnary(w io.Writer, n UnaryNode, level int) error {
	var (
		name   string
		prefix = strings.Repeat(" ", level)
	)
	switch n.Op {
	case token.Not:
		name = "not"
	case token.Sub:
		name = "reverse"
	default:
		name = "unknown"
	}
	return debugNode(w, name, prefix, func() error {
		return debug(n.Expr, level+1, w)
	})
}

func debugList(list []Node, level int, w io.Writer) error {
	for i := range list {
		if err := debug(list[i], level, w); err != nil {
			return err
		}
	}
	return nil
}

func debugArgs(args Node, level int, w io.Writer) error {
	seq, ok := args.(SeqNode)
	if !ok {
		return debug(args, level, w)
	}
	return debugList(seq.Nodes, level, w)
}

func debugFunc(w io.Writer, n FuncNode, level int) error {
	var (
		prefix = strings.Repeat(" ", level)
		name   = fmt.Sprintf("function(%s)", n.Ident)
		err    error
	)
	return debugNode(w, name, prefix, func() error {
		prefix = strings.Repeat(" ", level+1)
		err = debugNode(w, "parameters", prefix, func() error {
			return debugArgs(n.Args, level+2, w)
		})
		if err != nil {
			return err
		}
		return debug(n.Body, level+1, w)
	})
}

func debugCall(w io.Writer, n CallNode, level int) error {
	prefix := strings.Repeat(" ", level)
	return debugNode(w, "call", prefix, func() error {
		if err := debug(n.Ident, level+1, w); err != nil {
			return err
		}
		prefix = strings.Repeat(" ", level+1)
		return debugNode(w, "arguments", prefix, func() error {
			return debugArgs(n.Args, level+2, w)
		})
	})
}

func debugIf(w io.Writer, n IfNode, level int) error {
	var (
		prefix = strings.Repeat(" ", level)
		err    error
	)
	err = debugNode(w, "cdt", prefix, func() error {
		return debug(n.Cdt, level+1, w)
	})
	if err != nil {
		return err
	}
	err = debugNode(w, "csq", prefix, func() error {
		return debug(n.Csq, level+1, w)
	})
	if err != nil {
		return err
	}
	if n.Alt == nil {
		return nil
	}
	return debugNode(w, "alt", prefix, func() error {
		return debug(n.Alt, level+1, w)
	})
}
