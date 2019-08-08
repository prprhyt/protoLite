package model

import(
	"encoding/binary"
	"net"
)

/*
For IPv4
|id(4bytes)|offset(4byte)|frame_type(1byte)|frame_data()|

srcAddr and dstAddr: get from UDP header
Byte order: little endian
*/

type Packet struct {
	Src string
	Dst string
	Id uint32
	Offset uint32
	FrameType byte
	FrameData []byte
}

func (self *Packet) parseFromBytes(rawSrc []byte, remoteAddr net.Addr)(Packet)  {
	self.Src = remoteAddr.String()
	self.Id = binary.LittleEndian.Uint32(rawSrc[:4])
	self.Offset = binary.LittleEndian.Uint32(rawSrc[4:8])
	self.FrameType = rawSrc[8]
	self.FrameData = rawSrc[9:]
}