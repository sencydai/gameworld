package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
)

const (
	DEFAULT_TAG     int   = 0xccee
	DEFAULT_CRC_KEY int16 = 0x765d

	HEAD_SIZE = 12
)

var (
	ReadEOF = errors.New("pack.Read EOF")

	header = NewWriter(DEFAULT_TAG, int32(0), int16(0), DEFAULT_CRC_KEY).Bytes()
)

type LString string

func NewWriter(datas ...interface{}) *bytes.Buffer {
	writer := bytes.NewBuffer([]byte{})
	Write(writer, datas...)
	return writer
}

func GetBytes(datas ...interface{}) []byte {
	writer := NewWriter(datas...)
	return writer.Bytes()
}

func Read(reader *bytes.Reader, datas ...interface{}) {
	for _, data := range datas {
		switch v := data.(type) {
		case *bool, *int8, *uint8, *int16, *uint16, *int32, *uint32, *int64, *uint64, *float32, *float64:
			err := binary.Read(reader, binary.LittleEndian, v)
			if err != nil {
				panic(ReadEOF)
			}
		case *int:
			var vv int32
			err := binary.Read(reader, binary.LittleEndian, &vv)
			if err != nil {
				panic(ReadEOF)
			}
			*v = int(vv)
		case *string:
			var l uint16
			err := binary.Read(reader, binary.LittleEndian, &l)
			if err != nil {
				panic(ReadEOF)
			}
			s := make([]byte, l)
			n, _ := reader.Read(s)
			if uint16(n) < l {
				panic(ReadEOF)
			}
			*v = string(s)
			_, err = reader.ReadByte()
			if err != nil {
				panic(ReadEOF)
			}
		case *LString:
			var l uint64
			err := binary.Read(reader, binary.LittleEndian, &l)
			if err != nil {
				panic(ReadEOF)
			}
			s := make([]byte, l)
			n, _ := reader.Read(s)
			if uint64(n) < l {
				panic(ReadEOF)
			}
			*v = LString(s)
			_, err = reader.ReadByte()
			if err != nil {
				panic(ReadEOF)
			}
		default:
			panic("pack.Read invalid type " + reflect.TypeOf(data).String())
		}
	}
}

func Write(writer *bytes.Buffer, datas ...interface{}) {
	for _, data := range datas {
		switch v := data.(type) {
		case bool, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
			binary.Write(writer, binary.LittleEndian, v)
		case int:
			binary.Write(writer, binary.LittleEndian, int32(v))
		case []byte:
			writer.Write(v)
		case string:
			binary.Write(writer, binary.LittleEndian, uint16(len(v)))
			writer.Write([]byte(v))
			binary.Write(writer, binary.LittleEndian, byte(0))
		case LString:
			binary.Write(writer, binary.LittleEndian, uint64(len(v)))
			writer.Write([]byte(v))
			binary.Write(writer, binary.LittleEndian, byte(0))
		default:
			panic("pack.Write invalid type " + reflect.TypeOf(data).String())
		}
	}
}

func AllocPack(sysId, cmdId byte, data ...interface{}) *bytes.Buffer {
	writer := NewWriter(header, sysId, cmdId)
	Write(writer, data...)
	return writer
}

func EncodeWriter(writer *bytes.Buffer) []byte {
	data := writer.Bytes()
	encode(data)
	return data
}

func EncodeData(sysId, cmdId byte, data ...interface{}) []byte {
	writer := AllocPack(sysId, cmdId, data...)
	return EncodeWriter(writer)
}

func encode(data []byte) {
	copy(data[4:], GetBytes(len(data)-HEAD_SIZE))
}
