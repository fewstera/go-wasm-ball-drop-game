package main

import (
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	holeSize             = 120
	distanceBetweenLines = 80
	playerRadius         = 20
	playerColor          = "red"
	lineColor            = "green"
)

var (
	width    int
	height   int
	doc      js.Value
	ctx      js.Value
	lines    []*line
	playerX  int
	playerY  int
	gameOver bool
	score    int
)

type line struct {
	// Y position of the line
	Y int
	// Where the Hole in the line starts (x position)
	HoleStart int
	// The width the Hole
	HoleWidth int
}

func main() {
	done := make(chan struct{}, 0)

	doc = js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "canvas")
	width = canvasEl.Get("clientWidth").Int()
	height = canvasEl.Get("clientHeight").Int()
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)
	ctx = canvasEl.Call("getContext", "2d")

	playerX = width / 2
	playerY = height / 2

	setupControls()

	go updateGame(done)
	go updateCanvas()

	fmt.Println("Starting")
	fmt.Printf("Width: %v, height: %v\n", width, height)

	<-done
	gameOver = true
	updateCanvas()
}

func updateCanvas() {
	ctx.Call("clearRect", 0, 0, width, height)

	for _, line := range lines {
		ctx.Call("beginPath")
		ctx.Call("moveTo", 0, line.Y)
		// Draw first line segment
		ctx.Call("lineTo", line.HoleStart, line.Y)
		// Skip hole
		ctx.Call("moveTo", line.HoleStart+line.HoleWidth, line.Y)
		// Draw end of line
		ctx.Call("lineTo", width, line.Y)
		ctx.Call("stroke")
		ctx.Call("closePath")
	}

	// Draw player circle
	ctx.Call("beginPath")
	ctx.Set("fillStyle", playerColor)
	ctx.Call("arc", playerX, playerY, playerRadius, 0, math.Pi*2, true)
	ctx.Call("closePath")
	ctx.Call("fill")

	if !gameOver {
		// Draw score
		ctx.Set("font", "30px Helvetica bold")
		ctx.Set("fillStyle", playerColor)
		ctx.Call("fillText", getCommaSeperatedNumber(score), 20, 50)

		time.Sleep(time.Millisecond)
		updateCanvas()
	} else {
		finalTextLineOne := "YOU LOST."
		finalTextLineTwo := fmt.Sprintf("Final score: %v", getCommaSeperatedNumber(score))

		centerX := width / 2
		centerY := height / 2

		ctx.Set("font", "30px Helvetica bold")
		lineOneWidth := ctx.Call("measureText", finalTextLineOne).Get("width").Int()
		lineTwoWidth := ctx.Call("measureText", finalTextLineTwo).Get("width").Int()

		ctx.Call("beginPath")
		ctx.Call("rect", centerX-(lineTwoWidth/2)-20, centerY-80, lineTwoWidth+40, 160)
		ctx.Set("fillStyle", "grey")
		ctx.Call("fill")
		ctx.Call("stroke")
		ctx.Call("closePath")

		ctx.Call("beginPath")
		ctx.Set("fillStyle", "black")
		ctx.Call("fillText", finalTextLineOne, centerX-lineOneWidth/2, centerY-30)
		ctx.Call("fillText", finalTextLineTwo, centerX-lineTwoWidth/2, centerY+30)
		ctx.Call("closePath")
	}
}

func updateGame(done chan struct{}) {
	score++
	newestLinePosition := 0
	var closestLineToPlayer *line
	for _, line := range lines {
		line.Y = line.Y - 1
		newestLinePosition = line.Y

		if line.Y+5+playerRadius >= playerY && closestLineToPlayer == nil {
			closestLineToPlayer = line
		}
	}

	shouldAddNewLine := (height-newestLinePosition > distanceBetweenLines)
	if shouldAddNewLine {
		holePosition := rand.Intn(width - holeSize)
		lines = append(lines, &line{height, holePosition, holeSize})
	}

	shouldDeleteFirstLine := (lines[0].Y < 0)
	if shouldDeleteFirstLine {
		lines = lines[1:]
	}

	if closestLineToPlayer != nil {
		playerOnLine := math.Abs(float64(playerY-closestLineToPlayer.Y+playerRadius)) < 5

		holeMinX := closestLineToPlayer.HoleStart + playerRadius
		holeMaxX := closestLineToPlayer.HoleStart + closestLineToPlayer.HoleWidth - playerRadius
		playerIsInHole := playerX > holeMinX && playerX < holeMaxX

		if playerOnLine && !playerIsInHole {
			playerY = closestLineToPlayer.Y - playerRadius
		} else {
			playerY = playerY + 2
		}
	}

	if playerY+playerRadius > height {
		playerY = height - playerRadius
	}

	if playerY-playerRadius <= 0 {
		done <- struct{}{}
	} else {
		time.Sleep(time.Duration(10) * time.Millisecond)
		updateGame(done)
	}
}

func setupControls() {
	kCall := js.NewCallback(keypressHandler)
	doc.Call("addEventListener", "keydown", kCall)
}

func keypressHandler(args []js.Value) {
	event := args[0]
	key := event.Get("key").String()
	switch key {
	case "ArrowLeft", "a", "A", "4":
		playerX = playerX - 25
		if playerX < playerRadius {
			playerX = playerRadius
		}
	case "ArrowRight", "d", "D", "6":
		playerX = playerX + 25
		if playerX+playerRadius-width >= 0 {
			playerX = width - playerRadius
		}
	}
}

func getCommaSeperatedNumber(number int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d\n", number)
}
