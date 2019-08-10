package main

import (
	"encoding/binary"
	"fmt"
	"github.com/proto-lite/model"
	"github.com/proto-lite/model/frame"
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

type Client struct {
	SenderCh chan model.Packet
	ReceiverCh chan []byte
	conn net.Conn
	packets model.Packets
}

func NewClient(rawSrc []byte, remoteAddr net.Addr) *Client {
	conn, err := net.Dial("udp4", "localhost:8888")
	packets := model.Packets{}
	if err != nil {
		panic(err)
	}
	client :=  &Client{
		make(chan model.Packet),
		make(chan []byte),
		conn,
		packets,
	}
	go client.sendAsync(client.SenderCh)
	go client.recvPacket(client.ReceiverCh)
	go client.recv()
	return client
}


func (self *Client)Close(){
	self.conn.Close()
}

func (self *Client)recv() {
	for{
		ret :=make([]byte, model.GetPacketByteLength())
		self.conn.Read(ret)
		self.ReceiverCh <- ret
	}
}

func (self *Client)recvPacket(ch <- chan []byte) {
	for{
		i := <- ch
		packet := self.packets.AddPacketFromReceiveByte(i, self.conn.RemoteAddr())
		if(model.DataFrameType.GetByte() == model.GetFrameTypeFromRawData(i)){

		}else if(model.AckFrameType.GetByte() == model.GetFrameTypeFromRawData(i)){
			ackFrame := frame.NewAckFromBinary(packet.FrameData)
			lossPackets, acPackets := ackFrame.GetLossAndAcceptedPacketIDs()
			self.packets.AddLossPacketIDs(lossPackets)
			self.packets.AddAcceptPacketIDs(acPackets)
		}
	}
}

func (self *Client)resendLossPackets(){
	for _,i := range self.packets.GetLossPacketIDs(){
		self.SenderCh <- self.packets.AddResendPacket(self.packets.Packets[i])
	}
}

func (self *Client)send(packet model.Packet){
	self.SenderCh <- packet
}

func (self *Client)sendAsync(ch <-chan model.Packet)  {
	for {
		i := <- ch
		_, err := self.conn.Write(i.ToBytes())
		if err != nil {
			panic(err)
		}
	}
}

