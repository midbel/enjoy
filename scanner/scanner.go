package scanner

import (
	"bytes"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/midbel/enjoy/token"
)

type Scanner struct {
	input []byte
	cursor
	old cursor

	str        bytes.Buffer
	mode       scanMode
	keepAllEol bool
}

func Scan(r io.Reader) *Scanner {
	buf, _ := io.ReadAll(r)
	buf, _ = bytes.CutPrefix(buf, []byte{0xef, 0xbb, 0xbf})
	s := Scanner{
		input: buf,
		mode:  modeDefault,
	}
	s.cursor.Line = 1
	s.read()
	s.skip(isBlank)
	return &s
}

func (s *Scanner) Reset() {
	s.mode = modeDefault
	s.keepAllEol = false
}

func (s *Scanner) ToggleKeepEOL() {
	s.keepAllEol = !s.keepAllEol
}

func (s *Scanner) Scan() token.Token {
	defer s.reset()

	switch s.mode {
	case modeTpl:
		return s.scanTemplate()
	case modeSub:
		return s.scanSubstitution()
	default:
		return s.scanDefault()
	}
}

func (s *Scanner) prepare() token.Token {
	var tok token.Token
	tok.Offset = s.curr
	tok.Position = s.cursor.Position
	return tok
}

func (s *Scanner) scanTemplate() token.Token {
	if isSubstitution(s.char, s.peek()) {
		return s.toggleSubstitution()
	}
	tok := s.prepare()
	if isTemplate(s.char) {
		s.toggleTemplate(&tok)
		return tok
	}
	for !s.done() && !isSubstitution(s.char, s.peek()) && !isTemplate(s.char) {
		s.write()
		s.read()
	}
	tok.Type = token.String
	tok.Literal = s.literal()
	return tok
}

func (s *Scanner) scanSubstitution() token.Token {
	if s.char == rbrace {
		return s.toggleSubstitution()
	}
	return s.scanDefault()
}

func (s *Scanner) scanDefault() token.Token {
	tok := s.prepare()

	if isEOL(s.char) {
		s.scanEOL(&tok)
		return tok
	}
	s.skip(isBlank)
	if s.done() {
		tok.Type = token.EOF
		return tok
	}

	switch {
	case isComment(s.char, s.peek()):
		s.scanComment(&tok)
	case isQuote(s.char):
		s.scanString(&tok)
	case isLetter(s.char):
		s.scanIdent(&tok)
	case isDigit(s.char):
		s.scanNumber(&tok)
	case isEOL(s.char):
		s.scanEOL(&tok)
	case isTemplate(s.char):
		s.toggleTemplate(&tok)
	default:
		s.scanPunct(&tok)
	}
	if token.CanSkipBlanks(tok.Type) {
		s.skip(isBlank)
	}
	s.discardNL()
	return tok
}

func (s *Scanner) discardNL() {
	if !isNL(s.char) {
		return
	}
	s.save()
	s.skip(isBlank)
	switch s.char {
	case rparen, rbrace, rsquare:
	default:
		s.restore()
	}
}

func (s *Scanner) toggleSubstitution() token.Token {
	tok := s.prepare()
	if s.mode == modeSub {
		tok.Type = token.EndSub
		s.mode = modeTpl
		s.read()
	} else {
		tok.Type = token.BegSub
		s.mode = modeSub
		s.read()
		s.read()
	}
	return tok
}

func (s *Scanner) toggleTemplate(tok *token.Token) {
	tok.Type = token.Template
	if s.mode == modeDefault {
		s.mode = modeTpl
	} else {
		s.mode = modeDefault
	}
	s.read()
}

func (s *Scanner) scanEOL(tok *token.Token) {
	tok.Type = token.EOL
	if s.char == semicolon && s.keepAllEol {
		s.read()
		return
	}
	s.skip(isEOL)
}

