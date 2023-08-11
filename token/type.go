package token

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
