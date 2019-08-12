package main

import (
	"fmt"
	"github.com/proto-lite/model"
	"github.com/proto-lite/model/frame"
	"net"
)
func main() {
	server := NewServer(":8888")
	for {
		server.recv()
	}
}

type Server struct {
	SenderCh chan model.Packet
	ReceiverCh chan model.Packet
	ackPacketCh chan frame.AckAddr
	conn net.PacketConn
	recvPackets model.Packets
	sendPackets model.Packets
}

func NewServer(remoteAddrString string) *Server {
	conn, err := net.ListenPacket("udp", remoteAddrString)
	ackPacketCh := make(chan frame.AckAddr)
	ackPacketChDummy := make(chan frame.AckAddr)
	recvPackets := model.NewPackets(ackPacketCh)
	sendPackets := model.NewPackets(ackPacketChDummy)
	if err != nil {
		panic(err)
	}
	server :=  &Server{
		make(chan model.Packet),
		make(chan model.Packet),
		ackPacketCh,
		conn,
		*recvPackets,
		*sendPackets,
	}
	go server.sendAsync(server.SenderCh)
	go server.recvPacket(server.ReceiverCh)
	go server.reSendAckPacket(server.ackPacketCh)
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

func (self *Server)reSendAckPacket(ch <- chan frame.AckAddr) {
	for{
		i := <- ch
		self.send(self.sendPackets.AddNewAckPacket(i.SrcAddr, i.AckFrame.ToBytes()))
	}
}


func (self *Server)recvPacket(ch <- chan model.Packet) {
	for{
		i := <- ch
		self.recvPackets.AddPacketFromReceivePacket(i)
		if(model.DataFrameType.GetByte() == i.FrameType){
			dataFrame := frame.NewDATAFromReceiveBinary(i.FrameData)
			self.recvPackets.RecvData[i.Offset] = dataFrame.Data
			//fmt.Print(string(dataFrame.Data))

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