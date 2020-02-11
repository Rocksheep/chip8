package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/image/font"

	"chip8/chip8"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/text"
)

var windowWidth int = 500
var windowHeight int = 200

var screenWidth int = 256
var screenHeight int = 128

var screenX float64 = float64(windowWidth/2 - screenWidth/2)
var screenY float64 = float64(windowHeight/2 - screenHeight/2)

var processor *chip8.Chip8

var fontSize float64 = 6.7
var gameFont font.Face

var prevSpacePressed bool
var prevEnterPressed bool
var debugMode bool

func main() {
	input, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Draw Debug Information
	tt, _ := truetype.Parse(fonts.ArcadeN_ttf)
	const dpi = 72
	gameFont = truetype.NewFace(tt, &truetype.Options{Size: fontSize, DPI: dpi, Hinting: font.HintingVertical})

	processor = chip8.New()
	processor.LoadProgram(input)
	prevSpacePressed = false
	prevEnterPressed = false
	debugMode = true

	ebiten.SetMaxTPS(500)
	if err := ebiten.Run(update, windowWidth, windowHeight, 2, "Chip8"); err != nil {
		log.Fatal(err)
	}
}

func update(screen *ebiten.Image) error {
	pressedKeys := map[uint8]bool{}

	pressedKeys[0x1] = ebiten.IsKeyPressed(ebiten.Key1)
	pressedKeys[0x2] = ebiten.IsKeyPressed(ebiten.Key2)
	pressedKeys[0x3] = ebiten.IsKeyPressed(ebiten.Key3)
	pressedKeys[0xC] = ebiten.IsKeyPressed(ebiten.Key4)

	pressedKeys[0x4] = ebiten.IsKeyPressed(ebiten.KeyQ)
	pressedKeys[0x5] = ebiten.IsKeyPressed(ebiten.KeyW)
	pressedKeys[0x6] = ebiten.IsKeyPressed(ebiten.KeyE)
	pressedKeys[0xD] = ebiten.IsKeyPressed(ebiten.KeyR)

	pressedKeys[0x7] = ebiten.IsKeyPressed(ebiten.KeyA)
	pressedKeys[0x8] = ebiten.IsKeyPressed(ebiten.KeyS)
	pressedKeys[0x9] = ebiten.IsKeyPressed(ebiten.KeyD)
	pressedKeys[0xE] = ebiten.IsKeyPressed(ebiten.KeyF)

	pressedKeys[0xA] = ebiten.IsKeyPressed(ebiten.KeyZ)
	pressedKeys[0x0] = ebiten.IsKeyPressed(ebiten.KeyX)
	pressedKeys[0xB] = ebiten.IsKeyPressed(ebiten.KeyC)
	pressedKeys[0xF] = ebiten.IsKeyPressed(ebiten.KeyV)

	processor.SetPressedKeys(pressedKeys)

	spacePressed := ebiten.IsKeyPressed(ebiten.KeySpace)
	enterPressed := ebiten.IsKeyPressed(ebiten.KeyEnter)

	if debugMode {
		if spacePressed && spacePressed != prevSpacePressed {
			processor.Step()
		}
	} else {
		processor.Step()
	}

	if enterPressed && enterPressed != prevEnterPressed {
		debugMode = !debugMode
	}

	prevEnterPressed = enterPressed
	prevSpacePressed = spacePressed

	if ebiten.IsDrawingSkipped() {
		fmt.Println("Skipping")
		return nil
	}

	screen.Fill(color.RGBA{56, 60, 74, 0xFF})

	gameScreen, _ := ebiten.NewImage(screenWidth+10, screenHeight+10, ebiten.FilterDefault)
	gameScreenOptions := &ebiten.DrawImageOptions{}
	gameScreenOptions.GeoM.Translate(5, 5)
	drawGame(gameScreen)
	screen.DrawImage(gameScreen, gameScreenOptions)

	memoryWindow, _ := ebiten.NewImage(220, 85, ebiten.FilterDefault)
	memoryWindowOptions := &ebiten.DrawImageOptions{}
	memoryWindowOptions.GeoM.Translate(float64(screenWidth+20), 5)
	drawMemory(memoryWindow)
	screen.DrawImage(memoryWindow, memoryWindowOptions)

	registerWindow, _ := ebiten.NewImage(220, 85, ebiten.FilterDefault)
	registerWindowOptions := &ebiten.DrawImageOptions{}
	registerWindowOptions.GeoM.Translate(float64(screenWidth+20), 95)
	drawRegisters(registerWindow)
	screen.DrawImage(registerWindow, registerWindowOptions)

	return nil
}

func drawGame(screen *ebiten.Image) {
	drawBorders(screen)
	drawGameRect(screen)
	pixelImage, _ := ebiten.NewImage(4, 4, ebiten.FilterNearest)
	pixelImage.Fill(color.RGBA{0, 0, 0, 0xFF})
	buffer := processor.ScreenBuffer()
	for i, pixel := range buffer {
		if pixel == 1 {
			x := float64((i % 64) * 4)
			y := float64((i / 64) * 4)
			drawOptions := &ebiten.DrawImageOptions{}
			drawOptions.GeoM.Translate(8+x, 8+y)
			screen.DrawImage(pixelImage, drawOptions)
		}
	}
}

