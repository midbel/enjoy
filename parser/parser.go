package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/midbel/enjoy/ast"
	"github.com/midbel/enjoy/scanner"
	"github.com/midbel/enjoy/token"
)

type (
	prefixFunc func() (ast.Node, error)
	infixFunc  func(ast.Node) (ast.Node, error)
)

func ParseString(str string) (ast.Node, error) {
	return Parse(strings.NewReader(str))
}

func Parse(r io.Reader) (ast.Node, error) {
	return NewParser(r).Parse()
}

type Parser struct {
	prefix   map[rune]prefixFunc
	infix    map[rune]infixFunc
	keywords map[string]prefixFunc

	scan *scanner.Scanner
	curr token.Token
	peek token.Token

	allowDestructAssign int
}

func NewParser(r io.Reader) *Parser {
	p := Parser{
		scan:     scanner.Scan(r),
		prefix:   make(map[rune]prefixFunc),
		infix:    make(map[rune]infixFunc),
		keywords: make(map[string]prefixFunc),
	}

	p.registerPrefix(token.Number, p.parseNumber)
	p.registerPrefix(token.String, p.parseString)
	p.registerPrefix(token.Boolean, p.parseBool)
	p.registerPrefix(token.Ident, p.parseIdentifier)
	p.registerPrefix(token.Keyword, p.parseKeyword)
	p.registerPrefix(token.Template, p.parseTemplate)
	p.registerPrefix(token.Add, p.parseUnary)
	p.registerPrefix(token.Sub, p.parseUnary)
	p.registerPrefix(token.Not, p.parseUnary)
	p.registerPrefix(token.Bnot, p.parseUnary)
	p.registerPrefix(token.Increment, p.parseUnary)
	p.registerPrefix(token.Decrement, p.parseUnary)
	p.registerPrefix(token.Lparen, p.parseGroup)
	p.registerPrefix(token.Lbrace, p.parseBrace)
	p.registerPrefix(token.Lsquare, p.parseSquare)
	p.registerPrefix(token.Spread, p.parseSpread)

	p.registerInfix(token.Eq, p.parseBinary)
	p.registerInfix(token.Seq, p.parseBinary)
	p.registerInfix(token.Ne, p.parseBinary)
	p.registerInfix(token.Sne, p.parseBinary)
	p.registerInfix(token.Lt, p.parseBinary)
	p.registerInfix(token.Le, p.parseBinary)
	p.registerInfix(token.Gt, p.parseBinary)
	p.registerInfix(token.Ge, p.parseBinary)
	p.registerInfix(token.Add, p.parseBinary)
	p.registerInfix(token.Sub, p.parseBinary)
	p.registerInfix(token.Mul, p.parseBinary)
	p.registerInfix(token.Div, p.parseBinary)
	p.registerInfix(token.Pow, p.parseBinary)
	p.registerInfix(token.Mod, p.parseBinary)
	p.registerInfix(token.Nullish, p.parseBinary)
	p.registerInfix(token.And, p.parseBinary)
	p.registerInfix(token.Or, p.parseBinary)
	p.registerInfix(token.Lshift, p.parseBinary)
	p.registerInfix(token.Rshift, p.parseBinary)
	p.registerInfix(token.Band, p.parseBinary)
	p.registerInfix(token.Bor, p.parseBinary)
	p.registerInfix(token.Bxor, p.parseBinary)
	p.registerInfix(token.Assign, p.parseAssign)
	p.registerInfix(token.AddAssign, p.parseAssign)
	p.registerInfix(token.SubAssign, p.parseAssign)
	p.registerInfix(token.MulAssign, p.parseAssign)
	p.registerInfix(token.DivAssign, p.parseAssign)
	p.registerInfix(token.PowAssign, p.parseAssign)
	p.registerInfix(token.ModAssign, p.parseAssign)
	p.registerInfix(token.NullishAssign, p.parseAssign)
	p.registerInfix(token.AndAssign, p.parseAssign)
	p.registerInfix(token.OrAssign, p.parseAssign)
	p.registerInfix(token.LshiftAssign, p.parseAssign)
	p.registerInfix(token.RshiftAssign, p.parseAssign)
	p.registerInfix(token.BandAssign, p.parseAssign)
	p.registerInfix(token.BorAssign, p.parseAssign)
	p.registerInfix(token.BxorAssign, p.parseAssign)
	p.registerInfix(token.Question, p.parseTernary)
	p.registerInfix(token.Lparen, p.parseCall)
	p.registerInfix(token.Lsquare, p.parseIndex)
	p.registerInfix(token.Arrow, p.parseArrow)
	p.registerInfix(token.Dot, p.parseMember)
	p.registerInfix(token.Optional, p.parseMember)
	p.registerInfix(token.Keyword, p.parseOperatorKeyword)

	p.registerKeyword("let", p.parseLet)
	p.registerKeyword("const", p.parseConst)
	p.registerKeyword("if", p.parseIf)
	p.registerKeyword("else", p.parseElse)
	p.registerKeyword("switch", p.parseSwitch)
	p.registerKeyword("case", p.parseCase)
	p.registerKeyword("for", p.parseFor)
	p.registerKeyword("do", p.parseDo)
	p.registerKeyword("while", p.parseWhile)
	p.registerKeyword("break", p.parseBreak)
	p.registerKeyword("continue", p.parseContinue)
	p.registerKeyword("try", p.parseTry)
	p.registerKeyword("catch", p.parseCatch)
	p.registerKeyword("finally", p.parseFinally)
	p.registerKeyword("throw", p.parseThrow)
	p.registerKeyword("function", p.parseFunction)
	p.registerKeyword("return", p.parseReturn)
	p.registerKeyword("null", p.parseNull)
	p.registerKeyword("undefined", p.parseUndefined)
	p.registerKeyword("typeof", p.parseTypeOf)
	p.registerKeyword("export", p.parseExport)
	p.registerKeyword("import", p.parseImport)

	p.next()
	p.next()
	return &p
}

