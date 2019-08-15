package model

import (
	"encoding/binary"
	"github.com/protoLite/model/frame"
	"net"
)

/*
For IPv4: 3009byte
|id(4bytes)|offset(4byte)|alias_len(1byte)|alias(alias_len*4byte)|frame_type(1byte)|frame_data(MAX 3000byte)|

srcAddr and dstAddr: get from UDP header
Byte order: little endian
*/

type Packet struct {
	Src net.Addr
	Dst net.Addr
	Id uint32
	Offset uint32
	AliasIDs []uint32
	FrameType byte
	FrameData []byte

	/*frame.DATA
	frame.Ack*/
}

func GetPacketByteLength()(int){
	return 60009
}

func NewPacketFromReceiveByte(rawSrc []byte, srcAddr net.Addr, dstAddr net.Addr) *Packet {
	aliasLen := binary.LittleEndian.Uint32([]byte{rawSrc[8],0x00,0x00,0x00})
	aliasIDs := []uint32{}
	var i uint32 = 0
	for ;i<aliasLen;i++{
		aliasIDs = append(aliasIDs, binary.LittleEndian.Uint32(rawSrc[9+4*i:9+4*(i+1)]))
	}
	return &Packet{
		srcAddr,
		dstAddr,
		binary.LittleEndian.Uint32(rawSrc[:4]),
		binary.LittleEndian.Uint32(rawSrc[4:8]),
		aliasIDs,
		rawSrc[9+4*aliasLen],
		rawSrc[9+4*aliasLen+1:],
	}
}

func NewDataPacketFromPayload(id uint32, offset uint32,rawSrc []byte, aliasIDs []uint32) *Packet {
	return &Packet{
		nil,
		nil,
		id,
		offset,
		aliasIDs,
		DataFrameType.GetByte(),
		frame.NewDATAFromBinary(uint32(len(rawSrc)), rawSrc).ToBytes(),
	}
}

func GetFrameTypeFromRawData(rawSrc []byte) byte{
	aliasLen := binary.LittleEndian.Uint32([]byte{rawSrc[8],0x00,0x00,0x00})
	return rawSrc[9+4*aliasLen]
}

func NewAckPacketFromPayload(srcAddr net.Addr, id uint32, offset uint32,rawSrc []byte, aliasIDs []uint32) *Packet {
	return &Packet{
		srcAddr,
		nil,
		id,
		offset,
		aliasIDs,
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
	ret = append(ret,byte(len(self.AliasIDs)))
	for _,e := range self.AliasIDs{
		tmp = []byte{0x00,0x00,0x00,0x00}
		binary.LittleEndian.PutUint32(tmp, e)
		ret = append(ret, tmp...)
	}
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