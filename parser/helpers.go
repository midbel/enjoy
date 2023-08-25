package parser

import (
	"github.com/midbel/enjoy/ast"
)

func makeLet(ident ast.Node) ast.LetNode {
	return ast.LetNode{
		Ident: ident,
	}
}

func makeConst(ident ast.Node) ast.ConstNode {
	return ast.ConstNode{
		Ident: ident,
	}
}

func makeSpreadWithVar(ident string) ast.SpreadNode {
	return makeSpreadFrom(ast.CreateVar(ident))
}

func makeSpreadFrom(node ast.Node) ast.SpreadNode {
	return ast.SpreadNode{
		Node: node,
	}
}

func makeAssignNode(node ast.Node) ast.AssignNode {
	return ast.AssignNode{
		Ident: node,
	}
}

func makeIterOf(id, it ast.Node) ast.IterOfNode {
	return ast.IterOfNode{
		Ident: id,
		Iter:  it,
	}
}

func makeIterIn(id, it ast.Node) ast.IterInNode {
	return ast.IterInNode{
		Ident: id,
		Iter:  it,
	}
}

func makeCall(id ast.Node) ast.CallNode {
	return ast.CallNode{
		Ident: id,
	}
}
