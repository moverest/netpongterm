package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nsf/termbox-go"
)

func clientMain(serverAddrFlag string, lastClient bool) {
	client := NewClient(lastClient)

	err := client.InitGraphics()
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(serverAddrFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect()

	err = client.SendWindowSettings()
	if err != nil {
		log.Fatal(err)
	}

	client.WaitViewSettings()

	go client.EventLoop()
	client.LoopGameUpdate()

	fmt.Println(client)
}

// Client defines client object
type Client struct {
	Conn           *net.UDPConn
	WindowSettings WindowSettings // Set by client
	ViewSettings   ViewSettings   // Set by server
	GameState      GameState
}

// NewClient creates a new client object
func NewClient(last bool) *Client {
	return &Client{
		WindowSettings: WindowSettings{
			Last: Btoi8(last),
		},
	}
}

// InitGraphics initializes graphics
func (c *Client) InitGraphics() error {
	err := termbox.Init()
	if err != nil {
		return err
	}

	width, height := termbox.Size()
	c.WindowSettings.Width, c.WindowSettings.Height = int32(width), int32(height)

	return nil
}

// QuitGraphics stops graphics
func (c *Client) QuitGraphics() {
	termbox.Close()
}

// Connect method connects to the server
func (c *Client) Connect(serverAddr string) error {
	serverUDPAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return err
	}

	c.Conn, err = net.DialUDP("udp", nil, serverUDPAddr)
	if err != nil {
		return err
	}

	return nil
}

// Disconnect diconnects from the server
func (c *Client) Disconnect() error {
	return c.Conn.Close()
}

// SendWindowSettings sends window settings to server
func (c *Client) SendWindowSettings() error {
	return SendPacket(&c.WindowSettings, c.Conn, nil)
}

// WaitViewSettings waits for the server to send view settings and stores them
func (c *Client) WaitViewSettings() {
	for {
		_, _, packet, _ := ReadPacket(c.Conn)

		switch p := packet.(type) {
		case *ViewSettings:
			c.ViewSettings = *p
			return
		}
	}
}

// Update update screen
func (c *Client) Update(gameState *GameState) {
	c.GameState.Clear(&c.ViewSettings)
	c.GameState = *gameState
	c.GameState.Draw(&c.ViewSettings)
	termbox.Flush()
}

// LoopGameUpdate wait for the server for new game state
func (c *Client) LoopGameUpdate() {
	for {
		_, _, packet, _ := ReadPacket(c.Conn)

		switch p := packet.(type) {
		case *GameState:
			c.Update(p)
		}
	}
}

// SendPaddleMovement sends paddle movement to server
func (c *Client) SendPaddleMovement(relativeY int32) error {
	paddleMovement := &PaddleMovement{
		RelativeY: relativeY,
		Player:    c.ViewSettings.ControlledPaddle,
	}

	err := SendPacket(paddleMovement, c.Conn, nil)
	if err != nil {
		return err
	}

	return nil
}

// EventLoop catches events and process them
func (c *Client) EventLoop() error {
	for {
		event := termbox.PollEvent()
		if event.Type == termbox.EventKey {
			switch event.Key {
			case termbox.KeyArrowUp:
				err := c.SendPaddleMovement(-1)
				if err != nil {
					return err
				}
			case termbox.KeyArrowDown:
				err := c.SendPaddleMovement(1)
				if err != nil {
					return err
				}
			}
		}
	}
}
