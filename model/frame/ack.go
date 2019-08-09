package frame

import (
	"encoding/binary"
)

/*
proto-lite ACK Frame
In FrameData:
|LargestAcknowledged(4byte)|ACKRangeCount(4byte)|ACKRanges|

In ACKRanges
|ACKRange(4byte)|Gap(4byte)| * ACKRangeCount

references:
- https://tools.ietf.org/html/draft-ietf-quic-recovery-22
- https://asnokaze.hatenablog.com/entry/2019/07/04/023545
*/

type ACKRangeUnit struct {
	ACKRange uint32 /*local maximum packetID*/
	Gap uint32
}

func NewAckRanges(ACKRangeCount uint32,rawSrc []byte) []ACKRangeUnit{
	ret := []ACKRangeUnit{}
	var i uint32 = 0
	var l = uint32(len(rawSrc))
	for ; i < ACKRangeCount; i++ {
		if(l<8*i+8){
			break
		}
		ret = append(ret, ACKRangeUnit{binary.LittleEndian.Uint32(rawSrc[8*i:8*i+4]),binary.LittleEndian.Uint32(rawSrc[8*i+4:8*i+8])})
	}
	return ret
}

type Ack struct {
	LargestAcknowledged uint32
	ACKRangeCount uint32
	ACKRanges []ACKRangeUnit
}

func NewAckFromBinary(rawSrc []byte) *Ack {
	return &Ack{
		binary.LittleEndian.Uint32(rawSrc[:4]),
		binary.LittleEndian.Uint32(rawSrc[4:8]),
		NewAckRanges(binary.LittleEndian.Uint32(rawSrc[4:8]),rawSrc[8:]),
	}
}

func NewAck(acceptPacketIDs []uint32) *Ack {
	ACKRanges := []ACKRangeUnit{}
	LargestAcknowledged := acceptPacketIDs[len(acceptPacketIDs)-1]
	var prev uint32 = acceptPacketIDs[0]
	for _,i := range acceptPacketIDs[0:]{
		if(prev+1==i){
			prev+=1
		}else{
			ACKRanges = append(ACKRanges, ACKRangeUnit{prev, i-prev-1})
			prev=i
		}
	}
	return &Ack{
		LargestAcknowledged,
		uint32(len(ACKRanges)),
		ACKRanges,
	}
}

func (self *Ack) GetLossAndAcceptedPacketIDs()([]uint32, []uint32) {
	loss := []uint32{}
	accepted := []uint32{}
	for i,e := range self.ACKRanges{
		var j uint32 = e.ACKRange+1
		var k uint32 = 0
		for ;j<(j+e.Gap);j++ {
			loss = append(loss, j)
		}
		if 0<i {
			k = self.ACKRanges[i-1].ACKRange + self.ACKRanges[i-1].Gap + 1
		}
		for j=k;j<(e.ACKRange+1); j++ {
			accepted = append(accepted, j)
		}
	}
	k := self.ACKRanges[len(self.ACKRanges)-1].ACKRange+self.ACKRanges[len(self.ACKRanges)-1].Gap+1

	for i:=k;i<(self.LargestAcknowledged+1); i++ {
		accepted = append(accepted, i)
	}

	return loss,accepted
}

func (self *Ack) ToBytes()([]byte) {
	ret := []byte{}
	tmp := []byte{0x00, 0x00, 0x00, 0x00}
	binary.LittleEndian.PutUint32(tmp, self.LargestAcknowledged)
	ret = append(ret,tmp...)
	tmp = []byte{0x00, 0x00, 0x00, 0x00}
	binary.LittleEndian.PutUint32(tmp, self.ACKRangeCount)
	ret = append(ret,tmp...)
	for _,i := range self.ACKRanges{
		tmp = []byte{0x00, 0x00, 0x00, 0x00}
		binary.LittleEndian.PutUint32(tmp, i.ACKRange)
		ret = append(ret, tmp...)
		tmp = []byte{0x00, 0x00, 0x00, 0x00}
		binary.LittleEndian.PutUint32(tmp, i.Gap)
		ret = append(ret,tmp...)
	}
	return ret
}
