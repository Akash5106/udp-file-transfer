package protocol

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

const HeaderSize = 15

var ErrChecksumMismatch = errors.New("checksum mismatch")

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

func NewDataPacket(seq uint32, payload []byte) *Packet {
	return &Packet{
		SeqNum:  seq,
		Flags:   FlagData,
		Payload: payload,
	}
}

func NewACKPacket(ack uint32) *Packet {
	return &Packet{
		AckNum: ack,
		Flags:  FlagACK,
	}
}

func NewFINPacket() *Packet {
	return &Packet{
		Flags: FlagFIN,
	}
}

func computeChecksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

func (p *Packet) Marshal() ([]byte, error) {
	p.Length = uint16(len(p.Payload))

	buffer := make([]byte, HeaderSize+len(p.Payload))

	binary.BigEndian.PutUint32(buffer[0:4], p.SeqNum)
	binary.BigEndian.PutUint32(buffer[4:8], p.AckNum)
	buffer[8] = p.Flags
	binary.BigEndian.PutUint16(buffer[9:11], p.Length)
	binary.BigEndian.PutUint32(buffer[11:15], 0)
	copy(buffer[HeaderSize:], p.Payload)
	checksum := computeChecksum(buffer)
	p.Checksum = checksum
	binary.BigEndian.PutUint32(buffer[11:15], p.Checksum)
	return buffer, nil
}

func Unmarshal(buffer []byte) (*Packet, error) {
	if len(buffer) < HeaderSize {
		return nil, errors.New("packet too small")
	}

	storedChecksum := binary.BigEndian.Uint32(buffer[11:15])
	binary.BigEndian.PutUint32(buffer[11:15], 0)
	computedChecksum := computeChecksum(buffer)
	binary.BigEndian.PutUint32(buffer[11:15], storedChecksum)
	if computedChecksum != storedChecksum {
		return nil, ErrChecksumMismatch
	}

	packet := &Packet{}

	packet.SeqNum = binary.BigEndian.Uint32(buffer[0:4])
	packet.AckNum = binary.BigEndian.Uint32(buffer[4:8])
	packet.Flags = buffer[8]
	packet.Length = binary.BigEndian.Uint16(buffer[9:11])
	packet.Checksum = storedChecksum

	if len(buffer) < HeaderSize+int(packet.Length) {
		return nil, errors.New("truncated packet")
	}

	packet.Payload = make([]byte, packet.Length)
	copy(packet.Payload, buffer[HeaderSize:HeaderSize+int(packet.Length)])

	return packet, nil
}
