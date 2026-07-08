package transport

import (
	"fmt"
	"net"

	"github.com/Akash5106/udp-file-transfer/internal/file"
	"github.com/Akash5106/udp-file-transfer/internal/protocol"
)

type Receiver struct {
	conn        *net.UDPConn
	expectedSeq uint32
	windowSize  uint32
}

func NewReceiver(conn *net.UDPConn, windowSize uint32) *Receiver {
	return &Receiver{
		conn:        conn,
		expectedSeq: 0,
		windowSize:  windowSize,
	}
}

func (r *Receiver) ReceiveFile(writer *file.Writer) error {
	buffer := make([]byte, 1500)
	for {
		n, addr, err := r.conn.ReadFromUDP(buffer)
		if err != nil {
			return err
		}

		packet, err := protocol.Unmarshal(buffer[:n])
		if err != nil {
			return err
		}

		if packet.SeqNum != r.expectedSeq {
			fmt.Printf("Duplicate packet %d received. Sending ACK again.\n", packet.SeqNum)
			ack := protocol.NewACKPacket(packet.SeqNum)
			ackData, err := ack.Marshal()
			if err != nil {
				return err
			}
			_, err = r.conn.WriteToUDP(ackData, addr)
			if err != nil {
				return err
			}
			continue
		}

		err = writer.WriteChunk(packet.Payload)
		if err != nil {
			return err
		}

		fmt.Printf(
			"Received packet %d (%d bytes)\n",
			packet.SeqNum,
			packet.Length,
		)
		r.expectedSeq++
		ack := protocol.NewACKPacket(packet.SeqNum)

		ackData, err := ack.Marshal()
		if err != nil {
			return err
		}

		_, err = r.conn.WriteToUDP(ackData, addr)
		if err != nil {
			return err
		}
		fmt.Printf("Sent ACK %d\n", ack.AckNum)
	}
}
