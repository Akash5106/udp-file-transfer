package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Akash5106/udp-file-transfer/internal/file"
	"github.com/Akash5106/udp-file-transfer/internal/transport"
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
	sender := transport.NewSender(conn, 4)
	err = sender.SendFile(reader)
	if err != nil {
		log.Fatal(err)
	}
}
