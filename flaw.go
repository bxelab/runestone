package go_runestone

type Flaw int

const (
	EdictOutput Flaw = iota
	EdictRuneId
	InvalidScript
	Opcode
	SupplyOverflow
	TrailingIntegers
	TruncatedField
	UnrecognizedEvenTag
	UnrecognizedFlag
	Varint
)

var flawToString = map[Flaw]string{
	EdictOutput:         "edict output greater than transaction output count",
	EdictRuneId:         "invalid rune ID in edict",
	InvalidScript:       "invalid script in OP_RETURN",
	Opcode:              "non-pushdata opcode in OP_RETURN",
	SupplyOverflow:      "supply overflows u128",
	TrailingIntegers:    "trailing integers in body",
	TruncatedField:      "field with missing value",
	UnrecognizedEvenTag: "unrecognized even tag",
	UnrecognizedFlag:    "unrecognized field",
	Varint:              "invalid varint",
}

func (f Flaw) String() string {
	return flawToString[f]
}
