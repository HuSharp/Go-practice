package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer
	cnt := 0
	for cnt < len(ins) {
		def, err := Lookup(ins[cnt])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[cnt+1:])
		fmt.Fprintf(&out, "%04d %s\n", cnt, ins.fmtInstruction(def, operands))

		cnt += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCnt := len(def.OperandWidths)

	if len(operands) != operandCnt {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCnt)
	}

	switch operandCnt {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

type Opcode byte

const (
	OpNull Opcode = iota
	OpConstant
	OpPop
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual       // ==
	OpNotEqual    // !=
	OpGreaterThan // >=
	OpMinus       // -
	OpBang        // !
	OpJumpNotTruthy
	OpJump
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpNull:     {"OpNull", []int{}},
	OpConstant: {"OpConstant", []int{2}},
	OpPop:      {"OpPop", []int{}},
	// Infix Expression
	OpAdd:   {"OpAdd", []int{}},
	OpSub:   {"OpSub", []int{}},
	OpMul:   {"OpMul", []int{}},
	OpDiv:   {"OpDiv", []int{}},
	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},
	// Comparison
	OpEqual:       {"OpEqual", []int{}},
	OpNotEqual:    {"OpNotEqual", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},
	// Prefix Expression
	OpMinus: {"OpMinus", []int{}},
	OpBang:  {"OpBang", []int{}},
	// Conditional
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpJump:          {"OpJump", []int{2}},
	// Bind Value
	OpGetGlobal: {"OpGetGlobal", []int{2}},
	OpSetGlobal: {"OpSetGlobal", []int{2}},
	// Array as composite data type, operand: the number of elements in an array literal
	OpArray: {"OpArray", []int{2}},
	OpHash:  {"OpHash", []int{2}},
	OpIndex: {"OpIndex", []int{}},
	// Function
	// OpCall tells the VM to start executing the *object.CompiledFunction sitting on top of the stack.
	// OpReturnValue tells the VM to return the value on top of the stack to the calling context and to resume execution there.
	// OpReturn excepts that there is no explicit value to return but an implicit vm.Null.
	OpCall:        {"OpCall", []int{1}},
	OpReturnValue: {"OpReturnValue", []int{}},
	OpReturn:      {"OpReturn", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undifined", op)
	}
	return def, nil
}

// Make instruction for Opcode
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}
	// The first thing is to find out how long the resulting instruction is going to be.
	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}

	return instruction
}

// ReadOperands supposed to be Make’s counterpart
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
			fmt.Println("read ", ReadUint16(ins[offset:]))
		}
		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
