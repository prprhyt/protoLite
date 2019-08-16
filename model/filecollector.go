package model

import (
	"encoding/binary"
	"os"
	"strconv"
)

type FileCollector struct {
	StartEndOffset map[byte][]uint32 // key: file_id, value: [start_offset, end_offset]
	Data map[byte]map[uint32][]byte // key: offset, value: data
	FileList map[uint32]bool
	FinishFlagReceived map[byte]bool
}

func NewFileCollector() *FileCollector{
	filecollector := FileCollector{}
	filecollector.StartEndOffset = make(map[byte][]uint32)
	filecollector.Data = make(map[byte]map[uint32][]byte)
	filecollector.FileList = make(map[uint32]bool)
	filecollector.FinishFlagReceived = make(map[byte]bool)
	return &filecollector
}

func(self *FileCollector)SetFinishFlag(id byte){
	self.FinishFlagReceived[id] = true
}

func(self *FileCollector)GetFinishFlag(id byte)bool{
	_, exist := self.FinishFlagReceived[id]
	if(!exist){
		return false
	}
	return self.FinishFlagReceived[id]
}

func(self *FileCollector)SetStartOffset(id byte, offset uint32){
	_, exist := self.StartEndOffset[id]
	if(exist){
		return
	}
	self.StartEndOffset[id] = []uint32{offset,0}
}

func(self *FileCollector)SetEndOffset(id byte, offset uint32){
	_, exist := self.StartEndOffset[id]
	if(!exist){
		return
	}
	self.StartEndOffset[id][1] = offset
}

func(self *FileCollector)SetData(id byte,offset uint32, data []byte){
	self.SetStartOffset(id, offset)
	_, exist := self.Data[id]
	if(!exist){
		self.Data[id] = make(map[uint32][]byte)
	}
	self.Data[id][offset] = data
}

func(self *FileCollector)IsFilePacketComplete(id byte) bool{
	var i uint32 = 0
	for ;i< uint32(len(self.Data[id]));i++{
		_, exist := self.Data[id]
		if(!exist){
			return false
		}
	}
	return true
}

func(self *FileCollector) MakeFile(id byte){
	data := []byte{}
	for i := self.StartEndOffset[id][0];i<uint32(len(self.Data[id]))+self.StartEndOffset[id][0];i++{
		data = append(data, self.Data[id][i]...)
	}
	var fid uint32 = 0
	binary.LittleEndian.PutUint32([]byte{0,0,0,id}, fid)
	for ;; {
		_, exist := self.FileList[fid]
		if(!exist){
			break
		}
		fid = fid+256
	}
	file, err := os.Create(`dst/`+strconv.FormatUint(uint64(fid), 10)+`.bin`)
	if err != nil {
		// Openエラー処理
	}
	defer file.Close()
	file.Write(([]byte)(data))
	self.FileList[fid] = true
	delete(self.Data,id)
	delete(self.StartEndOffset,id)
	delete(self.FinishFlagReceived,id)

}

