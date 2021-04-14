package main

import (
	"fmt"

	"github.com/ChrisMcKenzie/go-6502/cpu"
)

const (
	MaxMemory uint32 = 1024 * 64
)

type word uint16

type Memory struct {
	data [MaxMemory]byte
}

func (m *Memory) Init() {
	for i := uint32(0); i < MaxMemory; i++ {
		m.data[i] = 0
	}
}

type Flags uint8

const (
	C Flags = 1 << iota
	Z
	I
	D
	B
	Unused
	V
	N
)

func (f *Flags) Set(b Flags) { *f = (b | *f) }
func (f *Flags) SetIf(i bool, b Flags) {
	if i {
		f.Set(b)
	}
}
func (f *Flags) Clear(b Flags)    { *f = (b &^ *f) }
func (f *Flags) Toggle(b Flags)   { *f = (b ^ *f) }
func (f *Flags) Has(b Flags) bool { return (b & *f) != 0 }

type CPU struct {
	PC word
	SP byte

	mem *Memory

	cycles int

	A, X, Y byte

	Flags Flags
}

func (c *CPU) Reset() {
	c.ResetToVector(0xFFFC)
}

func (c *CPU) ResetToVector(v word) {
	c.PC = v
	c.SP = 0xFF
	c.A, c.X, c.Y = 0, 0, 0
	c.Flags = 0
	c.mem.Init()
}

func (c *CPU) FetchByte() byte {
	d := c.mem.data[c.PC]
	c.PC++
	c.cycles--
	return d
}

func (c *CPU) ReadByte(a byte) byte {
	d := c.mem.data[a]
	c.cycles--
	return d
}

func (c *CPU) FetchWord() word {
	var data word = word(c.mem.data[c.PC])
	c.PC++
	data |= (word(c.mem.data[c.PC]) << 8)
	c.PC++
	c.cycles -= 2

	return data
}

func (c *CPU) SPToAdress() word {
	return 0x100 | word(c.SP)
}

func (c *CPU) PushWordToStack(w word) {
	c.mem.data[c.SPToAdress()] = byte(w >> 8)
	c.SP--
	c.mem.data[c.SPToAdress()] = byte(w & 0xFF)
	c.SP--
}

func (c *CPU) Run(t int) {
	c.cycles = t
	for c.cycles > 0 {
		ins := c.FetchByte()
		switch ins {
		case cpu.INS_LDA_IM:
			c.LDA_IM()
		case cpu.INS_LDA_ZP:
			c.LDA_ZP()
		case cpu.INS_JSR:
			c.JSR()
		case cpu.INS_RTS:
			c.RTS()
		default:
			fmt.Printf("unhandled instruction: %#x\n", ins)
		}
	}
}

func (c *CPU) RTS() {
	var a word = word(c.mem.data[c.SP])
	a |= (word(c.mem.data[c.SP]) << 8)

	c.PC = a + 1
	c.cycles -= 2
}

func (c *CPU) JSR() {
	var r word = c.FetchWord()
	c.PushWordToStack(c.PC - 1)
	c.PC = r
	c.cycles--
}

func (c *CPU) LDA_IM() {
	val := c.FetchByte()
	c.A = val
	c.SetFlags(c.A)
}

func (c *CPU) LDA_ZP() {
	zpa := c.FetchByte()
	val := c.ReadByte(zpa)
	c.A = val
	c.SetFlags(c.A)
}

func (c *CPU) SetFlags(r byte) {
	// Set Zero Flag
	c.Flags.SetIf((r == 0), Z)
	// Set NegativeFlag
	c.Flags.SetIf(((r & 0b10000000) > 0), N)
}

func main() {
	mem := &Memory{}
	c := &CPU{mem: mem}

	c.Reset()
	mem.data[0xFFFC] = cpu.INS_JSR
	mem.data[0xFFFD] = 0x42
	mem.data[0xFFFE] = 0x42
	mem.data[0x4242] = 0x60

	c.Run(7)
	fmt.Printf("SP:%#x\nPC:%#x\n", c.SP, c.PC)
}
