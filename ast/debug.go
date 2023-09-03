package ast

import (
	"fmt"
	"io"
	"strings"
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
	case DiscardNode:
		fmt.Fprint(w, prefix, "discard")
		fmt.Fprintln(w)
	case TemplateNode:
	case NullNode:
		fmt.Fprint(w, prefix, "null")
		fmt.Fprintln(w)
	case UndefinedNode:
		fmt.Fprint(w, prefix, "undefined")
		fmt.Fprintln(w)
	case SpreadNode:
	case ArrayNode:
	case ObjectNode:
	case ExportNode:
	case ExportFromNode:
	case ImportNode:
	case LetNode:
	case ConstNode:
	case AssignNode:
	case UnaryNode:
	case BinaryNode:
	case TryNode:
	case CatchNode:
	case ThrowNode:
	case IfNode:
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
