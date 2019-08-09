package model

import "net"

type Packets struct {
	packets  []Packet
	latestId uint32
	latestOffset uint32
}

func NewPackets()(*Packets){
	packets := &Packets{}
	packets.latestId = 0
	packets.latestOffset = 0
	return packets
}

func(self *Packets) addPacket(packet Packet){
	self.packets = append(self.packets, packet)
}

func(self *Packets) AddPacketFromReceiveByte(rawSrc []byte, srcAddr net.Addr){
	packet := NewPacketFromReceiveByte(rawSrc, srcAddr)
	self.addPacket(*packet)
}

func(self *Packets) AddNewDataPacket(rawSrc []byte){
	packet := NewDataPacketFromPayload(self.latestId, self.latestOffset, rawSrc)
	self.addPacket(*packet)
	self.latestId++
	self.latestOffset++
}

func(self *Packets) AddResendDataPacket(packet Packet){
	packet.Id = self.latestId
	self.addPacket(packet)
	self.latestId++
}

func(self *Packets) GetLatestPacket()(Packet){
	return self.packets[len(self.packets)-1]
}