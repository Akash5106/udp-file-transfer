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
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error creating socket:", err)
		return
	}
	defer conn.Close()

	writer, err := file.NewWriter("answer.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer writer.Close()
	receiver := transport.NewReceiver(conn, 4)
	err = receiver.ReceiveFile(writer)
	if err != nil {
		log.Fatal(err)
	}
}
