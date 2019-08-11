package main

import (
	"fmt"
	"github.com/proto-lite/model"
	"github.com/proto-lite/model/frame"
	"net"
)
func main() {
	server := NewServer("localhost:8888")
	for {
		server.recv()
	}
}

type Server struct {
	SenderCh chan model.Packet
	ReceiverCh chan model.Packet
	conn net.PacketConn
	recvPackets model.Packets
	sendPackets model.Packets
}

func NewServer(remoteAddrString string) *Server {
	conn, err := net.ListenPacket("udp", remoteAddrString)
	recvPackets := model.Packets{}
	sendPackets := model.Packets{}
	if err != nil {
		panic(err)
	}
	server :=  &Server{
		make(chan model.Packet),
		make(chan model.Packet),
		conn,
		recvPackets,
		sendPackets,
	}
	go server.sendAsync(server.SenderCh)
	go server.recvPacket(server.ReceiverCh)
	//go server.recv()
	return server
}

func (self *Server)recv() {
	for{
		ret :=make([]byte, model.GetPacketByteLength())
		_, remoteAddress, _ :=self.conn.ReadFrom(ret)
		fmt.Print("recv")
		packet := model.NewPacketFromReceiveByte(ret, remoteAddress, self.conn.LocalAddr())
		self.ReceiverCh <- *packet
	}
}

func (self *Server)recvPacket(ch <- chan model.Packet) {
	for{
		i := <- ch
		self.recvPackets.AddPacket(i)
		if(model.DataFrameType.GetByte() == i.FrameType){


		}else if(model.AckFrameType.GetByte() == i.FrameType){
			ackFrame := frame.NewAckFromBinary(i.FrameData)
			lossPackets, acPackets := ackFrame.GetLossAndAcceptedPacketIDs()
			self.recvPackets.AddLossPacketIDs(lossPackets)
			self.recvPackets.AddAcceptPacketIDs(acPackets)
		}
	}
}

func (self *Server)resendLossPackets(){
	for _,i := range self.recvPackets.GetLossPacketIDs(){
		self.SenderCh <- self.sendPackets.AddResendPacket(self.recvPackets.Packets[i])
	}
}

func (self *Server)send(packet model.Packet){
	self.SenderCh <- packet
}

func (self *Server)sendAsync(ch <-chan model.Packet)  {
	for {
		i := <- ch
		_, err := self.conn.WriteTo(i.ToBytes(), i.Src)
		if err != nil {
			panic(err)
		}
	}
}