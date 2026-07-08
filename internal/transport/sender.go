package transport

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/Akash5106/udp-file-transfer/internal/file"
	"github.com/Akash5106/udp-file-transfer/internal/protocol"
)

type Sender struct {
	conn       *net.UDPConn
	base       uint32
	nextSeq    uint32
	packets    map[uint32]*protocol.Packet
	acked      map[uint32]bool
	mu         sync.Mutex
	cond       *sync.Cond
	windowSize uint32
}

func NewSender(conn *net.UDPConn, windowSize uint32) *Sender {
	sender := &Sender{
		conn:       conn,
		base:       0,
		nextSeq:    0,
		packets:    make(map[uint32]*protocol.Packet),
		acked:      make(map[uint32]bool),
		windowSize: windowSize,
	}
	sender.cond = sync.NewCond(&sender.mu)
	return sender
}

func (s *Sender) SendFile(reader *file.Reader) error {
	go s.receiveACKs()
	finishedReading := false
	for {
		if finishedReading {
			s.mu.Lock()
			for s.base != s.nextSeq {
				s.cond.Wait()
			}
			s.mu.Unlock()
			fmt.Println("Transfer complete")
			return nil
		}
		chunk, err := reader.NextChunk()
		if err == io.EOF {
			fmt.Println("File finished")
			finishedReading = true
			continue
		}
		if err != nil {
			return err
		}
		s.mu.Lock()

		for s.nextSeq >= s.base+s.windowSize {
			s.cond.Wait()
		}

		seq := s.nextSeq
		s.nextSeq++

		s.mu.Unlock()
		packet := protocol.NewDataPacket(seq, chunk)
		s.mu.Lock()
		s.packets[packet.SeqNum] = packet
		s.mu.Unlock()
		data, err := packet.Marshal()
		if err != nil {
			return err
		}
		fmt.Printf("Sending packet %d\n", seq)
		_, err = s.conn.Write(data)
		if err != nil {
			return err
		}
	}
}

func (s *Sender) receiveACKs() {
	buffer := make([]byte, 1024)
	for {
		n, err := s.conn.Read(buffer)
		if err != nil {
			return
		}
		ack, err := protocol.Unmarshal(buffer[:n])
		if err != nil {
			return
		}
		if ack.Flags != protocol.FlagACK {
			return
		}
		s.mu.Lock()
		s.acked[ack.AckNum] = true
		fmt.Printf("Marked ACK %d\n", ack.AckNum)
		windowMoved := false
		for s.acked[s.base] {
			delete(s.acked, s.base)
			delete(s.packets, s.base)
			s.base++
			windowMoved = true
		}
		if windowMoved {
			s.cond.Signal()
		}
		s.mu.Unlock()
	}
}
