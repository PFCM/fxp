package midi

import (
	"errors"
	"fmt"
)

//go:generate stringer -type=MessageType,CV1MessageType

// MessageType is a UMP message type, a group of types of message.
type MessageType byte

const (
	MTUtility       MessageType = 0x0
	MTSystem        MessageType = 0x1
	MTChannelVoice1 MessageType = 0x2
	MTData          MessageType = 0x3
	MTChannelVoice2 MessageType = 0x4
	MTLongData      MessageType = 0x5
	// several reserved.
	MTFlexData MessageType = 0xD
	// 0xE is reserved
	MTUMPStream MessageType = 0xF
)

// messageTypeSizes is size in uint32s of each type of message.
var messageTypeSizes = []int{
	MTUtility:       1,
	MTSystem:        1,
	MTChannelVoice1: 1,
	MTData:          2,
	MTChannelVoice2: 2,
	MTLongData:      4,
	MTFlexData:      4,
	MTUMPStream:     4,
}

type Message struct {
	Type  MessageType
	Group byte
	// fields for 1.0 Channel Voice messages.
	CV1Type CV1MessageType
	Channel byte
	// MIDI note for note on/note off/poly pressure, but also
	// index for control change and program for program change.
	Note      byte
	Velocity  byte // for note {on, off}, {poly,channel} pressure.
	PitchBend uint16
}

// CV1MessageType is the type of a 1.0 Channel Voice message. Also the high 4
// bits of the first actual message byte (and the high 4 bits of the classic
// format, as they all have the first bit set).
type CV1MessageType byte

const (
	CV1NoteOff = CV1MessageType(0x8 | byte(iota))
	CV1NoteOn
	CV1PolyPressure
	CV1ControlChange
	CV1ProgramChange
	CV1ChannelPressure
	CV1PitchBend
)

func parseChannelVoice1(raw []uint32) (Message, []uint32, error) {
	// This message is only 32 bits.
	p, raw := raw[0], raw[1:]
	// Group is second-most significant set of 4 bits.
	g := byte(p>>24) & 0xF
	// The remaining 3 bytes are more or less the traditional bytes from the
	// old format.
	msg := Message{
		Type:    MTChannelVoice1,
		Group:   g,
		CV1Type: CV1MessageType((p >> 20) & 0xF),
		Channel: byte((p >> 16) & 0xF),
	}
	switch msg.CV1Type {
	case CV1NoteOff, CV1NoteOn, CV1PolyPressure, CV1ControlChange:
		// a byte of note, and a byte of velocity. High bit
		// _should_ be zero.
		msg.Note = byte(p>>8) & 0x7F
		msg.Velocity = byte(p) & 0x7F
	case CV1ProgramChange:
		msg.Note = byte(p>>8) & 0x7F
	case CV1ChannelPressure:
		msg.Velocity = byte(p>>8) & 0x7F
	case CV1PitchBend:
		low := uint16(p>>8) & 0x7F
		high := uint16(p) & 0x7F
		msg.PitchBend = (high << 7) | low
	default:
		return msg, nil, fmt.Errorf("invalid 1.0 Channel Voice message type: %d", msg.CV1Type)
	}
	return msg, raw, nil
}

// ParseMessage parses a single (possibly variable-length) UMP message from a
// slice of raw data. Returns the original slice, advanced to the start of the
// next message (or the end).
func ParseMessage(raw []uint32) (Message, []uint32, error) {
	if len(raw) == 0 {
		return Message{}, nil, errors.New("no input")
	}
	// The type is always the most significant 4 bits.
	t := MessageType(raw[0] >> 28)
	switch t {
	case MTChannelVoice1:
		return parseChannelVoice1(raw)
	}
	return Message{}, nil, errors.New("not implmemented")
}

// ParseMessages calls ParseMesage until the input is exhausted.
func ParseMessages(raw []uint32) ([]Message, error) {
	var messages []Message
	for len(raw) > 0 {
		msg, next, err := ParseMessage(raw)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
		raw = next
	}
	return messages, nil
}
