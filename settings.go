package main

var (
	debugFlag = false
)

// Game settings
const (
	paddleMargin           = 1
	initialTick            = 70
	playerLooseModeTimeOut = 12

	maxPaddleSize     = 10
	minPaddleSize     = 4
	initialPaddleSize = 6
)

var (
	initialBallSpeed = BallSpeed{X: 1, Y: 1}
	initialBall      = Ball{X: 5, Y: 5, Width: 2, Height: 1}
)

// Default command line flags
const (
	defaultModeFlag       = "client"
	defaultServerFlag     = "0.0.0.0:5454"
	defaultLastClientFlag = false
)
