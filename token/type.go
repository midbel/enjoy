package token

import (
	"fmt"
)

const (
	EOF rune = -(iota + 1)
	EOL
	Comment
	Ident
	Keyword
	String
	Number
	Boolean
	Dot
	Template
	BegSub
	EndSub
	Lparen
	Rparen
	Lsquare
	Rsquare
	Lbrace
	Rbrace
	Assign
	Increment
	Decrement
	Not
	And
	AndAssign
	Or
	OrAssign
	Eq
	Seq
	Sne
	Ne
	Lt
	Le
	Gt
	Ge
	Add
	AddAssign
	Sub
	SubAssign
	Mul
	MulAssign
	Div
	DivAssign
	Pow
	PowAssign
	Mod
	ModAssign
	Lshift
	LshiftAssign
	Rshift
	RshiftAssign
	Band
	BandAssign
	Bor
	BorAssign
	Bnot
	BnotAssign
	Bxor
	BxorAssign
	Comma
	Colon
	Question
	Nullish
	NullishAssign
	Optional
	Arrow
	Spread
	Invalid
)

func ConvertAssignToken(op rune) (rune, error) {
	switch op {
	default:
		return -1, fmt.Errorf("invalid assignment token")
	case Assign:
	case AddAssign:
		op = Add
	case SubAssign:
		op = Sub
	case MulAssign:
		op = Mul
	case DivAssign:
		op = Div
	case PowAssign:
		op = Pow
	case ModAssign:
		op = Mod
	case NullishAssign:
		op = Nullish
	case AndAssign:
		op = And
	case OrAssign:
		op = Or
	case LshiftAssign:
		op = Lshift
	case RshiftAssign:
		op = Rshift
	case BandAssign:
		op = Band
	case BorAssign:
		op = Bor
	case BxorAssign:
		op = Bxor
	}
	return op, nil
}

func CanSkipBlanks(k rune) bool {
	switch k {
	default:
		return false
	case Lparen:
	case Lsquare:
	case Lbrace:
	case Comma:
	}
	return true
}
