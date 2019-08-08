package frame

/*
proto-lite ACK Frame
In FrameData:
|LargestAcknowledged(4byte)|FirstACKRange(4byte)|Gap(4byte)|ACKRange(4byte)|

references:
- https://tools.ietf.org/html/draft-ietf-quic-recovery-22
- https://asnokaze.hatenablog.com/entry/2019/07/04/023545
*/

type Ack struct {
	LargestAcknowledged uint32
	//ACKRangeCount byte
	FirstACKRange uint32
	Gap uint32
	ACKRange uint32
}