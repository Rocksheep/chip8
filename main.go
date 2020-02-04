package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"

	"../chip8/chip8"

	"github.com/hajimehoshi/ebiten"
)

var windowWidth int = 300
var windowHeight int = 200

var screenWidth int = 256
var screenHeight int = 128

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
	chip8Border, _ := ebiten.NewImage(screenWidth+10, screenHeight+10, ebiten.FilterNearest)
	chip8Border.Fill(color.RGBA{0x00, 0xFF, 0x00, 0xFF})
	drawOptionsBorder := &ebiten.DrawImageOptions{}
	drawOptionsBorder.GeoM.Translate(screenX-5, screenY-5)
	screen.DrawImage(chip8Border, drawOptionsBorder)

	chip8Screen, _ := ebiten.NewImage(screenWidth+8, screenHeight+8, ebiten.FilterNearest)
	chip8Screen.Fill(color.RGBA{0x00, 0x00, 0x00, 0xFF})
	drawOptionsScreen := &ebiten.DrawImageOptions{}
	drawOptionsScreen.GeoM.Translate(screenX-4, screenY-4)
	screen.DrawImage(chip8Screen, drawOptionsScreen)

	pixelImage, _ := ebiten.NewImage(4, 4, ebiten.FilterNearest)
	pixelImage.Fill(color.RGBA{0, 0xFF, 0, 0xFF})
	buffer := processor.ScreenBuffer()
	for i, pixel := range buffer {
		if pixel == 1 {
			x := float64((i % 64) * 4)
			y := float64((i / 64) * 4)
			drawOptions := &ebiten.DrawImageOptions{}
			drawOptions.GeoM.Translate(screenX+x, screenY+y)
			screen.DrawImage(pixelImage, drawOptions)
		}
	}

	return nil
}

func startProcessor() {
	for true {
		processor.Step()
	}
}
