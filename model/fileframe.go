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
	i := 0
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
		data = append(data, []byte{id,0x00})
		data[i] = append(data[i], buf[:n]...)
		i++
	}
	data[len(data)-1][1] = 0x02 //FileDataWithFinFrameType
	return data
}

type FileSubFrameType int

const (
	FileDataFrameType FileSubFrameType = iota
	FileFinFrameType
	FileDataWithFinFrameType
)

func (e FileSubFrameType) GetByte() byte{
	switch e {
	case FileDataFrameType:
		return 0x00
	case FileFinFrameType:
		return 0x01
	case FileDataWithFinFrameType:
		return 0x02
	default:
		return 0xff
	}
}