func (s *Scanner) scanComment(tok *token.Token) {
	s.read()
	s.read()
	s.skip(isSpace)
	for !s.done() && !isNL(s.char) {
		s.write()
		s.read()
	}
	s.skip(isBlank)
	tok.Type = token.Comment
	tok.Literal = s.literal()
}

func (s *Scanner) scanString(tok *token.Token) {
	quote := s.char
	s.read()

	tok.Type = token.String
	for !s.done() && s.char != quote {
		if s.char == backslash {
			s.read()
			char, ok := escapes[s.char]
			if !ok {
				tok.Type = token.Invalid
				s.writeRune(backslash)
				s.writeRune(char)
				continue
			}
			if s.char == 'x' {
				s.read()
				char = s.runeFromRunes(2)
			} else if s.char == 'u' {
				s.read()
				char = s.runeFromRunes(4)
			}
			s.writeRune(char)
			if char == utf8.RuneError {
				tok.Type = token.Invalid
			}
			continue
		}
		s.write()
		s.read()
	}
	if s.char != quote {
		tok.Type = token.Invalid
	} else {
		s.read()
	}
	tok.Literal = s.literal()
}

func (s *Scanner) scanDigits(tok *token.Token, accept func(rune) bool) {
	s.write()
	s.read()
	for !s.done() && accept(s.char) {
		s.write()
		s.read()
		if s.char == underscore {
			s.read()
			if !isDigit(s.char) {
				tok.Type = token.Invalid
			}
		}
	}
	// if !accept(s.char) {
	// 	s.unread()
	// }
	tok.Literal = s.literal()
}

func (s *Scanner) scanNumber(tok *token.Token) {
	tok.Type = token.Number
	if k := s.peek(); s.char == '0' && k == 'b' || k == 'x' || k == 'o' {
		s.write()
		s.read()
		switch s.char {
		case 'o':
			s.scanDigits(tok, isOctal)
		case 'b':
			s.scanDigits(tok, isBin)
		case 'x':
			s.scanDigits(tok, isHex)
		}
		return
	}
	var zeros int
	if s.char == '0' {
		s.write()
		s.read()
		for !s.done() && s.char == '0' {
			s.write()
			s.read()
			zeros++
		}
		if zeros > 1 {
			tok.Type = token.Invalid
		}
		if !isDigit(s.char) && s.char != dot {
			tok.Literal = s.literal()
			return
		}
	}
	if s.char == underscore {
		s.read()
		tok.Type = token.Invalid
	}
	if s.char != dot {
		s.scanDigits(tok, isDigit)
	}
	if s.char != dot {
		return
	}
	s.write()
	s.read()
	if s.char == underscore {
		s.read()
		tok.Type = token.Invalid
	}
	s.scanDigits(tok, isDigit)
}

func (s *Scanner) scanIdent(tok *token.Token) {
	tok.Type = token.Ident
	for !s.done() && isAlpha(s.char) {
		s.write()
		s.read()
	}
	tok.Literal = s.literal()
	if token.IsKeyword(tok.Literal) {
		tok.Type = token.Keyword
	}
	if tok.Literal == "true" || tok.Literal == "false" {
		tok.Type = token.Boolean
	}
}

