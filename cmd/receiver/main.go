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
	// fmt.Println(addr)
	// buffer := make([]byte, 1024)
	// for {
	// 	n, addr, err := conn.ReadFromUDP(buffer)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		continue
	// 	}
	// 	fmt.Println("Received", n, "bytes of data from", addr)
	// 	fmt.Println(string(buffer[:n]))
	// 	_, err = conn.WriteToUDP(buffer[:n], addr)
	// 	if err != nil {
	// 		fmt.Println("Error sending mssg")
	// 		continue
	// 	}
	// 	fmt.Println("ACK sent")
	// }

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
