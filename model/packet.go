package model

import (
	"encoding/binary"
	"net"
)

/*
For IPv4: 3009byte
|id(4bytes)|offset(4byte)|frame_type(1byte)|frame_data(MAX 3000byte)|

srcAddr and dstAddr: get from UDP header
Byte order: little endian
*/

type Packet struct {
	Src net.Addr
	Dst net.Addr
	Id uint32
	Offset uint32
	FrameType byte
	FrameData []byte
}

func GetPacketByteLength()(int){
	return 30009
}

func NewPacketFromReceiveByte(rawSrc []byte, srcAddr net.Addr, dstAddr net.Addr) *Packet {
	return &Packet{
		srcAddr,
		dstAddr,
		binary.LittleEndian.Uint32(rawSrc[:4]),
		binary.LittleEndian.Uint32(rawSrc[4:8]),
		rawSrc[8],
		rawSrc[9:],
	}
}

func NewDataPacketFromPayload(id uint32, offset uint32,rawSrc []byte) *Packet {
	return &Packet{
		"",
		"",
		id,
		offset,
		DataFrameType.GetByte(),
		rawSrc,
	}
}

func GetFrameTypeFromRawData(rawSrc []byte) byte{
	return rawSrc[8]
}

func NewAckPacketFromPayload(id uint32, offset uint32,rawSrc []byte) *Packet {
	return &Packet{
		"",
		"",
		id,
		offset,
		AckFrameType.GetByte(),
		rawSrc,
	}
}

func (self *Packet) ToBytes()([]byte)  {
	ret := []byte{}
	tmp := []byte{0x00,0x00,0x00,0x00}
	binary.LittleEndian.PutUint32(tmp, self.Id)
	ret = append(ret,tmp...)
	tmp = []byte{0x00,0x00,0x00,0x00}
	binary.LittleEndian.PutUint32(tmp, self.Offset)
	ret = append(ret,tmp...)
	ret = append(ret, self.FrameType)
	ret = append(ret, self.FrameData...)
	return ret
}

type FrameType int

const (
	DataFrameType FrameType = iota
	AckFrameType
)

func (e FrameType) GetByte() byte{
	switch e {
	case DataFrameType:
		return 0x00
	case AckFrameType:
		return 0x01
	default:
		return 0xff
	}
}