func (s *Scanner) scanPunct(tok *token.Token) {
	switch s.char {
	case dot:
		tok.Type = token.Dot
		s.save()
		s.read()
		if s.char != dot {
			s.restore()
			break
		}
		s.read()
		if s.char != dot {
			s.restore()
			break
		}
		tok.Type = token.Spread
	case comma:
		tok.Type = token.Comma
	case colon:
		tok.Type = token.Colon
	case lbrace:
		tok.Type = token.Lbrace
	case rbrace:
		tok.Type = token.Rbrace
	case lparen:
		tok.Type = token.Lparen
	case rparen:
		tok.Type = token.Rparen
	case lsquare:
		tok.Type = token.Lsquare
	case rsquare:
		tok.Type = token.Rsquare
	case plus:
		tok.Type = token.Add
		if s.peek() == equal {
			s.read()
			tok.Type = token.AddAssign
		} else if s.peek() == plus {
			s.read()
			tok.Type = token.Increment
		}
	case minus:
		tok.Type = token.Sub
		if s.peek() == equal {
			s.read()
			tok.Type = token.SubAssign
		} else if s.peek() == minus {
			s.read()
			tok.Type = token.Decrement
		}
	case star:
		tok.Type = token.Mul
		if s.peek() == star {
			s.read()
			tok.Type = token.Pow
			if s.peek() == equal {
				s.read()
				tok.Type = token.PowAssign
			}
		} else if s.peek() == equal {
			s.read()
			tok.Type = token.MulAssign
		}
	case slash:
		tok.Type = token.Div
		if s.peek() == equal {
			s.read()
			tok.Type = token.DivAssign
		}
	case percent:
		tok.Type = token.Mod
		if s.peek() == equal {
			s.read()
			tok.Type = token.ModAssign
		}
	case ampersand:
		tok.Type = token.Band
		if s.peek() == ampersand {
			s.read()
			tok.Type = token.And
			if s.peek() == equal {
				s.read()
				tok.Type = token.AndAssign
			}
		} else if s.peek() == equal {
			s.read()
			tok.Type = token.BandAssign
		}
	case pipe:
		tok.Type = token.Bor
		if s.peek() == pipe {
			s.read()
			tok.Type = token.Or
			if s.peek() == equal {
				s.read()
				tok.Type = token.OrAssign
			}
		} else if s.peek() == equal {
			s.read()
			tok.Type = token.BorAssign
		}
	case tilde:
		tok.Type = token.Bnot
		if s.peek() == equal {
			s.read()
			tok.Type = token.BnotAssign
		}
	case caret:
		tok.Type = token.Bxor
		if s.peek() == equal {
			s.read()
			tok.Type = token.BxorAssign
		}
	case equal:
		tok.Type = token.Assign
		if s.peek() == equal {
			s.read()
			tok.Type = token.Eq
			if s.peek() == equal {
				s.read()
				tok.Type = token.Seq
			}
		} else if s.peek() == rangle {
			s.read()
			tok.Type = token.Arrow
		}
	case bang:
		tok.Type = token.Not
		if s.peek() == equal {
			s.read()
			tok.Type = token.Ne
			if s.peek() == equal {
				s.read()
				tok.Type = token.Sne
			}
		}
	case langle:
		tok.Type = token.Lt
		if s.peek() == equal {
			s.read()
			tok.Type = token.Le
		} else if s.peek() == langle {
			s.read()
			tok.Type = token.Lshift
			if s.peek() == equal {
				s.read()
				tok.Type = token.LshiftAssign
			}
		}
	case rangle:
		tok.Type = token.Gt
		if s.peek() == equal {
			s.read()
			tok.Type = token.Ge
		} else if s.peek() == rangle {
			s.read()
			tok.Type = token.Rshift
			if s.peek() == equal {
				s.read()
				tok.Type = token.RshiftAssign
			}
		}
	case question:
		tok.Type = token.Question
		if s.peek() == question {
			s.read()
			tok.Type = token.Nullish
			if s.peek() == equal {
				s.read()
				tok.Type = token.NullishAssign
			}
		} else if s.peek() == dot {
			s.read()
			tok.Type = token.Optional
		}
	default:
		tok.Type = token.Invalid
	}
	s.read()
}

func (s *Scanner) done() bool {
	return s.char == utf8.RuneError || s.char == 0
}

func (s *Scanner) runeFromRunes(n int) rune {
	var list []rune
	for i := 0; i < n; i++ {
		list = append(list, s.char)
		s.read()
	}
	i, err := strconv.ParseInt(string(list), 16, 64)
	if err != nil {
		return utf8.RuneError
	}
	return rune(i)
}

