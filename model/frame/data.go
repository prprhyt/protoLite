package frame

/*
proto-lite DATA Frame
In FrameData:
|data|

*/

type DATA struct {
	Data []byte
}

func NewDATAFromBinary(rawSrc []byte) *DATA {
	return &DATA{
		rawSrc,
	}
}

func (self *DATA) ToBytes()([]byte) {
	return self.Data
}
