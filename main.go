package main

import (
	"chip8/chip8"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten"
)

var windowWidth int = 300
var windowHeight int = 200

var screenWidth int = 128
var screenHeight int = 64

var screenX float64 = float64(windowWidth/2 - screenWidth/2)
var screenY float64 = float64(windowHeight/2 - screenHeight/2)

var processor *chip8.Chip8

func main() {
	input, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	processor = chip8.New()
	processor.LoadProgram(input)

	go startProcessor()

	if err := ebiten.Run(update, windowWidth, windowHeight, 2, "Chip8"); err != nil {
		log.Fatal(err)
	}
}

func update(screen *ebiten.Image) error {
	if ebiten.IsDrawingSkipped() {
		fmt.Println("Skipping")
		return nil
	}

	// 64 x 32
	chip8Display, _ := ebiten.NewImage(screenWidth, screenHeight, ebiten.FilterNearest)
	chip8Display.Fill(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	drawOptions := &ebiten.DrawImageOptions{}
	drawOptions.GeoM.Translate(screenX, screenY)
	screen.DrawImage(chip8Display, drawOptions)

	return nil
}

func startProcessor() {
	for true {
		processor.Step()
	}
}
