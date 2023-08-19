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

func Null() NullNode {
	return NullNode{}
}

type UndefinedNode struct{}

func Undefined() UndefinedNode {
	return UndefinedNode{}
}

type ArrayNode struct {
	List []Node
}

func Array(list []Node) ArrayNode {
	return ArrayNode{
		List: list,
	}
}

type ObjectNode struct {
	List map[string]Node
}

func Object(list map[string]Node) ObjectNode {
	return ObjectNode{
		List: list,
	}
}

type LabelNode struct {
	Ident string
}

func Label(ident string) LabelNode {
	return LabelNode{
		Ident: ident,
	}
}

type VarNode struct {
	Ident string
}

func CreateVar(id string) VarNode {
	return VarNode{
		Ident: id,
	}
}

type DiscardNode struct{}

func Discard() DiscardNode {
	return DiscardNode{}
}

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

func Break() BreakNode {
	return BreakNode{}
}

type ContinueNode struct {
	Label string
}

func Continue() ContinueNode {
	return ContinueNode{}
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

func BindArray(list []Node) BindingArrayNode {
	return BindingArrayNode{
		List: list,
	}
}

type BindingObjectNode struct {
	List map[string]Node
}

func BindObject(list map[string]Node) BindingObjectNode {
	return BindingObjectNode{
		List: list,
	}
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

type InNode struct {
	Left  Node
	Right Node
}

type InstanceOf struct {
	Left  Node
	Right Node
}
