package model

import (
	"os"
	"strconv"
)

type FileCollector struct {
	StartEndOffset map[uint32][]uint32 // key: file_id, value: [start_offset, end_offset]
	Data map[uint32]map[uint32][]byte // key: offset, value: data
	FileList map[uint32]bool
	FinishFlagReceived map[uint32]bool
	FileLength map[uint32]uint32
}

func NewFileCollector() *FileCollector{
	filecollector := FileCollector{}
	filecollector.StartEndOffset = make(map[uint32][]uint32)
	filecollector.Data = make(map[uint32]map[uint32][]byte)
	filecollector.FileList = make(map[uint32]bool)
	filecollector.FinishFlagReceived = make(map[uint32]bool)
	filecollector.FileLength = make(map[uint32]uint32)
	return &filecollector
}

func(self *FileCollector)SetFinishFlag(id uint32){
	self.FinishFlagReceived[id] = true
}

func(self *FileCollector)GetFinishFlag(id uint32)bool{
	_, exist := self.FinishFlagReceived[id]
	if(!exist){
		return false
	}
	return self.FinishFlagReceived[id]
}

func(self *FileCollector)SetStartOffset(id uint32, offset uint32){
	_, exist := self.StartEndOffset[id]
	if(exist){
		if(self.StartEndOffset[id][0]>offset){
			self.StartEndOffset[id][0] = offset
		}
		return
	}
	self.StartEndOffset[id] = []uint32{offset,0}
}

func(self *FileCollector)SetEndOffset(id uint32, offset uint32){
	_, exist := self.StartEndOffset[id]
	if(!exist){
		return
	}
	self.StartEndOffset[id][1] = offset
}


func(self *FileCollector)SetFileLength(id uint32,length uint32){
	self.FileLength[id] = length
}

func(self *FileCollector)SetData(id uint32,offset uint32, data []byte){
	self.SetStartOffset(id, offset)
	_, exist := self.Data[id]
	if(!exist){
		self.Data[id] = make(map[uint32][]byte)
	}
	self.Data[id][offset] = data
}

func(self *FileCollector)IsFilePacketComplete(id uint32) bool{
	var i uint32 = 0
	var length uint32 = 0
	for ;i< uint32(len(self.Data[id]));i++{
		_, exist := self.Data[id]
		if(!exist){
			return false
		}
		length+=uint32(len(self.Data[i]))
	}
	return length*8>=self.FileLength[id] //true
}

func(self *FileCollector) MakeFile(id uint32){
	data := []byte{}
	for i := self.StartEndOffset[id][0];i<uint32(len(self.Data[id]))+self.StartEndOffset[id][0];i++{
		data = append(data, self.Data[id][i]...)
	}
	file, err := os.Create(`dst/`+strconv.FormatUint(uint64(id), 10)+`.bin`)
	if err != nil {
		// Openエラー処理
	}
	defer file.Close()
	file.Write(([]byte)(data))
	self.FileList[id] = true
	delete(self.Data,id)
	delete(self.StartEndOffset,id)
	delete(self.FinishFlagReceived,id)

}

