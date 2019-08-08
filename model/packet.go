package model

import (
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

func NewPacket(rawSrc []byte, remoteAddr net.Addr) *Packet {
	return &Packet{
		remoteAddr.String(),
		"",
		binary.LittleEndian.Uint32(rawSrc[:4]),
		binary.LittleEndian.Uint32(rawSrc[4:8]),
		rawSrc[8],
		rawSrc[9:],
	}
}

func (self *Packet) ToBytes()([]byte)  {
	ret := []byte{}
	tmp := []byte{}
	binary.LittleEndian.PutUint32(tmp, self.Id)
	ret = append(ret,tmp...)
	tmp = []byte{}
	binary.LittleEndian.PutUint32(tmp, self.Offset)
	ret = append(ret, self.FrameType)
	ret = append(ret, self.FrameData...)
	return ret
}