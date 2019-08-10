package model

import "net"

type Packets struct {
	Packets  []Packet
	latestId uint32
	latestOffset uint32
	lossPacketID map[uint32]bool
	acceptPacketID map[uint32]bool
}

func NewPackets()(*Packets){
	packets := &Packets{}
	packets.latestId = 0
	packets.latestOffset = 0
	packets.lossPacketID = make(map[uint32]bool)
	packets.acceptPacketID = make(map[uint32]bool)
	return packets
}

func(self *Packets) AddPacket(packet Packet){
	self.Packets = append(self.Packets, packet)
}

func(self *Packets) AddPacketFromReceiveByte(rawSrc []byte, srcAddr net.Addr, dstAddr net.Addr)(Packet){
	packet := NewPacketFromReceiveByte(rawSrc, srcAddr, dstAddr)
	self.AddPacket(*packet)
	return *packet
}

func(self *Packets) AddNewDataPacket(rawSrc []byte)(Packet){
	packet := NewDataPacketFromPayload(self.latestId, self.latestOffset, rawSrc)
	self.AddPacket(*packet)
	self.latestId++
	self.latestOffset++
	return *packet
}

func(self *Packets) AddNewAckPacket(rawSrc []byte)(Packet){
	packet := NewAckPacketFromPayload(self.latestId, self.latestOffset, rawSrc)
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



