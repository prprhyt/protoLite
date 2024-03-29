package model

import (
	"encoding/binary"
	"fmt"
	"os"
)

/*
In DataFrame:
|id(4byte)|FileSubFrameType(1byte)|data|
*/


const MAX_FILE_SUB_FRAME_DATA_LENGTH = 2500

type FileFrame struct {
	Id uint32
	Data map[uint32][]byte // key: offset, value: data
}

func GetDataArrayFileFromFilePath(filePath string, id uint32)([][]byte){
	data := [][]byte{}
	f, err := os.Open(filePath)
	if err != nil{
		fmt.Println("error")
	}
	defer f.Close()
	buf := make([]byte, MAX_FILE_SUB_FRAME_DATA_LENGTH)
	i := 0
	tmp := []byte{0x00,0x00,0x00,0x00}
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
		data_ := []byte{}
		binary.LittleEndian.PutUint32(tmp, id)
		data_ = append(data_, tmp...)
		data_ = append(data_, []byte{0x00}...)
		data_ = append(data_,  buf[:n]...)
		data = append(data, data_)
		i++
	}
	data = data[:i]
	data[len(data)-1][4] = 0x02 //FileDataWithFinFrameType
	fileinfo, _ := f.Stat()
	b:=fileinfo.Size()
	binary.LittleEndian.PutUint32(tmp, uint32(b))
	data[0] = append(data[0],tmp...)
	data[0][4] = 0x03
	return data
}

type FileSubFrameType int

const (
	FileDataFrameType FileSubFrameType = iota
	FileFinFrameType
	FileDataWithFinFrameType
	FileStartWithData
)

func (e FileSubFrameType) GetByte() byte{
	switch e {
	case FileDataFrameType:
		return 0x00
	case FileFinFrameType:
		return 0x01
	case FileDataWithFinFrameType:
		return 0x02
	case FileStartWithData:
		return 0x03
	default:
		return 0xff
	}
}
