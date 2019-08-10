package main

import (
	"github.com/proto-lite/model"
	"github.com/proto-lite/model/frame"
	"net"
)
func main() {
	client := NewClient("localhost:8888")
	client.Send([]byte("Hello World from Client"))
}

type Client struct {
	SenderCh chan model.Packet
	ReceiverCh chan []byte
	conn net.Conn
	packets model.Packets
}

func NewClient(dstAddressString string) *Client {
	conn, err := net.Dial("udp4", dstAddressString)
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
		packet := self.packets.AddPacketFromReceiveByte(i, self.conn.LocalAddr(), self.conn.RemoteAddr())
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

func (self *Client)Send(data []byte){
	self.send(self.packets.AddNewDataPacket(data))
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

