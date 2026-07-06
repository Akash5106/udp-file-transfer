package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/Akash5106/udp-file-transfer/internal/file"
	"github.com/Akash5106/udp-file-transfer/internal/protocol"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:5454")
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error creating connection:", err)
		return
	}
	defer conn.Close()

	// message := []byte("Hello UDP")

	// _, err = conn.Write(message)
	// if err != nil {
	// 	fmt.Println("Error sending message:", err)
	// 	return
	// }

	// fmt.Println("Message sent!")
	// buffer := make([]byte, 1024)
	// n, addr, err := conn.ReadFromUDP(buffer)
	// if err != nil {
	// 	fmt.Println("Error receiving ACK")
	// 	return
	// }
	// fmt.Println("ACK received from ", addr, " : ", string(buffer[:n]))

	reader, err := file.NewReader("sample.txt", 20)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	seq := uint32(0)
	buffer := make([]byte, 1024)
	for {
		chunk, err := reader.NextChunk()
		if err == io.EOF {
			fmt.Println("File finished")
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		packet := protocol.NewDataPacket(seq, chunk)
		data, err := packet.Marshal()
		if err != nil {
			log.Fatal(err)
		}
		for {
			fmt.Printf("Sending packet %d\n", packet.SeqNum)
			_, err = conn.Write(data)
			if err != nil {
				fmt.Println("Error sending packet: ", err)
				continue
			}
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, err := conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					fmt.Println("ACK timeout... retransmitting")
					continue
				}
				log.Fatal(err)
			}
			ack, err := protocol.Unmarshal(buffer[:n])
			if err != nil {
				log.Fatal(err)
			}
			if ack.Flags != protocol.FlagACK {
				log.Fatal("expected ACK packet")
			}
			if ack.AckNum != packet.SeqNum {
				fmt.Printf("Expected ACK %d but got ACK %d\n", packet.SeqNum, ack.AckNum)
				continue
			}
			fmt.Printf("Received ACK %d\n", ack.AckNum)
			break
		}
		seq++
	}
}
