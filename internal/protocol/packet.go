package protocol

import (
	"encoding/binary"
	"errors"
)

const HeaderSize = 15

type Packet struct {
	SeqNum   uint32
	AckNum   uint32
	Flags    uint8
	Length   uint16
	Checksum uint32
	Payload  []byte
}

const (
	FlagData uint8 = 1 << iota
	FlagACK
	FlagFIN
)

func (p *Packet) Marshal() []byte {
	p.Length = uint16(len(p.Payload))

	buffer := make([]byte, HeaderSize+len(p.Payload))

	binary.BigEndian.PutUint32(buffer[0:4], p.SeqNum)
	binary.BigEndian.PutUint32(buffer[4:8], p.AckNum)
	buffer[8] = p.Flags
	binary.BigEndian.PutUint16(buffer[9:11], p.Length)
	binary.BigEndian.PutUint32(buffer[11:15], p.Checksum)

	copy(buffer[HeaderSize:], p.Payload)

	return buffer
}

func Unmarshal(buffer []byte) (*Packet, error) {
	if len(buffer) < HeaderSize {
		return nil, errors.New("packet too small")
	}

	packet := &Packet{}

	packet.SeqNum = binary.BigEndian.Uint32(buffer[0:4])
	packet.AckNum = binary.BigEndian.Uint32(buffer[4:8])
	packet.Flags = buffer[8]
	packet.Length = binary.BigEndian.Uint16(buffer[9:11])
	packet.Checksum = binary.BigEndian.Uint32(buffer[11:15])

	if len(buffer) < HeaderSize+int(packet.Length) {
		return nil, errors.New("truncated packet")
	}

	packet.Payload = make([]byte, packet.Length)
	copy(packet.Payload, buffer[HeaderSize:HeaderSize+int(packet.Length)])

	return packet, nil
}
