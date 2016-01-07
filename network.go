package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

// MagicKey is used to check packet protocol
const MagicKey = int64(583276045987)

// Packet Types
const (
	WindowSettingsPacket = iota
	ViewSettingsPacket   = iota
	GameStatePacket      = iota
	PaddleMovementPacket = iota
)

// Errors
var (
	ErrUnknowPacketType = errors.New("unknown packet type error")
	ErrInvalidMagicKey  = errors.New("invalid magic key")
)

// Packet interface
type Packet interface{}

// PacketHeader stores the magic key used to check protocol and packet type
type PacketHeader struct {
	MagicKey int64
	Type     int8
}

// WindowSettings stores client window size
type WindowSettings struct {
	Width  int32
	Height int32
	Last   int8
}

// ViewSettings stores client view settings
type ViewSettings struct {
	XOffset          int32
	Height           int32
	Width            int32
	ControlledPaddle int8
}

// ReadPacket waits for a new UDP packet and decode it
func ReadPacket(conn *net.UDPConn) (*PacketHeader, *net.UDPAddr, Packet, error) {
	buf := make([]byte, 1024)

	_, senderAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, nil, nil, err
	}

	reader := bytes.NewReader(buf)
	packetHeader := &PacketHeader{}

	err = binary.Read(reader, binary.LittleEndian, packetHeader)
	if err != nil {
		return nil, nil, nil, err
	}

	err = packetHeader.Check()
	if err != nil {
		return nil, nil, nil, err
	}

	switch packetHeader.Type {
	case WindowSettingsPacket:
		winSettings := &WindowSettings{}
		err = binary.Read(reader, binary.LittleEndian, winSettings)
		if err != nil {
			return nil, nil, nil, err
		}
		return packetHeader, senderAddr, winSettings, nil

	case ViewSettingsPacket:
		viewSettings := &ViewSettings{}
		err = binary.Read(reader, binary.LittleEndian, viewSettings)
		if err != nil {
			return nil, nil, nil, err
		}

		return packetHeader, senderAddr, viewSettings, nil

	case GameStatePacket:
		gameState := &GameState{}
		err = binary.Read(reader, binary.LittleEndian, gameState)
		if err != nil {
			return nil, nil, nil, err
		}

		return packetHeader, senderAddr, gameState, nil

	case PaddleMovementPacket:
		paddleMovement := &PaddleMovement{}
		err = binary.Read(reader, binary.LittleEndian, paddleMovement)
		if err != nil {
			return nil, nil, nil, err
		}

		return packetHeader, senderAddr, paddleMovement, nil
	default:
		return nil, nil, nil, ErrUnknowPacketType
	}
}

// SendPacket sends UDP packet
func SendPacket(packet Packet, conn *net.UDPConn, addr *net.UDPAddr) error {

	if packet == nil {
		return ErrUnknowPacketType
	}

	packetHeader := &PacketHeader{
		MagicKey: MagicKey,
	}

	switch packet.(type) {
	case *WindowSettings:
		packetHeader.Type = WindowSettingsPacket
	case *ViewSettings:
		packetHeader.Type = ViewSettingsPacket
	case *GameState:
		packetHeader.Type = GameStatePacket
	case *PaddleMovement:
		packetHeader.Type = PaddleMovementPacket
	default:
		return ErrUnknowPacketType
	}

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, packetHeader)
	if err != nil {
		return err
	}

	err = binary.Write(buf, binary.LittleEndian, packet)
	if err != nil {
		return err
	}

	if addr == nil {
		conn.Write(buf.Bytes())
	} else {
		_, err = conn.WriteTo(buf.Bytes(), addr)
		if err != nil {
			return err
		}
	}

	return nil
}

// Check returns nil if h is a valid packet header
func (h *PacketHeader) Check() error {
	if h.MagicKey != MagicKey {
		return ErrInvalidMagicKey
	}

	return nil
}
