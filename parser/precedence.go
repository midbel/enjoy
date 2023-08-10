package parser

import (
	"github.com/midbel/enjoy/token"
)

const (
	powLowest int = iota
	powComma
	powAssign
	powRelation
	powBitwise
	powCompare
	powShift
	powAdd
	powMul
	powUnary
	powIncr
	powCall
	powObject
	powGroup
)

type powerSet map[rune]int

var powers = powerSet{
	token.Comma:         powComma,
	token.Arrow:         powComma,
	token.Dot:           powObject,
	token.Optional:      powObject,
	token.Lparen:        powCall,
	token.Lsquare:       powCall,
	token.Add:           powAdd,
	token.Sub:           powAdd,
	token.Mul:           powMul,
	token.Div:           powMul,
	token.Mod:           powMul,
	token.Pow:           powMul,
	token.Lshift:        powShift,
	token.Rshift:        powShift,
	token.Spread:        powAssign,
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
	token.Question:      powAssign,
	token.And:           powRelation,
	token.Or:            powRelation,
	token.Nullish:       powRelation,
	token.Band:          powBitwise,
	token.Bor:           powBitwise,
	token.Bxor:          powBitwise,
	token.Eq:            powCompare,
	token.Seq:           powCompare,
	token.Ne:            powCompare,
	token.Sne:           powCompare,
	token.Lt:            powCompare,
	token.Le:            powCompare,
	token.Gt:            powCompare,
	token.Ge:            powCompare,
}

func (ps powerSet) Get(kind rune) int {
	return ps[kind]
}
