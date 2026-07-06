package main

import (
	"fmt"
	"io"
	"log"
	"net"

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
	for {
		chunk, err := reader.NextChunk()
		if err == io.EOF {
			fmt.Println("File finished")
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		packet := &protocol.Packet{
			Flags:   protocol.FlagData,
			Payload: chunk,
		}
		data, err := packet.Marshal()
		fmt.Printf("Sent packet (%d bytes)\n", len(data))
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Error sending packet: ", err)
			continue
		}
	}
}
