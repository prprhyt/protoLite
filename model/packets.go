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
	sentButUnknownStatePacketID map[uint32]bool
	acceptPacketID map[uint32]bool
	SenderAckCh chan frame.AckAddr
	packetIDAlias map[uint32]uint32
	RecvData map[uint32][]byte
}

func NewPackets(SenderAckCh chan frame.AckAddr)(*Packets){
	packets := &Packets{}
	packets.latestId = 0
	packets.latestOffset = 0
	packets.lossPacketID = make(map[uint32]bool)
	packets.sentButUnknownStatePacketID = make(map[uint32]bool)
	packets.acceptPacketID = make(map[uint32]bool)
	packets.packetIDAlias = make(map[uint32]uint32)
	packets.RecvData = make( map[uint32][]byte)
	packets.SenderAckCh = SenderAckCh
	return packets
}

func(self *Packets) AddPacket(packet Packet){
	self.Packets = append(self.Packets, packet)
}

func(self *Packets) AddPacketFromReceivePacket(packet Packet)(Packet){
	self.AddPacket(packet)
	self.AddAcceptPacketIDs([]uint32{packet.Id})
	self.AddAcceptPacketIDs(packet.AliasIDs)
	if self.latestId+1 == packet.Id{
		self.latestId++
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
	packet := NewDataPacketFromPayload(self.latestId, self.latestOffset, rawSrc, []uint32{})
	self.AddPacket(*packet)
	self.latestId++
	self.latestOffset++
	return *packet
}

func(self *Packets) AddNewAckPacket(srcAddr net.Addr ,rawSrc []byte)(Packet){
	packet := NewAckPacketFromPayload(srcAddr, self.latestId, self.latestOffset, rawSrc, []uint32{})
	self.AddPacket(*packet)
	self.latestId++
	self.latestOffset++
	return *packet
}

func(self *Packets) AddResendPacket(packet Packet)(Packet){
	alias := []uint32{}
	self.packetIDAlias[self.latestId] = packet.Id
	packet.Id = self.latestId
	id := self.latestId
	for;;{
		_, exist := self.packetIDAlias[id]
		if(!exist){
			break
		}
		id := self.packetIDAlias[id]
		alias = append(alias, id)
	}
	packet.AliasIDs = alias
	self.AddPacket(packet)
	self.latestId++
	return packet
}

func(self *Packets) GetLatestPacket()(Packet){
	return self.Packets[len(self.Packets)-1]
}


func(self *Packets) AddSentButUnknownStatePacketIDs(ids []uint32){
	for _,i := range ids{
		self.sentButUnknownStatePacketID[i] = true
		_, exist := self.acceptPacketID[i]
		if(exist){
			delete(self.sentButUnknownStatePacketID,i)
		}
		_, exist = self.lossPacketID[i]
		if(exist){
			delete(self.sentButUnknownStatePacketID,i)
		}
	}
}

func(self *Packets) GetSentButUnknownStatePacketIDs()([]uint32){
	ks := []uint32{}
	for k, e := range self.sentButUnknownStatePacketID {
		if(e){
			ks = append(ks, k)
		}
	}
	return ks
}

func(self *Packets) AddAcceptPacketIDs(ids []uint32){
	for _,i := range ids{
		self.acceptPacketID[i] = true
		_, exist := self.lossPacketID[i]
		if(exist){
			delete(self.lossPacketID,i)
		}
		_, exist = self.sentButUnknownStatePacketID[i]
		if(exist){
			delete(self.sentButUnknownStatePacketID,i)
		}
	}
}

func(self *Packets) AddLossPacketIDs(ids []uint32){
	for _,i := range ids{
		if (self.acceptPacketID[i]){
			continue
		}
		self.lossPacketID[i] = true
		_, exist := self.sentButUnknownStatePacketID[i]
		if(exist){
			delete(self.sentButUnknownStatePacketID,i)
		}
	}
}

func(self *Packets) GetLossPacketIDs()([]uint32){
	ks := []uint32{}
	for k, e := range self.lossPacketID {
		if(e){
			ks = append(ks, k)
		}
	}
	return ks
}



