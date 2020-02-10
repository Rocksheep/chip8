package chip8

import (
	"fmt"
	"math/rand"
)

// Chip8 ...
type Chip8 struct {
	memory           [4096]byte
	generalRegisters [16]byte // V
	registerI        uint16   // I
	soundTimer       byte
	delayTimer       byte
	stackPointer     byte
	programCounter   uint16
	stack            [16]uint16
	screenBuffer     [2048]byte // 64*32
}

type DebugCommand struct {
	Address     uint16
	Instruction uint16
}

// New Create a new instance
func New() *Chip8 {
	memory := [4096]byte{
		/* '0' */ 0xF0, 0x90, 0x90, 0x90, 0xF0,
		/* '1' */ 0x20, 0x60, 0x20, 0x20, 0x70,
		/* '2' */ 0xF0, 0x10, 0xF0, 0x80, 0xF0,
		/* '3' */ 0xF0, 0x10, 0xF0, 0x10, 0xF0,
		/* '4' */ 0x90, 0x90, 0xF0, 0x10, 0x10,
		/* '5' */ 0xF0, 0x80, 0xF0, 0x10, 0xF0,
		/* '6' */ 0xF0, 0x80, 0xF0, 0x90, 0xF0,
		/* '7' */ 0xF0, 0x10, 0x20, 0x40, 0x40,
		/* '8' */ 0xF0, 0x90, 0xF0, 0x90, 0xF0,
		/* '9' */ 0xF0, 0x90, 0xF0, 0x10, 0xF0,
		/* 'A' */ 0xF0, 0x90, 0xF0, 0x90, 0x90,
		/* 'B' */ 0xE0, 0x90, 0xE0, 0x90, 0xE0,
		/* 'C' */ 0xF0, 0x80, 0x80, 0x80, 0xF0,
		/* 'D' */ 0xE0, 0x80, 0x80, 0x80, 0xE0,
		/* 'E' */ 0xF0, 0x80, 0xF0, 0x80, 0xF0,
		/* 'F' */ 0xF0, 0x80, 0xF0, 0x80, 0x80,
	}

	return &Chip8{
		memory,
		[16]byte{},
		0,
		0,
		0,
		0,
		0x200,
		[16]uint16{},
		[2048]byte{},
	}
}

