package protocol

import (
	"bytes"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	original := &Packet{
		SeqNum:   42,
		AckNum:   17,
		Flags:    FlagData,
		Checksum: 12345,
		Payload:  []byte("Hello UDP"),
	}

	data := original.Marshal()

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
