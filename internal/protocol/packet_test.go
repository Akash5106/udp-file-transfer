package protocol

import (
	"bytes"
	"errors"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	original := &Packet{
		SeqNum:  42,
		AckNum:  17,
		Flags:   FlagData,
		Payload: []byte("Hello UDP"),
	}

	data, err := original.Marshal()

	decoded, err := Unmarshal(data)
	if err != nil {
		t.Fatal(err)
	}

	if original.SeqNum != decoded.SeqNum {
		t.Errorf("SeqNum mismatch")
	}

	if original.AckNum != decoded.AckNum {
		t.Errorf("AckNum mismatch")
	}

	if original.Flags != decoded.Flags {
		t.Errorf("Flags mismatch")
	}

	if original.Length != decoded.Length {
		t.Errorf("Length mismatch")
	}

	if original.Checksum != decoded.Checksum {
		t.Errorf("Checksum mismatch")
	}

	if !bytes.Equal(original.Payload, decoded.Payload) {
		t.Errorf("Payload mismatch")
	}

}

func TestChecksumMismatch(t *testing.T) {
	packet := &Packet{
		SeqNum:  1,
		AckNum:  0,
		Flags:   FlagData,
		Payload: []byte("Hello World"),
	}

	data, err := packet.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt one byte in the payload.
	data[HeaderSize] ^= 0xFF

	_, err = Unmarshal(data)

	if !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("expected checksum mismatch, got %v", err)
	}
}
