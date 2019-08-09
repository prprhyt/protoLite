package main

import (
	"github.com/proto-lite/model"
	"net"
)
func main() {
	packets := model.NewPackets()
	Ch := make(chan model.Packet)
	go send(Ch)
	data := []byte("Hello from Server")
	packets.AddNewDataPacket(data)
	Ch <- packets.GetLatestPacket()
	/*
	buffer := make([]byte, 1500)
	length, err := conn.Read(buffer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received: %s\n", string(buffer[:length]))
	*/
}

func send(ch <-chan model.Packet)  {
	conn, err := net.Dial("udp4", "localhost:8888")
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	for {
		i := <- ch
		_, err = conn.Write(i.ToBytes())
		if err != nil {
			panic(err)
		}
	}
}

