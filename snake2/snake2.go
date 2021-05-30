package main
/*
30 May 21 -- Now that I think I understand the code more, I'm going to play w/ it a bit.
             The origin of the coordinate system is top left, as is typical in GUI's.
             There are several magic numbers in the code (as per Rob Pike's definition).
             I think I'll test this belief w/ some of my own coding.  Yeah, I'm right.
 */


import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"image/color"
	"time"
)

const PPS = 10  // pixels per square
const snooze = 250  // sleep time for game function.  In ms.
const windowSize = 80  // unit is squares, not pixels.  To get pixels must multiply by PPS.

type snakePart struct {
	x, y float32
}

//type moveType int  I don't think this was needed, so I'm testing that.  Yeah, I was right about that, too.

const (
	moveUp int = iota
	moveDown
	moveLeft
	moveRight
)

var snakeParts []snakePart
var game *fyne.Container
var head *canvas.Rectangle
var move = moveUp

func keyTyped(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyUp:
		move = moveUp
	case fyne.KeyDown:
		move = moveDown
	case fyne.KeyLeft:
		move = moveLeft
	case fyne.KeyRight:
		move = moveRight
	}
}

func main() {
	a := app.New()

	w := a.NewWindow("Snake")
	w.Resize(fyne.NewSize(windowSize * PPS, windowSize * PPS))
	w.SetFixedSize(true)

	game = setupGame()
	w.SetContent(game)
	w.Canvas().SetOnTypedKey(keyTyped)

	go runGame()
	w.ShowAndRun()
}

func refreshGame() {
	for i, seg := range snakeParts {
		rect := game.Objects[i]
		rect.Move(fyne.NewPos(seg.x*PPS, seg.y*PPS))
	}

	game.Refresh()
}

func runGame() {
	nextPart := snakePart{snakeParts[0].x, snakeParts[0].y - 1}
	for {
		headOldPos := fyne.NewPos(snakeParts[0].x*PPS, snakeParts[0].y*PPS)
		headNewPos := fyne.NewPos(nextPart.x*PPS, nextPart.y*PPS)
		headAniFunc := func(p fyne.Position) {
			head.Move(p)
			canvas.Refresh(head)
		}
		canvas.NewPositionAnimation(headOldPos, headNewPos, time.Millisecond*250, headAniFunc).Start()

		end := len(snakeParts) - 1
		tailOldPos := fyne.NewPos(snakeParts[end].x*PPS, snakeParts[end].y*PPS)
		tailNewPos := fyne.NewPos(snakeParts[end-1].x*PPS, snakeParts[end-1].y*PPS)
		tailAniFunc := func(p fyne.Position) {
			tail := game.Objects[end]
			tail.Move(p)
			canvas.Refresh(tail)
		}
		canvas.NewPositionAnimation(tailOldPos, tailNewPos, time.Millisecond*250, tailAniFunc).Start()

		time.Sleep(time.Millisecond * snooze)
		for i := len(snakeParts) - 1; i >= 1; i-- {
			snakeParts[i] = snakeParts[i-1]
		}
		snakeParts[0] = nextPart
		refreshGame()
		switch move {
		case moveUp:
			nextPart = snakePart{nextPart.x, nextPart.y - 1}
		case moveDown:
			nextPart = snakePart{nextPart.x, nextPart.y + 1}
		case moveLeft:
			nextPart = snakePart{nextPart.x - 1, nextPart.y}
		case moveRight:
			nextPart = snakePart{nextPart.x + 1, nextPart.y}
		}
	}
}

func setupGame() *fyne.Container {
	// the original code has too many magic numbers.  I'm changing that.
	var segments []fyne.CanvasObject
	startX := windowSize / 2
	startY := windowSize / 4
	for i := 0; i < windowSize/2; i++ {
		seg := snakePart{float32(startX), float32(startY + i)}
		snakeParts = append(snakeParts, seg)

		r := canvas.NewRectangle(&color.RGBA{G: 0x66, A: 0xff})
		r.Resize(fyne.NewSize(PPS, PPS))  // Pixels Per Square
		r.Move(fyne.NewPos(float32(startX*PPS), float32(PPS * (startY + i))))
		segments = append(segments, r)
	}

	head = canvas.NewRectangle(&color.RGBA{G: 0x66, A: 0xff})
	head.Resize(fyne.NewSize(PPS, PPS))
	head.Move(fyne.NewPos(snakeParts[0].x*PPS, snakeParts[0].y*PPS))
	segments = append(segments, head)
	return container.NewWithoutLayout(segments...)
}