func (p *Parser) Parse() (ast.Node, error) {
	n, err := p.parse()
	if err != nil {
		p.reset()
	}
	return n, err
}

func (p *Parser) parse() (ast.Node, error) {
	var b ast.BlockNode
	for !p.done() {
		p.skip(token.Comment)
		n, err := p.parseNode(powLowest)
		if err != nil {
			return nil, err
		}
		b.Nodes = append(b.Nodes, n)
		if p.is(token.EOL) {
			p.next()
		}
		p.skip(token.Comment)
	}
	if len(b.Nodes) == 1 {
		return b.Nodes[0], nil
	}
	return b, nil
}

func (p *Parser) parseNode(pow int) (ast.Node, error) {
	fn, ok := p.prefix[p.curr.Type]
	if !ok {
		return nil, p.unexpected()
	}
	left, err := fn()
	if err != nil {
		return nil, err
	}
	for !p.done() && !p.eol() && pow < p.power() {
		fn, ok := p.infix[p.curr.Type]
		if !ok {
			return nil, p.unexpected()
		}
		left, err = fn(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *Parser) parseGroup() (ast.Node, error) {
	if err := p.expect(token.Lparen); err != nil {
		return nil, err
	}
	var seq ast.SeqNode
	for !p.done() && !p.is(token.Rparen) {
		n, err := p.parseNode(powComma)
		if err != nil {
			return nil, err
		}
		seq.Nodes = append(seq.Nodes, n)
		switch {
		case p.is(token.Comma):
			p.next()
			if p.is(token.Rparen) {
				return nil, p.unexpected()
			}
		case p.is(token.Rparen):
		default:
			return nil, p.unexpected()
		}
	}
	return seq, p.expect(token.Rparen)
}

func (p *Parser) parseBinding(let bool) (ast.Node, bool, error) {
	p.enableDestructuring()
	defer p.disableDestructuring()

	p.next()
	node, err := p.parseNode(powAssign)
	if err != nil {
		return nil, false, err
	}
	if let {
		if _, ok := node.(ast.VarNode); p.is(token.EOL) {
			if !ok {
				return nil, false, p.unexpected()
			}
			return node, true, p.expect(token.EOL)
		}
	}
	return node, false, p.expect(token.Assign)
}

func (p *Parser) parseLet() (ast.Node, error) {
	node, done, err := p.parseBinding(true)
	if err != nil || done {
		return makeLet(node), err
	}
	bind := makeLet(node)
	bind.Expr, err = p.parseNode(powLowest)
	return bind, err
}

func (p *Parser) parseConst() (ast.Node, error) {
	node, _, err := p.parseBinding(false)
	if err != nil {
		return nil, err
	}
	bind := makeConst(node)
	bind.Expr, err = p.parseNode(powLowest)
	return bind, err
}

func (p *Parser) parseBrace() (ast.Node, error) {
	if p.isDestructuringAllowed() {
		return p.parseObjectBinding()
	}
	return p.parseObject()
}

func (p *Parser) parseObject() (ast.Node, error) {
	if err := p.expect(token.Lbrace); err != nil {
		return nil, err
	}
	list := make(map[string]ast.Node)
	for !p.done() && !p.is(token.Rbrace) {
		if !p.is(token.Ident) && !p.is(token.String) && !p.is(token.Number) && !p.is(token.Boolean) {
			return nil, p.unexpected()
		}
		ident := p.curr.Literal
		p.next()
		if p.is(token.Comma) || p.is(token.Rbrace) {
			list[ident] = ast.CreateVar(ident)
			if p.is(token.Comma) {
				p.next()
			}
			continue
		}
		if err := p.expect(token.Colon); err != nil {
			return nil, err
		}
		node, err := p.parseNode(powComma)
		if err != nil {
			return nil, err
		}
		list[ident] = node
		switch {
		case p.is(token.Comma):
			p.next()
		case p.is(token.Rbrace):
		default:
			return nil, p.unexpected()
		}
	}
	return ast.Object(list), p.expect(token.Rbrace)
}

func (p *Parser) parseObjectBinding() (ast.Node, error) {
	if err := p.expect(token.Lbrace); err != nil {
		return nil, err
	}
	list := make(map[string]ast.Node)
	for !p.done() && !p.is(token.Rbrace) {
		if p.is(token.Spread) {
			p.next()
			if !p.is(token.Ident) {
				return nil, p.unexpected()
			}
			list[p.curr.Literal] = makeSpreadWithVar(p.curr.Literal)
			p.next()
			if !p.is(token.Rbrace) {
				return nil, p.unexpected()
			}
			continue
		}
		if !p.is(token.Ident) && !p.is(token.String) && !p.is(token.Number) && !p.is(token.Boolean) {
			return nil, p.unexpected()
		}
		var (
			ident = p.curr.Literal
			value ast.AssignNode
			err   error
		)
		p.next()
		if p.is(token.Colon) {
			p.next()
			switch {
			case p.is(token.Lbrace):
				value.Ident, err = p.parseObjectBinding()
				if err != nil {
					return nil, err
				}
			case p.is(token.Ident):
				value.Ident = ast.CreateVar(p.curr.Literal)
				p.next()
			default:
				return nil, p.unexpected()
			}
		} else {
			value.Ident = ast.CreateVar(ident)
		}
		if p.is(token.Assign) {
			p.next()
			value.Expr, err = p.parseNode(powComma)
			if err != nil {
				return nil, err
			}
		}
		list[ident] = value
		switch {
		case p.is(token.Comma):
			p.next()
		case p.is(token.Rbrace):
		default:
			return nil, p.unexpected()
		}
	}
	return ast.BindObject(list), p.expect(token.Rbrace)
}

func (p *Parser) parseSquare() (ast.Node, error) {
	if p.isDestructuringAllowed() {
		return p.parseArrayBinding()
	}
	return p.parseArray()
}

func (p *Parser) parseArray() (ast.Node, error) {
	if err := p.expect(token.Lsquare); err != nil {
		return nil, err
	}
	var list []ast.Node
	for !p.done() && !p.is(token.Rsquare) {
		if p.is(token.Comma) {
			p.next()
			list = append(list, ast.Discard())
			continue
		}
		n, err := p.parseNode(powComma)
		if err != nil {
			return nil, err
		}
		list = append(list, n)
		switch {
		case p.is(token.Comma):
			p.next()
		case p.is(token.Rsquare):
		default:
			return nil, p.unexpected()
		}
	}
	return ast.Array(list), p.expect(token.Rsquare)
}

func (p *Parser) parseArrayBinding() (ast.Node, error) {
	if err := p.expect(token.Lsquare); err != nil {
		return nil, err
	}
	var (
		list []ast.Node
		err  error
	)
	for !p.done() && !p.is(token.Rsquare) {
		if p.is(token.Comma) {
			p.next()
			list = append(list, ast.Discard())
			continue
		}
		var node ast.Node
		switch {
		case p.is(token.Comma):
			node = ast.Discard()
			p.next()
		case p.is(token.Ident):
			node = ast.CreateVar(p.curr.Literal)
			p.next()
		case p.is(token.Lbrace):
			node, err = p.parseObjectBinding()
		case p.is(token.Spread):
			p.next()
			if p.is(token.Ident) {
				node = makeSpreadWithVar(p.curr.Literal)
				p.next()
			} else {
				node, err = p.parseArrayBinding()
				if err == nil {
					node = makeSpreadFrom(node)
				}
			}
		default:
			return nil, p.unexpected()
		}
		if err != nil {
			return nil, err
		}
		if _, ok := node.(ast.SpreadNode); !ok && p.is(token.Assign) {
			ass := makeAssignNode(node)
			p.next()
			ass.Expr, err = p.parseNode(powComma)
			node = ass
		}
		list = append(list, node)
		switch {
		case p.is(token.Comma):
			p.next()
			if _, ok := node.(ast.SpreadNode); ok && p.is(token.Rsquare) {
				return nil, p.unexpected()
			}
		case p.is(token.Rsquare):
		default:
			return nil, p.unexpected()
		}
	}
	return ast.BindArray(list), p.expect(token.Rsquare)
}

func (p *Parser) parseNull() (ast.Node, error) {
	defer p.next()
	return ast.Null(), nil
}

func (p *Parser) parseUndefined() (ast.Node, error) {
	defer p.next()
	return ast.Undefined(), nil
}

func (p *Parser) parseOperatorKeyword(left ast.Node) (ast.Node, error) {
	switch p.curr.Literal {
	case "in":
		return p.parseIn(left)
	case "instanceof":
		return p.parseInstanceOf(left)
	default:
		return nil, p.unexpected()
	}
}

func (p *Parser) parseIn(left ast.Node) (ast.Node, error) {
	if err := p.expect(token.Keyword); err != nil {
		return nil, err
	}
	var (
		err  error
		node ast.InNode
	)
	node = ast.InNode{
		Left: left,
	}
	node.Right, err = p.parseNode(powLowest)
	return node, err
}

func (p *Parser) parseInstanceOf(left ast.Node) (ast.Node, error) {
	if err := p.expect(token.Keyword); err != nil {
		return nil, err
	}
	var (
		err  error
		node ast.InstanceOfNode
	)
	node = ast.InstanceOfNode{
		Left: left,
	}
	node.Right, err = p.parseNode(powLowest)
	return node, err
}

func (p *Parser) parseTypeOf() (ast.Node, error) {
	p.next()
	var (
		node ast.TypeofNode
		err  error
	)
	node.Node, err = p.parseNode(powLowest)
	return node, err
}

func (p *Parser) parseExport() (ast.Node, error) {
	parseFrom := func() (string, error) {
		if err := p.expectKW("from"); err != nil {
			return "", err
		}
		if !p.is(token.String) {
			return "", p.unexpected()
		}
		defer p.next()
		return p.curr.Literal, nil
	}

	parseList := func() (ast.Node, error) {
		p.next()
		var list []ast.Node
		for !p.done() && !p.is(token.Rbrace) {
			switch {
			case p.is(token.Ident):
			case p.is(token.String):
			case p.is(token.Keyword) && p.curr.Literal == "default":
			default:
				return nil, p.unexpected()
			}
			var (
				ident = p.curr.Literal
				alias string
			)
			p.next()
			if p.is(token.Keyword) && p.curr.Literal == "as" {
				p.next()
				if !p.is(token.Ident) && !p.is(token.String) {
					return nil, p.unexpected()
				}
				alias = p.curr.Literal
				p.next()
			}
			list = append(list, ast.Alias(ident, alias))
			switch {
			case p.is(token.Comma):
				p.next()
				if p.is(token.Rbrace) {
					return nil, p.unexpected()
				}
			case p.is(token.Rbrace):
			default:
				return nil, p.unexpected()
			}
		}
		return ast.Sequence(list), p.expect(token.Rbrace)
	}

	p.next()
	switch {
	case p.is(token.Keyword):
		var node ast.ExportNode
		if p.curr.Literal == "default" {
			node.Default = true
			p.next()
		}
		n, err := p.parseNode(powLowest)
		if err != nil {
			return nil, err
		}
		node.Node = n
		return node, nil
	case p.is(token.Lbrace):
		node, err := parseList()
		if err != nil {
			return nil, err
		}
		if p.is(token.Keyword) && p.curr.Literal == "from" {
			file, err := parseFrom()
			return ast.ExportFrom(node, file), err
		}
		return ast.Export(node), nil
	case p.is(token.Mul):
		p.next()
		var ident ast.Node
		if p.is(token.Keyword) && p.curr.Literal == "as" {
			p.next()
			ident = ast.CreateVar(p.curr.Literal)
			p.next()
		}
		file, err := parseFrom()
		if err != nil {
			return nil, err
		}
		return ast.ExportFrom(ident, file), nil
	default:
		return nil, p.unexpected()
	}
}

func (p *Parser) parseImport() (ast.Node, error) {
	parseFrom := func() (string, error) {
		p.next()
		if err := p.expectKW("from"); err != nil {
			return "", err
		}
		if !p.is(token.String) {
			return "", p.unexpected()
		}
		defer p.next()
		return p.curr.Literal, nil
	}

	parseStar := func(prev ast.Node) (ast.Node, error) {
		p.next()
		err := p.expectKW("as")
		if err != nil {
			return nil, err
		}
		var (
			ident = ast.CreateVar(p.curr.Literal)
			file  string
		)
		if file, err = parseFrom(); err != nil {
			return nil, err
		}
		n := ast.Import(ident, file)
		n.Default = prev
		return n, nil
	}

	parseList := func(prev ast.Node) (ast.Node, error) {
		p.next()
		var (
			list    []ast.Node
			seendef bool
		)
		for !p.done() && !p.is(token.Rbrace) {
			switch {
			case p.is(token.Ident):
			case p.is(token.String):
			case p.is(token.Keyword) && p.curr.Literal == "default":
				if seendef {
					return nil, p.unexpected()
				}
				seendef = true
			default:
				return nil, p.unexpected()
			}
			var (
				ident = p.curr.Literal
				alias string
			)
			p.next()
			if p.is(token.Keyword) && p.curr.Literal == "as" {
				p.next()
				if !p.is(token.Ident) && !p.is(token.String) {
					return nil, p.unexpected()
				}
				alias = p.curr.Literal
				p.next()
			}
			if ident == "default" && alias == "" {
				return nil, fmt.Errorf("missing alias")
			}
			list = append(list, ast.Alias(ident, alias))
			switch {
			case p.is(token.Comma):
				p.next()
				if p.is(token.Rbrace) {
					return nil, p.unexpected()
				}
			case p.is(token.Rbrace):
			default:
				return nil, p.unexpected()
			}
		}
		if !p.is(token.Rbrace) {
			return nil, p.unexpected()
		}
		file, err := parseFrom()
		if err != nil {
			return nil, err
		}
		n := ast.Import(ast.Sequence(list), file)
		n.Default = prev
		return n, nil
	}

	p.next()
	switch {
	case p.is(token.String):
		defer p.next()
		return ast.Import(nil, p.curr.Literal), nil
	case p.is(token.Mul):
		return parseStar(nil)
	case p.is(token.Ident):
		ident := ast.CreateVar(p.curr.Literal)
		if file, err := parseFrom(); err == nil {
			return ast.Import(ident, file), nil
		}
		if err := p.expect(token.Comma); err != nil {
			return nil, err
		}
		if p.is(token.Mul) {
			return parseStar(ident)
		}
		if p.is(token.Lbrace) {
			return parseList(ident)
		}
		return nil, p.unexpected()
	case p.is(token.Lbrace):
		return parseList(nil)
	default:
		return nil, p.unexpected()
	}
}

func (p *Parser) parseIf() (ast.Node, error) {
	p.next()
	var (
		node ast.IfNode
		err  error
	)
	node.Cdt, err = p.parseCondition()
	if err != nil {
		return nil, err
	}
	node.Csq, err = p.parseBody()
	if err != nil {
		return nil, err
	}
	if p.is(token.Keyword) && p.curr.Literal == "else" {
		node.Alt, err = p.parseKeyword()
	}
	return node, err
}

func (p *Parser) parseElse() (ast.Node, error) {
	p.next()
	if p.is(token.Keyword) && p.curr.Literal == "if" {
		return p.parseKeyword()
	}
	return p.parseBody()
}

func (p *Parser) parseSwitch() (ast.Node, error) {
	p.next()
	var (
		node ast.SwitchNode
		err  error
	)
	node.Cdt, err = p.parseCondition()
	if err != nil {
		return nil, err
	}
	if err = p.expect(token.Lbrace); err != nil {
		return nil, err
	}
	for !p.done() && !p.is(token.Rbrace) {
		if !p.is(token.Keyword) {
			return nil, p.unexpected()
		}
		if p.curr.Literal == "default" {
			break
		}
		if p.curr.Literal != "case" {
			return nil, p.unexpected()
		}
		b, err := p.parseCase()
		if err != nil {
			return nil, err
		}
		node.Cases = append(node.Cases, b)
	}
	if p.is(token.Keyword) && p.curr.Literal == "default" {
		node.Default, err = p.parseDefault()
		return node, err
	}
	return node, p.expect(token.Rbrace)
}

func (p *Parser) parseDefault() (ast.Node, error) {
	p.next()
	if err := p.expect(token.Colon); err != nil {
		return nil, err
	}
	var nodes []ast.Node
	p.skip(token.EOL)
	for !p.done() && !p.is(token.Rbrace) {
		n, err := p.parseNode(powLowest)
		if err != nil {
			return nil, err
		}
		p.skip(token.EOL)
		nodes = append(nodes, n)
	}
	return blockOrNode(nodes), p.expect(token.Rbrace)
}

func (p *Parser) parseCase() (ast.Node, error) {
	p.next()
	var (
		clause ast.CaseNode
		err    error
	)
	clause.Predicate, err = p.parseNode(powAssign)
	if err != nil {
		return nil, err
	}
	if err = p.expect(token.Colon); err != nil {
		return nil, err
	}
	var nodes []ast.Node
	p.skip(token.EOL)
	for !p.done() && !p.is(token.Rbrace) {
		if p.is(token.Keyword) && (p.curr.Literal == "case" || p.curr.Literal == "default") {
			break
		}
		n, err := p.parseNode(powLowest)
		if err != nil {
			return nil, err
		}
		p.skip(token.EOL)
		nodes = append(nodes, n)
	}
	clause.Body = blockOrNode(nodes)
	return clause, nil
}

func (p *Parser) parseForeach() (ast.Node, bool, error) {
	n, err := p.parseNode(powLowest)
	if err == nil {
		return n, false, err
	}
	if !p.is(token.Assign) && !p.is(token.Keyword) {
		return nil, false, err
	}
	var (
		loop ast.LoopNode
		kw   = p.curr.Literal
	)
	p.next()
	it, err := p.parseNode(powLowest)
	if err != nil {
		return nil, false, err
	}
	switch kw {
	default:
		return nil, false, p.unexpected()
	case "of":
		loop.Iter = makeIterOf(n, it)
	case "in":
		loop.Iter = makeIterIn(n, it)
	}
	if err := p.expect(token.Rparen); err != nil {
		return nil, false, err
	}
	loop.Body, err = p.parseBody()
	return loop, true, err
}

func (p *Parser) parseFor() (ast.Node, error) {
	p.scan.ToggleKeepEOL()
	p.next()
	var (
		node ast.ForNode
		err  error
	)
	if err := p.expect(token.Lparen); err != nil {
		return nil, err
	}
	if !p.is(token.EOL) {
		n, done, err := p.parseForeach()
		if err != nil || done {
			p.scan.ToggleKeepEOL()
			return n, err
		}
		node.Init = n
	}
	if err = p.expect(token.EOL); err != nil {
		return nil, err
	}
	if !p.is(token.EOL) {
		node.Cdt, err = p.parseNode(powLowest)
		if err != nil {
			return nil, err
		}
	}
	if err = p.expect(token.EOL); err != nil {
		return nil, err
	}
	if !p.is(token.Rparen) {
		node.Incr, err = p.parseNode(powLowest)
		if err != nil {
			return nil, err
		}
	}
	if err := p.expect(token.Rparen); err != nil {
		return nil, err
	}
	p.scan.ToggleKeepEOL()
	node.Body, err = p.parseBody()
	return node, err
}

func (p *Parser) parseDo() (ast.Node, error) {
	p.next()
	var (
		node ast.DoNode
		err  error
	)
	node.Body, err = p.parseBody()
	if err != nil {
		return nil, err
	}
	if !p.is(token.Keyword) && p.curr.Literal != "while" {
		return nil, p.unexpected()
	}
	p.next()
	node.Cdt, err = p.parseCondition()
	return node, err
}

func (p *Parser) parseWhile() (ast.Node, error) {
	p.next()
	var (
		node ast.WhileNode
		err  error
	)
	node.Cdt, err = p.parseCondition()
	if err != nil {
		return nil, err
	}
	node.Body, err = p.parseBody()
	return node, err
}

func (p *Parser) parseBreak() (ast.Node, error) {
	p.next()
	var label string
	if p.is(token.Ident) {
		label = p.curr.Literal
		p.next()
	}
	return ast.Break(label), nil
}

func (p *Parser) parseContinue() (ast.Node, error) {
	p.next()
	var label string
	if p.is(token.Ident) {
		label = p.curr.Literal
		p.next()
	}
	return ast.Continue(label), nil
}

func (p *Parser) parseTry() (ast.Node, error) {
	p.next()
	var (
		node ast.TryNode
		err  error
	)
	node.Try, err = p.parseBody()
	if err != nil {
		return nil, err
	}
	if p.is(token.Keyword) && p.curr.Literal == "catch" {
		node.Catch, err = p.parseKeyword()
		if err != nil {
			return nil, err
		}
	}
	if p.is(token.Keyword) && p.curr.Literal == "finally" {
		node.Finally, err = p.parseKeyword()
	}
	return node, err
}

func (p *Parser) parseCatch() (ast.Node, error) {
	p.next()
	var (
		catch ast.CatchNode
		err   error
	)
	catch.Ident, err = p.parseCondition()
	if err != nil {
		return nil, err
	}
	catch.Body, err = p.parseBody()
	return catch, err
}

func (p *Parser) parseFinally() (ast.Node, error) {
	p.next()
	return p.parseBody()
}

func (p *Parser) parseThrow() (ast.Node, error) {
	p.next()
	var (
		throw ast.ThrowNode
		err   error
	)
	throw.Node, err = p.parseNode(powLowest)
	return throw, err
}

func (p *Parser) parseFunction() (ast.Node, error) {
	p.next()
	var (
		fn  ast.FuncNode
		err error
	)
	if p.is(token.Ident) {
		fn.Ident = p.curr.Literal
		p.next()
	}
	if fn.Args, err = p.parseArgs(); err != nil {
		return nil, err
	}
	fn.Body, err = p.parseBody()
	return fn, err
}

func (p *Parser) parseArgs() (ast.Node, error) {
	p.enableDestructuring()
	defer p.disableDestructuring()

	if err := p.expect(token.Lparen); err != nil {
		return nil, err
	}

	var seq ast.SeqNode
	for !p.done() && !p.is(token.Rparen) {
		n, err := p.parseNode(powComma)
		if err != nil {
			return nil, err
		}
		seq.Nodes = append(seq.Nodes, n)
		if _, ok := n.(ast.SpreadNode); ok && !p.is(token.Rparen) {
			return nil, p.unexpected()
		}
		switch {
		case p.is(token.Comma):
			p.next()
			if p.is(token.Rparen) {
				return nil, p.unexpected()
			}
		case p.is(token.Rparen):
		default:
			return nil, p.unexpected()
		}
	}
	return seq, p.expect(token.Rparen)
}

func (p *Parser) parseReturn() (ast.Node, error) {
	p.next()
	var (
		ret ast.ReturnNode
		err error
	)
	ret.Node, err = p.parseNode(powLowest)
	return ret, err
}

func (p *Parser) parseBinary(left ast.Node) (ast.Node, error) {
	bin := ast.BinaryNode{
		Op:   p.curr.Type,
		Left: left,
	}
	p.next()
	right, err := p.parseNode(powers.Get(bin.Op))
	if err != nil {
		return nil, err
	}
	bin.Right = right
	return bin, nil
}

func (p *Parser) parseAssign(left ast.Node) (ast.Node, error) {
	var (
		node = makeAssignNode(left)
		op   = p.curr.Type
	)
	p.next()

	if p.isDestructuringAllowed() {
		p.disableDestructuring()
		defer p.enableDestructuring()
	}

	expr, err := p.parseNode(powAssign)
	if err != nil {
		return nil, err
	}
	op, err = token.ConvertAssignToken(op)
	if err != nil {
		return nil, err
	}
	if op != token.Assign {
		expr = ast.BinaryNode{
			Op:    op,
			Left:  left,
			Right: expr,
		}
	}
	node.Expr = expr
	return node, nil
}

func (p *Parser) parseTernary(left ast.Node) (ast.Node, error) {
	p.next()
	node := ast.IfNode{
		Cdt: left,
	}
	csq, err := p.parseNode(powAssign)
	if err != nil {
		return nil, err
	}
	node.Csq = csq
	if err = p.expect(token.Colon); err != nil {
		return nil, err
	}
	node.Alt, err = p.parseNode(powLowest)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (p *Parser) parseIndex(left ast.Node) (ast.Node, error) {
	if err := p.expect(token.Lsquare); err != nil {
		return nil, err
	}
	var (
		node ast.IndexNode
		err  error
	)
	node.Expr = left
	node.Index, err = p.parseNode(powLowest)
	if err != nil {
		return nil, err
	}
	return node, p.expect(token.Rsquare)
}

func (p *Parser) parseCall(left ast.Node) (ast.Node, error) {
	call := makeCall(left)
	args, err := p.parseGroup()
	if err != nil {
		return nil, err
	}
	call.Args = args
	return call, nil
}

func (p *Parser) parseArrow(left ast.Node) (ast.Node, error) {
	var (
		fn  ast.ArrowNode
		err error
	)
	fn.Args = left
	p.next()
	switch {
	case p.is(token.Lparen):
		p.next()
		fn.Body, err = p.parseObject()
		if err == nil {
			err = p.expect(token.Rparen)
		}
	case p.is(token.Lbrace):
		fn.Body, err = p.parseBody()
	default:
		fn.Body, err = p.parseNode(powLowest)
	}
	return fn, err
}

func (p *Parser) parseMember(left ast.Node) (ast.Node, error) {
	p.next()
	node := ast.MemberNode{
		Curr: left,
	}
	next, err := p.parseNode(powObject)
	if err != nil {
		return nil, err
	}
	node.Next = next
	return node, nil
}

func (p *Parser) parseUnary() (ast.Node, error) {
	node := ast.UnaryNode{
		Op: p.curr.Type,
	}
	p.next()
	expr, err := p.parseNode(powUnary)
	if err != nil {
		return nil, err
	}
	node.Expr = expr
	return node, err
}

func (p *Parser) parseNumber() (ast.Node, error) {
	defer p.next()
	n, err := strconv.ParseFloat(p.curr.Literal, 64)
	if err != nil {
		return nil, err
	}
	return ast.CreateValue(n), nil
}

func (p *Parser) parseString() (ast.Node, error) {
	defer p.next()
	return ast.CreateValue(p.curr.Literal), nil
}

func (p *Parser) parseBool() (ast.Node, error) {
	defer p.next()
	n, err := strconv.ParseBool(p.curr.Literal)
	if err != nil {
		return nil, err
	}
	return ast.CreateValue(n), nil
}

func (p *Parser) parseIdentifier() (ast.Node, error) {
	node := ast.CreateVar(p.curr.Literal)
	p.next()
	if p.is(token.Colon) {
		p.next()
		return ast.Label(node.Ident), nil
	}
	return node, nil
}

func (p *Parser) parseKeyword() (ast.Node, error) {
	parse, ok := p.keywords[p.curr.Literal]
	if !ok {
		return nil, p.unexpected()
	}
	return parse()
}

func (p *Parser) parseTemplate() (ast.Node, error) {
	if err := p.expect(token.Template); err != nil {
		return nil, err
	}
	var node ast.TemplateNode
	for !p.done() && !p.is(token.Template) {
		if p.is(token.String) {
			node.Nodes = append(node.Nodes, ast.CreateValue(p.curr.Literal))
			p.next()
			continue
		}
		if !p.is(token.BegSub) {
			return nil, p.unexpected()
		}
		p.next()
		n, err := p.parseNode(powLowest)
		if err != nil {
			return nil, err
		}
		node.Nodes = append(node.Nodes, n)
		if err := p.expect(token.EndSub); err != nil {
			return nil, err
		}
	}
	return node, p.expect(token.Template)
}

func (p *Parser) parseSpread() (ast.Node, error) {
	p.next()
	node, err := p.parseNode(powAssign)
	return makeSpreadFrom(node), err
}

func (p *Parser) parseBody() (ast.Node, error) {
	if err := p.expect(token.Lbrace); err != nil {
		return nil, err
	}
	p.skip(token.EOL)
	var b ast.BlockNode
	for !p.done() && !p.is(token.Rbrace) {
		n, err := p.parseNode(powLowest)
		if err != nil {
			return nil, err
		}
		p.skip(token.EOL)
		b.Nodes = append(b.Nodes, n)
	}
	if err := p.expect(token.Rbrace); err != nil {
		return nil, err
	}
	if len(b.Nodes) == 1 {
		return b.Nodes[0], nil
	}
	return b, nil
}

func (p *Parser) parseCondition() (ast.Node, error) {
	if err := p.expect(token.Lparen); err != nil {
		return nil, err
	}
	expr, err := p.parseNode(powLowest)
	if err != nil {
		return nil, err
	}
	return expr, p.expect(token.Rparen)
}

func (p *Parser) reset() {
	for !p.is(token.EOL) && !p.done() {
		p.next()
	}
	p.scan.Reset()
	p.resetDestructuring()
}

func (p *Parser) isDestructuringAllowed() bool {
	return p.allowDestructAssign > 0
}

func (p *Parser) enableDestructuring() {
	p.allowDestructAssign++
}

func (p *Parser) disableDestructuring() {
	p.allowDestructAssign--
}

func (p *Parser) resetDestructuring() {
	p.allowDestructAssign = 0
}

func (p *Parser) registerPrefix(kind rune, fn prefixFunc) {
	p.prefix[kind] = fn
}

func (p *Parser) registerInfix(kind rune, fn infixFunc) {
	p.infix[kind] = fn
}

func (p *Parser) registerKeyword(kw string, fn prefixFunc) {
	p.keywords[kw] = fn
}

func (p *Parser) skip(kind rune) {
	for p.is(kind) {
		p.next()
	}
}

func (p *Parser) power() int {
	return powers.Get(p.curr.Type)
}

func (p *Parser) expectKW(kw string) error {
	if !p.is(token.Keyword) {
		return p.unexpected()
	}
	if p.curr.Literal == kw {
		p.next()
		return nil
	}
	return p.unexpected()
}

func (p *Parser) expect(kind rune) error {
	if p.is(kind) {
		p.next()
		return nil
	}
	return p.unexpected()
}

func (p *Parser) unexpected() error {
	pos := p.curr.Position
	return fmt.Errorf("(%d:%d) unexpected token %s", pos.Line, pos.Column, p.curr)
}

func (p *Parser) done() bool {
	return p.is(token.EOF)
}

func (p *Parser) eol() bool {
	return p.is(token.EOL)
}

func (p *Parser) is(kind rune) bool {
	return p.curr.Type == kind
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}
