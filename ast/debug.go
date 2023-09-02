package ast

import (
	"io"
)

func Debug(n Node, w io.Writer) error {
	return debug(n, 0, w)
}

func debug(n Node, level int, w io.Writer) error {
	switch n.(type) {
	case ValueNode[float64]:
	case ValueNode[string]:
	case ValueNode[bool]:
	case VarNode:
	case DiscardNode:
	case TemplateNode:
	case NullNode:
	case UndefinedNode:
	case ArrayNode:
	case ObjectNode:
	case ExportNode:
	case ExportFromNode:
	case ImportNode:
	case LabelNode:
	default:
	}
	return nil
}
