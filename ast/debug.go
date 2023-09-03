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
			for i := range n.Nodes {
				if err := debug(n.Nodes[i], level+1, w); err != nil {
					return err
				}
			}
			return nil
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
			for i := range n.List {
				if err := debug(n.List[i], level+1, w); err != nil {
					return err
				}
			}
			return nil
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
	case TryNode:
	case CatchNode:
	case ThrowNode:
	case BlockNode:
		return debugNode(w, "block", prefix, func() error {
			for i := range n.Nodes {
				if err := debug(n.Nodes[i], level+1, w); err != nil {
					return err
				}
			}
			return nil
		})
	case IfNode:
		return debugNode(w, "if", prefix, func() error {
			return debugIf(w, n, level+1)
		})
	case SwitchNode:
	case CaseNode:
	case WhileNode:
	case DoNode:
	case ForNode:
	case LabelNode:
	case BreakNode:
	case ContinueNode:
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
	return nil
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