func (chip8 *Chip8) Step() {
	var data uint16 = uint16(chip8.memory[chip8.programCounter])<<8 | uint16(chip8.memory[chip8.programCounter+1])
	// fmt.Printf("%X\n", data)

	switch data & 0xF000 {
	case 0x0000:
		if data&0xFF == 0xEE {
			chip8.stackPointer--
			chip8.programCounter = chip8.stack[chip8.stackPointer] + 2
		}
		break
	case 0x1000:
		chip8.programCounter = data & 0xFFF
		break
	case 0x2000:
		chip8.call(data & 0x0FFF)
		break
	case 0x3000:
		register := (data & 0x0F00) >> 8
		value := byte(data)
		if chip8.generalRegisters[register] == value {
			chip8.programCounter += 2
		}
		chip8.programCounter += 2
		break
	case 0x4000:
		register := (data & 0x0F00) >> 8
		value := byte(data)
		if chip8.generalRegisters[register] != value {
			chip8.programCounter += 2
		}
		chip8.programCounter += 2
		break
	case 0x6000:
		register := (data & 0x0F00) >> 8
		value := byte(data)
		chip8.generalRegisters[register] = value
		chip8.programCounter += 2
		break
	case 0x7000:
		register := (data & 0x0F00) >> 8
		chip8.generalRegisters[register] += byte(data & 0xFF)
		chip8.programCounter += 2
		break
	case 0x8000:
		registerA := (data & 0xF00) >> 8
		registerB := (data & 0xF0) >> 4

		switch data & 0xF {
		case 0x0:
			chip8.generalRegisters[registerA] = chip8.generalRegisters[registerB]
			chip8.programCounter += 2
			break
		case 0x2:
			chip8.generalRegisters[registerA] &= chip8.generalRegisters[registerB]
			chip8.programCounter += 2
			break
		case 0x4:
			result := uint16(chip8.generalRegisters[registerA]) + uint16(chip8.generalRegisters[registerB])
			if result > 255 {
				chip8.generalRegisters[0xF] = 1
			} else {
				chip8.generalRegisters[0xF] = 0
			}
			chip8.generalRegisters[registerA] = byte(result)
			chip8.programCounter += 2
			break
		case 0x5:
			if chip8.generalRegisters[registerA] > chip8.generalRegisters[registerB] {
				chip8.generalRegisters[0xF] = 1
			} else {
				chip8.generalRegisters[0xF] = 0
			}
			chip8.generalRegisters[registerA] -= chip8.generalRegisters[registerB]
			chip8.programCounter += 2
			break
		}
		break
	case 0xA000:
		chip8.registerI = data & 0x0FFF
		chip8.programCounter += 2
		break
	case 0xC000:
		register := (data & 0xF00) >> 8
		chip8.generalRegisters[register] = byte(data) & byte(rand.Intn(255))
		chip8.programCounter += 2
		break
	case 0xD000:
		chip8.draw(data)
		break
	case 0xE000:
		//TODO: Add actual keyboard checks
		if data&0xFF == 0xA1 {
			chip8.programCounter += 2
		}
		chip8.programCounter += 2
		break
	case 0xF000:
		switch data & 0xFF {
		case 0x07:
			register := data & 0x0F00 >> 8
			chip8.generalRegisters[register] = chip8.delayTimer
			chip8.programCounter += 2
		case 0x15:
			value := byte(data & 0x0F00 >> 8)
			chip8.delayTimer = value
			chip8.programCounter += 2
			break
		case 0x18:
			register := (data & 0xF00) >> 8
			chip8.soundTimer = chip8.generalRegisters[register]
			chip8.programCounter += 2
			break
		case 0x29:
			register := (data & 0xF00) >> 8
			value := uint16(chip8.generalRegisters[register])
			chip8.registerI = value * 5
			chip8.programCounter += 2
			break
		case 0x33:
			value := data & 0x0F00 >> 8
			for i := uint16(3); i > 0; i-- {
				chip8.memory[chip8.registerI+i-1] = byte(value % 10)
				value /= 10
			}
			chip8.programCounter += 2
			break
		case 0x65:
			limit := data & 0x0F00 >> 8
			for i := uint16(0); i <= limit; i++ {
				chip8.generalRegisters[i] = chip8.memory[chip8.registerI+i]
			}
			chip8.programCounter += 2
			break
		}
	default:
		fmt.Printf("Unknown command: %X\n", data)
	}
}

func (chip8 *Chip8) call(address uint16) {
	chip8.stack[chip8.stackPointer] = chip8.programCounter
	chip8.stackPointer++
	chip8.programCounter = address
}

func (chip8 *Chip8) draw(data uint16) {
	vXAddress := (data & 0x0F00) >> 8
	vYAddress := (data & 0x00F0) >> 4
	spriteSize := byte(data & 0x000F)

	vX := uint16(chip8.generalRegisters[vXAddress])
	vY := uint16(chip8.generalRegisters[vYAddress]) * 64
	chip8.generalRegisters[0xF] = 0

	for line := byte(0); line < spriteSize; line++ {
		spriteAddress := chip8.registerI + uint16(line)
		sprite := chip8.memory[spriteAddress]

		y := vY + 64*uint16(line)
		for i := uint16(0); i < 8; i++ {
			x := (vX + i) % 64
			prevValue := chip8.screenBuffer[x+y]
			newValue := prevValue ^ ((sprite >> (7 - i)) & 1)
			if prevValue != newValue && newValue == 0 {
				chip8.generalRegisters[0xF] = 1
			}
			chip8.screenBuffer[x+y] = newValue
		}
	}

	chip8.programCounter += 2
}

func (chip8 *Chip8) LoadProgram(program []byte) {
	programStart := 0x200

	for i := 0; i < len(program); i++ {
		chip8.memory[i+programStart] = program[i]
	}
}

func (chip8 *Chip8) ScreenBuffer() [2048]byte {
	return chip8.screenBuffer
}

func (chip8 *Chip8) GetNextCommands(n uint16) []DebugCommand {
	commands := make([]DebugCommand, n)
	for i := uint16(0); i < n; i++ {
		address := chip8.programCounter + (i * 2)
		command := uint16(chip8.memory[address])<<8 | uint16(chip8.memory[address+1])

		commands[i] = DebugCommand{address, command}
	}

	return commands
}

func (chip8 *Chip8) GetMemoryAtAddress(address uint16) byte {
	return chip8.memory[address]
}

func (chip8 *Chip8) V() [16]byte {
	return chip8.generalRegisters
}

func (chip8 *Chip8) I() uint16 {
	return chip8.registerI
}
