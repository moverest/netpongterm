package main

import (
	"fmt"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// BallSpeed stores ball speed
type BallSpeed struct {
	X int32
	Y int32
}

// Ball stores ball position and size
//   (X,Y)
//     +-------+
//     |       |
//     |       |  Height
//     +-------+
//       Width
type Ball struct {
	X      int32
	Y      int32
	Width  int32
	Height int32
}

// Paddle stores a player paddle position and size
// (X,Y) being the center of the paddle
type Paddle struct {
	X    int32
	Y    int32
	Size int32
}

// GameMode defines game modes
type GameMode int8

const (
	// GameModePlaying indicates that the game is on
	GameModePlaying GameMode = iota
	// GameModeLeftPlayerLoosed indicates that the left player has lost
	GameModeLeftPlayerLoosed
	// GameModeRightPlayerLoosed indicates that the right player has lost
	GameModeRightPlayerLoosed
)

// GameState stores game state
type GameState struct {
	Mode        GameMode
	ModeTimeOut int16

	Ball      Ball
	BallSpeed BallSpeed

	LeftPlayerPaddle  Paddle
	RightPlayerPaddle Paddle

	Tick float64

	Height int32
	Width  int32

	LeftPlayerScore  int32
	RightPlayerScore int32

	Debug int8
}

// Paddle to move in PaddleMovement.Player
const (
	MoveNonePaddle  = iota
	MoveLeftPaddle  = iota
	MoveRightPaddle = iota
)

// PaddleMovement defines a mouve action
type PaddleMovement struct {
	RelativeY int32
	Player    int8
}

// Btoi8 converts bool to int8
func Btoi8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

// I8tob converts int to bool
func I8tob(n int8) bool {
	return n != 0
}

// NewGameState creates a game state
func NewGameState(width int32, height int32) *GameState {
	return &GameState{
		Ball:      initialBall,
		BallSpeed: initialBallSpeed,
		LeftPlayerPaddle: Paddle{
			X:    paddleMargin,
			Y:    height / 2,
			Size: initialPaddleSize,
		},
		RightPlayerPaddle: Paddle{
			X:    width - paddleMargin - 1,
			Y:    height / 2,
			Size: initialPaddleSize,
		},
		Height: height,
		Width:  width,
		Tick:   initialTick,
		Debug:  Btoi8(debugFlag),
		Mode:   GameModePlaying,
	}
}

// EnterMode changes the game mode
func (g *GameState) EnterMode(mode GameMode) {
	switch mode {
	case GameModePlaying:
		g.Mode = GameModePlaying
		g.ModeTimeOut = 0
		g.Tick = initialTick
	case GameModeLeftPlayerLoosed:
		g.Mode = GameModeLeftPlayerLoosed
		g.ModeTimeOut = playerLooseModeTimeOut
		g.Tick = initialTick

		if g.LeftPlayerPaddle.Size < 10 {
			g.LeftPlayerPaddle.Size++

			if g.RightPlayerPaddle.Size > minPaddleSize {
				g.RightPlayerPaddle.Size--
			}
		}
	case GameModeRightPlayerLoosed:
		g.Mode = GameModeRightPlayerLoosed
		g.ModeTimeOut = playerLooseModeTimeOut
		g.Tick = initialTick

		if g.RightPlayerPaddle.Size < 10 {
			g.RightPlayerPaddle.Size++

			if g.LeftPlayerPaddle.Size > minPaddleSize {
				g.LeftPlayerPaddle.Size--
			}
		}
	}
}

// Update updates the game state
func (g *GameState) Update() {
	switch g.Mode {
	case GameModePlaying:
		g.Ball.X += g.BallSpeed.X
		g.Ball.Y += g.BallSpeed.Y

		if g.Ball.X <= g.LeftPlayerPaddle.X {
			if g.Ball.Y+g.Ball.Height-1 > g.LeftPlayerPaddle.Y-g.LeftPlayerPaddle.Size/2 &&
				g.Ball.Y < g.LeftPlayerPaddle.Y+g.LeftPlayerPaddle.Size/2 {

			} else {
				g.RightPlayerScore++
				g.Ball.Y = g.LeftPlayerPaddle.Y
				g.EnterMode(GameModeLeftPlayerLoosed)
			}

			g.Ball.X = g.LeftPlayerPaddle.X + 1
			g.BallSpeed.X *= -1

		} else if g.Ball.X+g.Ball.Width >= g.RightPlayerPaddle.X {
			if g.Ball.Y+g.Ball.Height-1 > g.RightPlayerPaddle.Y-g.RightPlayerPaddle.Size/2 &&
				g.Ball.Y < g.RightPlayerPaddle.Y+g.RightPlayerPaddle.Size/2 {

			} else {
				g.LeftPlayerScore++
				g.Ball.Y = g.RightPlayerPaddle.Y
				g.EnterMode(GameModeRightPlayerLoosed)
			}

			g.Ball.X = g.RightPlayerPaddle.X - g.Ball.Width
			g.BallSpeed.X *= -1
		}

		if g.Ball.Y <= 0 {
			g.Ball.Y = 0
			g.BallSpeed.Y *= -1
		} else if g.Ball.Y+g.Ball.Height >= g.Height {
			g.Ball.Y = g.Height - g.Ball.Height
			g.BallSpeed.Y *= -1
		}

		g.Tick = g.Tick*99/100 + .35
	case GameModeLeftPlayerLoosed:
		g.Ball.Y = g.LeftPlayerPaddle.Y

		g.ModeTimeOut--
		if g.ModeTimeOut < 0 {
			g.EnterMode(GameModePlaying)
		}

	case GameModeRightPlayerLoosed:
		g.Ball.Y = g.RightPlayerPaddle.Y

		g.ModeTimeOut--
		if g.ModeTimeOut < 0 {
			g.EnterMode(GameModePlaying)
		}
	}
}

// Draw draws ball on screen
func (b *Ball) Draw(viewSettings *ViewSettings) {
	for i := int32(0); i < b.Width; i++ {
		for j := int32(0); j < b.Height; j++ {
			termbox.SetCell(int(b.X-viewSettings.XOffset+i), int(b.Y+j), 'O', termbox.ColorRed, termbox.ColorRed)
		}
	}
}

// Clear clears ball on screen
func (b *Ball) Clear(viewSettings *ViewSettings) {
	for i := int32(0); i < b.Width; i++ {
		for j := int32(0); j < b.Height; j++ {
			termbox.SetCell(int(b.X-viewSettings.XOffset+i), int(b.Y+j), ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

// Draw draws paddle on screen
func (p *Paddle) Draw(viewSettings *ViewSettings) {
	x := p.X - viewSettings.XOffset
	for y := p.Y - p.Size/2; y < p.Size/2+p.Y; y++ {
		termbox.SetCell(int(x), int(y), '#', termbox.ColorWhite, termbox.ColorWhite)
	}
}

// Clear clears paddle on screen
func (p *Paddle) Clear(viewSettings *ViewSettings) {
	x := p.X - viewSettings.XOffset
	for y := p.Y - p.Size/2; y < p.Size/2+p.Y; y++ {
		termbox.SetCell(int(x), int(y), ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
}

// DrawDegug prints debug info on screen
func (g *GameState) DrawDegug(viewsettings *ViewSettings) {
	printText(fmt.Sprint(g, "       "), 2, 2)
	printText(fmt.Sprint(viewsettings, "      "), 2, 3)
}

// Draw draws game on screen
func (g *GameState) Draw(viewSettings *ViewSettings) {
	g.DrawPlayerScore()
	if I8tob(g.Debug) {
		g.DrawDegug(viewSettings)
	}
	g.Ball.Draw(viewSettings)
	g.LeftPlayerPaddle.Draw(viewSettings)
	g.RightPlayerPaddle.Draw(viewSettings)
}

// Clear clears game on screen
func (g *GameState) Clear(viewSettings *ViewSettings) {
	g.Ball.Clear(viewSettings)
	g.LeftPlayerPaddle.Clear(viewSettings)
	g.RightPlayerPaddle.Clear(viewSettings)
}

// MovePaddle moves the paddle according de `paddleMovement`
func (g *GameState) MovePaddle(paddleMovement *PaddleMovement) {
	switch paddleMovement.Player {
	case MoveLeftPaddle:
		g.LeftPlayerPaddle.RelaviteMovePaddle(paddleMovement.RelativeY, g.Height)
	case MoveRightPaddle:
		g.RightPlayerPaddle.RelaviteMovePaddle(paddleMovement.RelativeY, g.Height)
	}

	if g.Mode == GameModeLeftPlayerLoosed || g.Mode == GameModeRightPlayerLoosed {
		g.ModeTimeOut++
		g.Update()
	}
}

// RelaviteMovePaddle moves the paddle by relavtiveY
func (p *Paddle) RelaviteMovePaddle(relavtiveY int32, gameHeight int32) {
	p.Y += relavtiveY
	if p.Y < p.Size/2 {
		p.Y = p.Size / 2
	} else if p.Y > gameHeight-p.Size/2 {
		p.Y = gameHeight - p.Size/2
	}
}

func printText(s string, x int, y int) {
	for _, r := range s {
		termbox.SetCell(x, y, r, termbox.ColorDefault, termbox.ColorDefault)

		w := runewidth.RuneWidth(r)
		if w == 0 || (w == 2 && runewidth.IsAmbiguousWidth(r)) {
			w = 1
		}
		x += w
	}
}

// DrawPlayerScore draws players score
func (g *GameState) DrawPlayerScore() {
	printText(fmt.Sprint(g.LeftPlayerScore, " - ", g.RightPlayerScore), 2, 0)
}
