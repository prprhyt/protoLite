package model

import (
	"fmt"
	"os"
)

/*
In DataFrame:
|id(1byte)|FileSubFrameType(1byte)|data|
*/


const MAX_FILE_SUB_FRAME_DATA_LENGTH = 1000

type FileFrame struct {
	FileName string
	Id byte
	Data map[uint32][]byte // key: offset, value: data
}

func GetDataArrayFileFromFilePath(filePath string, id byte)([][]byte){
	data := [][]byte{}
	f, err := os.Open(filePath)
	if err != nil{
		fmt.Println("error")
	}
	defer f.Close()
	buf := make([]byte, MAX_FILE_SUB_FRAME_DATA_LENGTH)
	for {
		// nはバイト数を示す
		n, err := f.Read(buf)
		// バイト数が0になることは、読み取り終了を示す
		if n == 0{
			break
		}
		if err != nil{
			break
		}
		data = append(data, buf[:n])
	}
	return data
}

type FileSubFrameType int

const (
	FileNameFrameType FileSubFrameType = iota
	FileDataFrameType
	FileFinFrameType
)

func (e FileSubFrameType) GetByte() byte{
	switch e {
	case FileNameFrameType:
		return 0x00
	case FileDataFrameType:
		return 0x01
	case FileFinFrameType:
		return 0x02
	default:
		return 0xff
	}
}
