package parser

import (
	"github.com/midbel/enjoy/token"
)

const (
	powLowest int = iota
	powComma
	powAssign
	powOr
	powAnd
	powBitOr
	powBitXor
	powBitAnd
	powEqual
	powCompare
	powShift
	powAdd
	powMul
	powPow
	powUnary
	powPostfix
	powNew
	powObject
	powGroup
)

type powerSet map[rune]int

var powers = powerSet{
	token.Comma:         powComma,
	token.Arrow:         powAssign,
	token.Spread:        powAssign,
	token.Question:      powAssign,
	token.Assign:        powAssign,
	token.Colon:         powAssign,
	token.AddAssign:     powAssign,
	token.SubAssign:     powAssign,
	token.MulAssign:     powAssign,
	token.DivAssign:     powAssign,
	token.PowAssign:     powAssign,
	token.ModAssign:     powAssign,
	token.NullishAssign: powAssign,
	token.AndAssign:     powAssign,
	token.OrAssign:      powAssign,
	token.LshiftAssign:  powAssign,
	token.RshiftAssign:  powAssign,
	token.BandAssign:    powAssign,
	token.BorAssign:     powAssign,
	token.BxorAssign:    powAssign,
	token.Nullish:       powOr,
	token.Or:            powOr,
	token.And:           powAnd,
	token.Bor:           powBitOr,
	token.Bxor:          powBitXor,
	token.Band:          powBitAnd,
	token.Eq:            powEqual,
	token.Seq:           powEqual,
	token.Ne:            powEqual,
	token.Sne:           powEqual,
	token.Lt:            powCompare,
	token.Le:            powCompare,
	token.Gt:            powCompare,
	token.Ge:            powCompare,
	token.Lshift:        powShift,
	token.Rshift:        powShift,
	token.Add:           powAdd,
	token.Sub:           powAdd,
	token.Mul:           powMul,
	token.Div:           powMul,
	token.Mod:           powMul,
	token.Pow:           powPow,
	token.Dot:           powObject,
	token.Optional:      powObject,
	token.Lparen:        powObject,
	token.Lsquare:       powObject,
}

func (ps powerSet) Get(kind rune) int {
	return ps[kind]
}
