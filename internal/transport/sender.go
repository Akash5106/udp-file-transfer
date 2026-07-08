package transport

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

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
	timers     map[uint32]*time.Timer
	windowSize uint32
}

func NewSender(conn *net.UDPConn, windowSize uint32) *Sender {
	sender := &Sender{
		conn:       conn,
		base:       0,
		nextSeq:    0,
		packets:    make(map[uint32]*protocol.Packet),
		acked:      make(map[uint32]bool),
		timers:     make(map[uint32]*time.Timer),
		windowSize: windowSize,
	}
	sender.cond = sync.NewCond(&sender.mu)
	return sender
}

func (s *Sender) sendPacket(packet *protocol.Packet) error {
	data, err := packet.Marshal()
	if err != nil {
		return err
	}

	fmt.Printf("Sending packet %d\n", packet.SeqNum)

	_, err = s.conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sender) startTimer(seq uint32) {
	timer := time.AfterFunc(2*time.Second, func() {
		s.handleTimeout(seq)
	})

	s.mu.Lock()
	s.timers[seq] = timer
	s.mu.Unlock()
}

func (s *Sender) handleTimeout(seq uint32) {
	s.mu.Lock()
	p, exists := s.packets[seq]
	if !exists {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	if err := s.sendPacket(p); err != nil {
		fmt.Println("Retransmission failed:", err)
		return
	}
	s.mu.Lock()
	_, exists = s.packets[seq]
	s.mu.Unlock()

	if exists {
		s.startTimer(seq)
	}
	fmt.Printf("Retransmitted packet %d\n", seq)
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
		err = s.sendPacket(packet)
		if err != nil {
			return err
		}
		s.startTimer(seq)
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
		timer, exists := s.timers[ack.AckNum]
		if exists {
			timer.Stop()
			delete(s.timers, ack.AckNum)
		}
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
