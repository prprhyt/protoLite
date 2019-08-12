package main

import (
	"github.com/proto-lite/model"
	"github.com/proto-lite/model/frame"
	"log"
	"net"
	"strconv"
	"time"
)
func main() {
	server := NewSubServer(":8889")
	client := NewClient("192.168.22.1:0", "192.168.22.2:8888", *server)
	go func(){
		for{
			ret :=make([]byte, model.GetPacketByteLength())
			server.conn.ReadFrom(ret)
			client.RecvPacketWithoutChannel(ret)
			//client.ReceiverCh <- ret
		}
	}()
	defer client.Close()
	for i:=0; ;i++  {
		client.Send([]byte(strconv.Itoa(i)))
		time.Sleep(500 * time.Millisecond)
	}
}

func NewSubServer(remoteAddrString string) *SubServer{
	conn, err := net.ListenPacket("udp4", remoteAddrString)
	if err != nil {
		panic(err)
	}
	return &SubServer{
		conn,
	}
}

type SubServer struct {
	conn net.PacketConn
}

type Client struct {
	SenderCh chan model.Packet
	ReceiverCh chan []byte
	conn net.Conn
	recvPackets model.Packets
	sendPackets model.Packets
	server SubServer
}

func NewClient(srcAddressString string, dstAddressString string, server SubServer) *Client {

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
		server,
	}
	client.SenderCh = make(chan model.Packet)
	go client.sendAsync(client.SenderCh)
	client.ReceiverCh = make(chan []byte)
	go client.RecvPacket(client.ReceiverCh)
	//go client.recv()

	go func() {
		t := time.NewTicker(2000 * time.Millisecond)
		for {
			select {
			case <-t.C:
				//TODO: Resend timer
				//client.resendUnkownStatePackets()
			}
		}
		t.Stop()
	}()

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

func (self *Client)RecvPacketWithoutChannel(i[]byte) {
	packet := model.NewPacketFromReceiveByte(i, self.conn.LocalAddr(), self.conn.RemoteAddr())
	if(model.DataFrameType.GetByte() == model.GetFrameTypeFromRawData(i)){

	}else if(model.AckFrameType.GetByte() == model.GetFrameTypeFromRawData(i)){
		ackFrame := frame.NewAckFromBinary(packet.FrameData)
		lossPackets, acPackets := ackFrame.GetLossAndAcceptedPacketIDs()
		self.recvPackets.AddLossPacketIDs(lossPackets)
		self.recvPackets.AddAcceptPacketIDs(acPackets)
		self.resendLossPackets()
	}
}

func (self *Client)RecvPacket(ch <- chan []byte) {
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

func (self *Client)resendUnkownStatePackets(){
	packetID := []uint32{}
	for _,i := range self.sendPackets.GetSentButUnknownStatePacketIDs(){
		packet := self.sendPackets.AddResendPacket(self.sendPackets.Packets[i])
		self.SenderCh <- packet
		packetID = append(packetID, packet.Id)
	}
	self.sendPackets.AddSentButUnknownStatePacketIDs(packetID)
}


func (self *Client)resendLossPackets(){
	packetID := []uint32{}
	for _,i := range self.recvPackets.GetLossPacketIDs(){
		packet := self.sendPackets.AddResendPacket(self.sendPackets.Packets[i])
		self.SenderCh <- packet
		packetID = append(packetID, packet.Id)
	}
	self.sendPackets.AddSentButUnknownStatePacketIDs(packetID)
}

func (self *Client)Send(data []byte){
	packet := self.sendPackets.AddNewDataPacket(data)
	self.send(packet)
	self.sendPackets.AddSentButUnknownStatePacketIDs([]uint32{packet.Id})
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

