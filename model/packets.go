package model

import (
	"github.com/proto-lite/model/frame"
	"net"
	"sort"
)

type Packets struct {
	Packets  []Packet
	latestId uint32
	latestOffset uint32
	lossPacketID map[uint32]bool
	acceptPacketID map[uint32]bool
	SenderAckCh chan frame.AckAddr
}

func NewPackets(SenderAckCh chan frame.AckAddr)(*Packets){
	packets := &Packets{}
	packets.latestId = 0
	packets.latestOffset = 0
	packets.lossPacketID = make(map[uint32]bool)
	packets.acceptPacketID = make(map[uint32]bool)
	packets.SenderAckCh = SenderAckCh
	return packets
}

func(self *Packets) AddPacket(packet Packet){
	self.Packets = append(self.Packets, packet)
}

func(self *Packets) AddPacketFromReceivePacket(packet Packet)(Packet){
	self.AddPacket(packet)
	self.AddAcceptPacketIDs([]uint32{packet.Id})
	if self.latestId+1 == packet.Id{
		return packet
	}

	acPackets := []uint32{}
	for i, j := range self.acceptPacketID {
		if(j){
			acPackets = append(acPackets, i)
		}
	}
	sort.Slice(acPackets, func(i, j int) bool {
		return acPackets[i] < acPackets[j]
	})
	self.SenderAckCh <- frame.AckAddr{
		*frame.NewAck(acPackets),
		packet.Src,
	}
	return packet
}

func(self *Packets) AddPacketFromReceiveByte(rawSrc []byte, srcAddr net.Addr, dstAddr net.Addr)(Packet){
	packet := NewPacketFromReceiveByte(rawSrc, srcAddr, dstAddr)
	return self.AddPacketFromReceivePacket(*packet)
}

func(self *Packets) AddNewDataPacket(rawSrc []byte)(Packet){
	packet := NewDataPacketFromPayload(self.latestId, self.latestOffset, rawSrc)
	self.AddPacket(*packet)
	self.latestId++
	self.latestOffset++
	return *packet
}

func(self *Packets) AddNewAckPacket(srcAddr net.Addr ,rawSrc []byte)(Packet){
	packet := NewAckPacketFromPayload(srcAddr, self.latestId, self.latestOffset, rawSrc)
	self.AddPacket(*packet)
	self.latestId++
	self.latestOffset++
	return *packet
}

func(self *Packets) AddResendPacket(packet Packet)(Packet){
	packet.Id = self.latestId
	self.AddPacket(packet)
	self.latestId++
	return packet
}

func(self *Packets) GetLatestPacket()(Packet){
	return self.Packets[len(self.Packets)-1]
}

func(self *Packets) AddAcceptPacketIDs(ids []uint32){
	for _,i := range ids{
		self.acceptPacketID[i] = true
		_, exist := self.lossPacketID[i]
		if(exist){
			delete(self.lossPacketID,i)
		}
	}
}

func(self *Packets) AddLossPacketIDs(ids []uint32){
	for _,i := range ids{
		if (self.acceptPacketID[i]){
			continue
		}
		self.lossPacketID[i] = true
	}
}

func(self *Packets) GetLossPacketIDs()([]uint32){
	ks := []uint32{}
	for k, _ := range self.lossPacketID {
		if(self.lossPacketID[k]){
			ks = append(ks, k)
		}
	}
	return ks
}



