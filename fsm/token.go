package fsm

type RawTokenKind int

const (
	TokenDigit RawTokenKind = iota
	TokenKey
	TokenRepeat
	TokenSystem
)

type RawToken struct {
	Kind  RawTokenKind
	Value string
}