func (s *Scanner) read() {
	if s.curr >= len(s.input) {
		s.char = utf8.RuneError
		return
	}
	r, n := utf8.DecodeRune(s.input[s.next:])
	if r == utf8.RuneError {
		s.char = r
		s.next = len(s.input)
		return
	}
	s.old.Position = s.cursor.Position
	if r == nl {
		s.cursor.Line++
		s.cursor.Column = 0
	}
	s.cursor.Column++
	s.char, s.curr, s.next = r, s.next, s.next+n
}

func (s *Scanner) unread() {
	c, z := utf8.DecodeRune(s.input[s.curr:])
	s.char, s.curr, s.next = c, s.curr-z, s.curr
}

func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRune(s.input[s.next:])
	return r
}

func (s *Scanner) reset() {
	s.str.Reset()
}

func (s *Scanner) write() {
	s.writeRune(s.char)
}

func (s *Scanner) writeRune(char rune) {
	s.str.WriteRune(char)
}

func (s *Scanner) literal() string {
	return s.str.String()
}

func (s *Scanner) skip(accept func(rune) bool) {
	if s.done() {
		return
	}
	for accept(s.char) && !s.done() {
		s.read()
	}
}

func (s *Scanner) save() {
	s.old = s.cursor
}

func (s *Scanner) restore() {
	s.cursor = s.old
}

type scanMode int

const (
	modeDefault scanMode = iota
	modeTpl
	modeSub
)

type cursor struct {
	char rune
	curr int
	next int
	token.Position
}

var escapes = map[rune]rune{
	'0':       0,
	squote:    squote,
	dquote:    dquote,
	backslash: backslash,
	'n':       nl,
	'r':       cr,
	'v':       '\v',
	't':       tab,
	'b':       '\b',
	'f':       '\f',
	'x':       utf8.RuneError,
	'u':       utf8.RuneError,
}

const (
	lbrace     = '{'
	rbrace     = '}'
	lparen     = '('
	rparen     = ')'
	lsquare    = '['
	rsquare    = ']'
	langle     = '<'
	rangle     = '>'
	space      = ' '
	tab        = '\t'
	nl         = '\n'
	cr         = '\r'
	squote     = '\''
	dquote     = '"'
	underscore = '_'
	pound      = '#'
	dot        = '.'
	plus       = '+'
	minus      = '-'
	star       = '*'
	slash      = '/'
	percent    = '%'
	ampersand  = '&'
	pipe       = '|'
	question   = '?'
	bang       = '!'
	equal      = '='
	comma      = ','
	colon      = ':'
	semicolon  = ';'
	tilde      = '~'
	caret      = '^'
	backtick   = '`'
	dollar     = '$'
	backslash  = '\\'
)

func isSubstitution(r, k rune) bool {
	return r == dollar && k == lbrace
}

func isTemplate(r rune) bool {
	return r == backtick
}

func isComment(r, k rune) bool {
	return (r == slash && r == k) || (r == pound && k == bang)
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == underscore
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isBin(r rune) bool {
	return r == '0' || r == '1'
}

func isOctal(r rune) bool {
	return r >= '0' && r <= '7'
}

func isHex(r rune) bool {
	return isDigit(r) || (r >= 'a' && r <= 'f') && (r >= 'A' && r <= 'F')
}

func isAlpha(r rune) bool {
	return isLetter(r) || isDigit(r)
}

func isSpace(r rune) bool {
	return r == space || r == tab
}

func isQuote(r rune) bool {
	return isSingle(r) || isDouble(r)
}

func isDouble(r rune) bool {
	return r == dquote
}

func isSingle(r rune) bool {
	return r == squote
}

func isNL(r rune) bool {
	return r == nl || r == cr
}

func isEOL(r rune) bool {
	return isNL(r) || r == semicolon
}

func isBlank(r rune) bool {
	return isSpace(r) || isNL(r)
}