func drawMemory(screen *ebiten.Image) {
	drawBorders(screen)
	width, height := screen.Size()
	background, _ := ebiten.NewImage(width-2, height-2, ebiten.FilterDefault)
	background.Fill(color.RGBA{56, 60, 74, 0xFF})
	backgroundOptions := &ebiten.DrawImageOptions{}
	backgroundOptions.GeoM.Translate(1, 1)
	screen.DrawImage(background, backgroundOptions)

	commands := processor.GetNextCommands(10)
	for i := 0; i < 10; i++ {
		instructionText := fmt.Sprintf("%04X", commands[i].Instruction)

		switch commands[i].Instruction & 0xF000 {
		case 0x0000:
			switch commands[i].Instruction & 0xFF {
			case 0xE0:
				instructionText = "CLS"
				break
			case 0xEE:
				instructionText = "RET"
				break
			}
			break
		case 0x1000:
			addr := commands[i].Instruction & 0xFFF
			instructionText = fmt.Sprintf("JMP %03X", addr)
			break
		case 0x2000:
			addr := commands[i].Instruction & 0xFFF
			instructionText = fmt.Sprintf("CALL %03X", addr)
			break
		case 0x6000:
			v := commands[i].Instruction & 0xF00 >> 8
			kk := byte(commands[i].Instruction)
			instructionText = fmt.Sprintf("SET v%X, %02X", v, kk)
			break
		case 0x7000:
			v := commands[i].Instruction & 0xF00 >> 8
			kk := byte(commands[i].Instruction)
			instructionText = fmt.Sprintf("ADD v%X, %02X", v, kk)
			break
		case 0x8000:
			switch commands[i].Instruction & 0xF {
			case 0x3:
				vX := commands[i].Instruction & 0xF00 >> 8
				vY := commands[i].Instruction & 0xF0 >> 4
				instructionText = fmt.Sprintf("XOR V%X, V%X", vX, vY)
				break
			case 0x6:
				v := commands[i].Instruction & 0xF00 >> 8
				instructionText = fmt.Sprintf("SHR v%X", v)
				break
			}
			break
		case 0xA000:
			instructionText = fmt.Sprintf("SET I, %03X", commands[i].Instruction&0xFFF)
			break
		case 0xD000:
			v1 := commands[i].Instruction & 0xF00 >> 8
			v2 := commands[i].Instruction & 0xF0 >> 4
			n := commands[i].Instruction & 0xF
			instructionText = fmt.Sprintf("DRW v%X, v%X, %X", v1, v2, n)
			break
		case 0xF000:
			value := (commands[i].Instruction & 0xF00) >> 8
			switch commands[i].Instruction & 0xFF {
			case 0x0A:
				instructionText = fmt.Sprintf("LD v%X, K", value)
				break
			case 0x1E:
				instructionText = fmt.Sprintf("ADD I, v%X", value)
				break
			case 0x29:
				instructionText = fmt.Sprintf("LD F, v%X", value)
				break
			case 0x33:
				instructionText = fmt.Sprintf("LD B, v%X", value)
				break
			case 0x55:
				instructionText = fmt.Sprintf("LD [I], v%X", value)
				break
			case 0x65:
				instructionText = fmt.Sprintf("LD v%X, [I]", value)
				break
			}
			break
		}

		value := fmt.Sprintf("%04X - %s", commands[i].Address, instructionText)
		text.Draw(screen, value, gameFont, 3, int(fontSize+3)+i*int(fontSize+2), color.White)
	}
}

func drawRegisters(screen *ebiten.Image) {
	drawBorders(screen)
	width, height := screen.Size()
	background, _ := ebiten.NewImage(width-2, height-2, ebiten.FilterDefault)
	background.Fill(color.RGBA{56, 60, 74, 0xFF})
	backgroundOptions := &ebiten.DrawImageOptions{}
	backgroundOptions.GeoM.Translate(1, 1)
	screen.DrawImage(background, backgroundOptions)
	colX := width / 2

	v := processor.V()
	for i := 0; i < 10; i++ {
		value := fmt.Sprintf("V%X = %02X", i, v[i])
		text.Draw(screen, value, gameFont, 3, int(fontSize+3)+i*int(fontSize+2), color.White)
	}

	for i := 10; i < len(v); i++ {
		value := fmt.Sprintf("V%X = %02X", i, v[i])
		text.Draw(screen, value, gameFont, colX, int(fontSize+3)+(i-10)*int(fontSize+2), color.White)
	}

	value := fmt.Sprintf("I = %04X (%02X)", processor.I(), processor.GetMemoryAtAddress(processor.I()))
	text.Draw(screen, value, gameFont, colX, int(fontSize+3)+7*int(fontSize+2), color.White)
}

func drawBorders(screen *ebiten.Image) {
	width, height := screen.Size()
	chip8BorderLight, _ := ebiten.NewImage(width, height, ebiten.FilterNearest)
	chip8BorderLight.Fill(color.RGBA{0x7c, 0x81, 0x8c, 0xFF})
	drawOptionsBorderLight := &ebiten.DrawImageOptions{}
	screen.DrawImage(chip8BorderLight, drawOptionsBorderLight)

	chip8BorderDark, _ := ebiten.NewImage(width-1, height-1, ebiten.FilterNearest)
	chip8BorderDark.Fill(color.RGBA{0x00, 0x00, 0x00, 0xFF})
	drawOptionsBorderDark := &ebiten.DrawImageOptions{}
	screen.DrawImage(chip8BorderDark, drawOptionsBorderDark)
}

func drawGameRect(screen *ebiten.Image) {
	width, height := screen.Size()
	chip8Screen, _ := ebiten.NewImage(width-2, height-2, ebiten.FilterNearest)
	chip8Screen.Fill(color.RGBA{0xc2, 0xc5, 0xcc, 0xFF})
	drawOptionsScreen := &ebiten.DrawImageOptions{}
	drawOptionsScreen.GeoM.Translate(1, 1)
	screen.DrawImage(chip8Screen, drawOptionsScreen)
}
