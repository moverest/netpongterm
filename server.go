package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func serverMain(serverFlag string) {
	server, err := NewServer(serverFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	server.WaitForNewClients()

	server.GenerateViewSettings()

	err = server.SendViewSettings()
	if err != nil {
		log.Fatal(err)
	}

	server.InitGame()

	go server.EventLoop()
	err = server.LoopGameUpdate()
	if err != nil {
		log.Fatal(err)
	}
}

// Event is the event interface
type Event interface{}

// Server defines server object
type Server struct {
	Clients   []ClientConn
	Conn      *net.UDPConn
	GameState *GameState
	EventChan chan Event
}

// ClientConn stores client address and settings
type ClientConn struct {
	Addr           *net.UDPAddr
	WindowSettings WindowSettings
	ViewSettings   ViewSettings
}

// NewServer creates a new server instance
func NewServer(serverAddr string) (*Server, error) {
	serverUDPAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, err
	}

	server := &Server{}

	server.Conn, err = net.ListenUDP("udp", serverUDPAddr)
	if err != nil {
		return nil, err
	}

	server.EventChan = make(chan Event)

	return server, nil
}

// Close closes the connexions
func (s *Server) Close() {
	s.Conn.Close()
}

// WaitForNewClients waits for new clients to connect until the last client connects
func (s *Server) WaitForNewClients() {
	log.Output(1, "Waiting for new clients")

	for {
		_, clientAdrr, packet, err := ReadPacket(s.Conn)
		if err != nil {
			log.Output(1, err.Error())
			break
		}

		switch p := packet.(type) {
		case *WindowSettings:
			client := ClientConn{
				Addr:           clientAdrr,
				WindowSettings: *p,
			}

			s.Clients = append(s.Clients, client)
			log.Output(1, fmt.Sprintln("Client added: ", client))

			if I8tob(p.Last) {
				log.Output(1, "Last client added")
				return
			}
		}
	}
}

// GenerateViewSettings computes view settings for all clients
func (s *Server) GenerateViewSettings() {
	xOffset := int32(0)
	height := s.Clients[0].WindowSettings.Height

	for i := range s.Clients {
		if height > s.Clients[i].WindowSettings.Height {
			height = s.Clients[i].WindowSettings.Height
		}

		s.Clients[i].ViewSettings.XOffset = xOffset
		xOffset += s.Clients[i].WindowSettings.Width
	}

	for i := range s.Clients {
		s.Clients[i].ViewSettings.Height = height
		s.Clients[i].ViewSettings.Width = xOffset
		s.Clients[i].ViewSettings.ControlledPaddle = MoveNonePaddle
	}

	s.Clients[len(s.Clients)-1].ViewSettings.ControlledPaddle = MoveRightPaddle
	s.Clients[0].ViewSettings.ControlledPaddle = MoveLeftPaddle
}

// SendViewSettings sends view settings to all clients
func (s *Server) SendViewSettings() error {
	for i := range s.Clients {

		err := SendPacket(&s.Clients[i].ViewSettings, s.Conn, s.Clients[i].Addr)
		if err != nil {
			return err
		}

		log.Output(1, fmt.Sprint("View settings sent: ", s.Clients[i]))
	}

	return nil
}

// InitGame initializes the game
func (s *Server) InitGame() {
	s.GameState = NewGameState(
		s.Clients[0].ViewSettings.Width,
		s.Clients[0].ViewSettings.Height,
	)
}

// GameUpdate update the game state
func (s *Server) GameUpdate() {
	s.GameState.Update()
}

// SendGameState sends the game state to all clients
func (s *Server) SendGameState() error {
	// log.Output(1, fmt.Sprint(s.GameState))

	for i := range s.Clients {
		err := SendPacket(s.GameState, s.Conn, s.Clients[i].Addr)
		if err != nil {
			return err
		}
	}

	return nil
}

// LoopGameUpdate is the game loop
func (s *Server) LoopGameUpdate() error {
	tick := int32(s.GameState.Tick)
	ticker := time.NewTicker(time.Millisecond * time.Duration(tick))

	for {
		select {
		case <-ticker.C:
			s.GameUpdate()

			err := s.SendGameState()
			if err != nil {
				return err
			}

			if int32(s.GameState.Tick) != tick {
				ticker.Stop()
				tick = int32(s.GameState.Tick)
				ticker = time.NewTicker(time.Millisecond * time.Duration(tick))
			}
		case event := <-s.EventChan:
			switch p := event.(type) {
			case *PaddleMovement:
				s.GameState.MovePaddle(p)
				err := s.SendGameState()
				if err != nil {
					return err
				}
			}
		}

	}
}

// EventLoop loop event sent from clients and process them
func (s *Server) EventLoop() {
	for {
		_, _, packet, err := ReadPacket(s.Conn)
		if err != nil {
			break
		}

		switch p := packet.(type) {
		case *PaddleMovement:
			s.EventChan <- p
		}
	}
}
