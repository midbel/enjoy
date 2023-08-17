package ast

type Node interface {
	// Token() Token
}

type ValueNode[T float64 | string | bool] struct {
	Literal T
}

func CreateValue[T float64 | string | bool](v T) ValueNode[T] {
	return ValueNode[T]{
		Literal: v,
	}
}

type NullNode struct{}

type UndefinedNode struct{}

type ArrayNode struct {
	List []Node
}

type ObjectNode struct {
	List map[string]Node
}

type VarNode struct {
	Ident string
}

type DiscardNode struct{}

type TemplateNode struct {
	Nodes []Node
}

type SeqNode struct {
	Nodes []Node
}

type SpreadNode struct {
	Node
}

type BlockNode struct {
	Nodes []Node
}

type BreakNode struct {
	Label string
}

type ContinueNode struct {
	Label string
}

type MemberNode struct {
	Curr Node
	Next Node
}

type LetNode struct {
	Ident Node
	Expr  Node
}

type ConstNode struct {
	Ident Node
	Expr  Node
}

type AssignNode struct {
	Ident Node
	Expr  Node
}

type BindingArrayNode struct {
	List []Node
}

type BindingObjectNode struct {
	List map[string]Node
}

type UnaryNode struct {
	Op   rune
	Expr Node
}

type BinaryNode struct {
	Op    rune
	Left  Node
	Right Node
}

type IndexNode struct {
	Expr  Node
	Index Node
}

type TryNode struct {
	Try     Node
	Catch   Node
	Finally Node
}

type CatchNode struct {
	Ident Node
	Body  Node
}

type ThrowNode struct {
	Node
}

type IfNode struct {
	Cdt Node
	Csq Node
	Alt Node
}

type SwitchNode struct {
	Cdt   Node
	Cases []Node
}

type CaseNode struct {
	Cdt  Node
	Body Node
}

type WhileNode struct {
	Cdt  Node
	Body Node
}

type DoNode struct {
	Cdt  Node
	Body Node
}

type ForNode struct {
	Init Node
	Cdt  Node
	Incr Node
	Body Node
}

type ForInNode struct {
	Ident string
	Expr  Node
	Body  Node
}

type ForOfNode struct {
	Ident string
	Expr  Node
	Body  Node
}

type FuncNode struct {
	Ident string
	Args  Node
	Body  Node
}

type ArrowNode struct {
	Args Node
	Body Node
}

type ReturnNode struct {
	Node
}

type CallNode struct {
	Ident Node
	Args  Node
}

type TypeofNode struct {
	Node
}
