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
	buffer      map[uint32]*protocol.Packet
	windowSize  uint32
}

func NewReceiver(conn *net.UDPConn, windowSize uint32) *Receiver {
	return &Receiver{
		conn:        conn,
		expectedSeq: 0,
		buffer:      make(map[uint32]*protocol.Packet),
		windowSize:  windowSize,
	}
}

func (r *Receiver) inWindow(seq uint32) bool {
	return seq >= r.expectedSeq &&
		seq < r.expectedSeq+r.windowSize
}

func (r *Receiver) sendACK(seq uint32, addr *net.UDPAddr) error {
	ack := protocol.NewACKPacket(seq)

	ackData, err := ack.Marshal()
	if err != nil {
		return err
	}

	_, err = r.conn.WriteToUDP(ackData, addr)
	if err != nil {
		return err
	}
	return nil
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
		if packet.SeqNum < r.expectedSeq {
			fmt.Printf("Old packet %d received. Sending ACK again.\n", packet.SeqNum)

			err = r.sendACK(packet.SeqNum, addr)
			if err != nil {
				return err
			}

			continue
		}
		if !r.inWindow(packet.SeqNum) {
			fmt.Printf("Not in window, Ignoring packet : %d\n", packet.SeqNum)
			continue
		}

		_, exists := r.buffer[packet.SeqNum]
		if exists {
			fmt.Printf("Duplicate packet %d received\n", packet.SeqNum)
			err = r.sendACK(packet.SeqNum, addr)
			if err != nil {
				return err
			}
			continue
		}

		r.buffer[packet.SeqNum] = packet
		err = r.sendACK(packet.SeqNum, addr)
		if err != nil {
			return err
		}
		for {
			p, exists := r.buffer[r.expectedSeq]
			if !exists {
				fmt.Printf("Flush stopped. Waiting for packet %d\n", r.expectedSeq)
				break
			}
			fmt.Printf("Flushing packet %d\n", p.SeqNum)
			err = writer.WriteChunk(p.Payload)
			if err != nil {
				return err
			}
			delete(r.buffer, r.expectedSeq)
			fmt.Printf("Delivered packet : %d\n", p.SeqNum)
			r.expectedSeq++
		}
	}
}
