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
