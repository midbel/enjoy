package token

import (
	"fmt"
)

type Token struct {
	Type    rune
	Literal string
	Offset  int
	Position
}

type Position struct {
	Line   int
	Column int
}

func (t Token) String() string {
	var prefix string
	switch t.Type {
	case EOF:
		return "<eof>"
	case EOL:
		return "<eol>"
	case Arrow:
		return "<arrow>"
	case Dot:
		return "<dot>"
	case Spread:
		return "<spread>"
	case Lbrace:
		return "<lbrace>"
	case Rbrace:
		return "<rbrace>"
	case Lparen:
		return "<lparen>"
	case Rparen:
		return "<rparen>"
	case Lsquare:
		return "<lsquare>"
	case Rsquare:
		return "<rsquare>"
	case BegSub:
		return "<begin-sub>"
	case EndSub:
		return "<end-sub>"
	case Template:
		return "<template>"
	case Colon:
		return "<colon>"
	case Comma:
		return "<comma>"
	case Question:
		return "<question>"
	case And:
		return "<and>"
	case AndAssign:
		return "<and-assign>"
	case Or:
		return "<or>"
	case OrAssign:
		return "<or-assign>"
	case Band:
		return "<bin-and>"
	case BandAssign:
		return "<bin-and-assign>"
	case Bor:
		return "<bin-or>"
	case BorAssign:
		return "<bin-or-assign>"
	case Bnot:
		return "<bin-not>"
	case BnotAssign:
		return "<bin-not-assign>"
	case Bxor:
		return "<bin-xor>"
	case BxorAssign:
		return "<bin-xor-assign>"
	case Assign:
		return "<assign>"
	case Increment:
		return "<increment>"
	case Decrement:
		return "<decrement>"
	case Add:
		return "<add>"
	case AddAssign:
		return "<add-assign>"
	case Sub:
		return "<sub>"
	case SubAssign:
		return "<sub-assign>"
	case Div:
		return "<div>"
	case DivAssign:
		return "<div-assign>"
	case Mul:
		return "<mul>"
	case MulAssign:
		return "<mul-assign>"
	case Pow:
		return "<pow>"
	case PowAssign:
		return "<pow-assign>"
	case Mod:
		return "<mod>"
	case ModAssign:
		return "<mod-assign>"
	case Not:
		return "<not>"
	case Eq:
		return "<eq>"
	case Seq:
		return "<strict-eq>"
	case Ne:
		return "<ne>"
	case Sne:
		return "<string-ne>"
	case Lt:
		return "<lt>"
	case Le:
		return "<le>"
	case Gt:
		return "<gt>"
	case Ge:
		return "<ge>"
	case Nullish:
		return "<nullish>"
	case NullishAssign:
		return "<nullish-assign>"
	case Optional:
		return "<optional>"
	case Keyword:
		prefix = "keyword"
	case String:
		prefix = "string"
	case Number:
		prefix = "number"
	case Boolean:
		prefix = "boolean"
	case Comment:
		prefix = "comment"
	case Ident:
		prefix = "identifier"
	case Invalid:
		prefix = "invalid"
	default:
		prefix = "unknown"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}
