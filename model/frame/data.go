package frame

import "encoding/binary"

/*
protoLite DATA Frame
In FrameData:
|len|data|

*/

type DATA struct {
	len uint32
	Data []byte
}

func NewDATAFromReceiveBinary(rawSrc []byte) *DATA {
	length := binary.LittleEndian.Uint32(rawSrc[:4])
	data := rawSrc[4:4+length]
	return &DATA{
		length,
		data,
	}
}

func NewDATAFromBinary(len uint32, payload []byte) *DATA {
	return &DATA{
		len,
		payload,
	}
}

func (self *DATA) ToBytes()([]byte) {
	ret := []byte{0x00,0x00,0x00,0x00}
	binary.LittleEndian.PutUint32(ret, self.len)
	ret = append(ret, self.Data...)
	return ret
}
