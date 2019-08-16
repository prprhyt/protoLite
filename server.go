package main

import (
	"fmt"
	"github.com/protoLite/model"
	"github.com/protoLite/model/frame"
	"log"
	"net"
	"os"
	"time"
)
func main() {
	srcAddr := "192.168.22.1"
	dstAddr := "192.168.22.2"
	if(3==len(os.Args)){
		srcAddr = os.Args[1]
		dstAddr = os.Args[2]
	}
	client := NewSubClient(dstAddr+":0", srcAddr+":8889")
	server := NewServer(":8888", *client)
	for {
		server.recv()
	}
}

type SubClient struct {
	Conn net.UDPConn
}

func NewSubClient(srcAddressString string, dstAddressString string) *SubClient{
	localUdpAddr, err := net.ResolveUDPAddr("udp4", srcAddressString)
	if err != nil {
		log.Fatal(err)
	}

	remoteUdpAddr, err := net.ResolveUDPAddr("udp4", dstAddressString)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialUDP("udp4", localUdpAddr, remoteUdpAddr)

	return &SubClient{
		*conn,
	}
}

type Server struct {
	SenderCh chan model.Packet
	ReceiverCh chan model.Packet
	ackPacketCh chan frame.AckAddr
	conn net.PacketConn
	recvPackets model.Packets
	sendPackets model.Packets
	Client SubClient
	FileSubFrames []model.FileFrame
	FileCollector model.FileCollector
}

func NewServer(remoteAddrString string, client SubClient) *Server {
	conn, err := net.ListenPacket("udp", remoteAddrString)
	ackPacketCh := make(chan frame.AckAddr)
	ackPacketChDummy := make(chan frame.AckAddr)
	recvPackets := model.NewPackets(ackPacketCh)
	sendPackets := model.NewPackets(ackPacketChDummy)
	fileSubFrames := []model.FileFrame{}
	filecollector := model.NewFileCollector()
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
		client,
		fileSubFrames,
		*filecollector,
	}
	go server.sendAsync(server.SenderCh)
	go server.recvPacket(server.ReceiverCh)
	go server.reSendAckPacket(server.ackPacketCh)

	/*go func() {
		t := time.NewTicker(500 * time.Millisecond)
		for {
			select {
			case <-t.C:
				server.PrintlnReceivePackets()
			}
		}
		t.Stop()
	}()*/

	return server
}

func (self *Server)PrintlnReceivePackets(){
	/*recvPacketOffsets := []uint32{}
	for k,_ := range self.recvPackets.RecvData{
		recvPacketOffsets = append(recvPacketOffsets, k)
	}
	sort.Slice(recvPacketOffsets, func(i, j int) bool {
		return recvPacketOffsets[i] < recvPacketOffsets[j]
	})
	if len(recvPacketOffsets)==0{
		return
	}
	for _,e := range recvPacketOffsets{
		//fmt.Print(strconv.Itoa(self.recvPackets.RecvData[e])+" ")
		fmt.Print(strconv.Itoa(int(e))+" ")
	}
	fmt.Print("\n")*/
	if len(self.recvPackets.RecvData)==0{
		return
	}
	//fmt.Print("ReceivePackets: "+strconv.Itoa(len(self.recvPackets.RecvData)))
	//fmt.Print("\n")
}

func (self *Server)recv() {
	for{
		ret :=make([]byte, model.GetPacketByteLength())
		_, remoteAddress, _ :=self.conn.ReadFrom(ret)
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
			self.FileCollector.SetData(dataFrame.Data[0], i.Offset, dataFrame.Data[2:])
			if(model.FileFinFrameType.GetByte() == dataFrame.Data[1]){
				if(self.FileCollector.GetFinishFlag(dataFrame.Data[0])){
					if(self.FileCollector.IsFilePacketComplete(dataFrame.Data[0])){
						self.FileCollector.MakeFile(dataFrame.Data[0])
					}
				}
			}
			if(model.FileFinFrameType.GetByte() == dataFrame.Data[1]){
				//fmt.Print("receive fin")
				if(self.FileCollector.IsFilePacketComplete(dataFrame.Data[0])){
					self.FileCollector.MakeFile(dataFrame.Data[0])
				}
			}
			if(model.FileDataWithFinFrameType.GetByte() == dataFrame.Data[1]){
				//fmt.Print("receive fin")
				self.FileCollector.SetData(dataFrame.Data[0], i.Offset, dataFrame.Data[2:])
				if(self.FileCollector.IsFilePacketComplete(dataFrame.Data[0])){
					self.FileCollector.MakeFile(dataFrame.Data[0])
				}
			}


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
	var waitMs time.Duration = 1
	for {
		i := <- ch
		_, err := self.Client.Conn.Write(i.ToBytes())
		if err != nil {
			//panic(err)
			waitMs+=5
		}else {
			if(1<=waitMs-3){
				waitMs-=3
			}else{
				waitMs = 1
			}
		}
		if(0<waitMs){
			time.Sleep(waitMs*time.Millisecond)
			if err != nil {
				self.SenderCh <- i
			}
		}
	}
}