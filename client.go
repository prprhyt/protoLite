package main

import (
	"github.com/proto-lite/model"
	"github.com/proto-lite/model/frame"
	"log"
	"net"
	"time"
)
func main() {
	client := NewClient("192.168.22.1:0", "192.168.22.2:8888")
	defer client.Close()
	client.Send([]byte("Hello World from Client"))
	for ; ;  {
		time.Sleep(1)
	}
}

type Client struct {
	SenderCh chan model.Packet
	ReceiverCh chan []byte
	conn net.Conn
	recvPackets model.Packets
	sendPackets model.Packets
}

func NewClient(srcAddressString string, dstAddressString string) *Client {

	localUdpAddr, err := net.ResolveUDPAddr("udp4", srcAddressString)
	if err != nil {
		log.Fatal(err)
	}

	remoteUdpAddr, err := net.ResolveUDPAddr("udp4", dstAddressString)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialUDP("udp4", localUdpAddr, remoteUdpAddr)
	ackPacketChDummy := make(chan frame.AckAddr)
	recvPackets := model.NewPackets(ackPacketChDummy)
	sendPackets := model.NewPackets(ackPacketChDummy)
	if err != nil {
		panic(err)
	}
	client :=  &Client{
		nil,
		nil,
		conn,
		*recvPackets,
		*sendPackets,
	}
	client.SenderCh = make(chan model.Packet)
	go client.sendAsync(client.SenderCh)
	client.ReceiverCh = make(chan []byte)
	go client.recvPacket(client.ReceiverCh)
	go client.recv()
	return client
}

func(self *Client)CreateChannel(){
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
		packet := self.recvPackets.AddPacketFromReceiveByte(i, self.conn.LocalAddr(), self.conn.RemoteAddr())
		if(model.DataFrameType.GetByte() == model.GetFrameTypeFromRawData(i)){

		}else if(model.AckFrameType.GetByte() == model.GetFrameTypeFromRawData(i)){
			ackFrame := frame.NewAckFromBinary(packet.FrameData)
			lossPackets, acPackets := ackFrame.GetLossAndAcceptedPacketIDs()
			self.recvPackets.AddLossPacketIDs(lossPackets)
			self.recvPackets.AddAcceptPacketIDs(acPackets)
			self.resendLossPackets()
		}
	}
}

func (self *Client)resendLossPackets(){
	for _,i := range self.recvPackets.GetLossPacketIDs(){
		self.SenderCh <- self.sendPackets.AddResendPacket(self.sendPackets.Packets[i])
	}
}

func (self *Client)Send(data []byte){
	self.send(self.sendPackets.AddNewDataPacket(data))
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

