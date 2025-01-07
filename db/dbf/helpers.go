package dbf

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// appendSlice data[]byte to slice and returns the new byte slice.
// Change this function if you want to tweak memory allocation for the database.
func appendSlice(slice, data []byte) []byte {
	slice = append(slice, data...)
	return slice
}

func readFile(filename string) ([]byte, []byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}

	// Look for associated dbt file
	var (
		memo []byte

		base = strings.TrimSuffix(filename, filepath.Ext(filename))
	)
	for _, dbtPath := range []string{base + ".dbt", base + ".DBT"} {
		dbtFile, err := os.Open(dbtPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, nil, err
		}

		memo, err = ioutil.ReadAll(dbtFile)
		if err != nil {
			return nil, nil, err
		}
		break
	}

	return data, memo, err
}

func uint32ToBytes(x uint32) []byte {
	var buf [4]byte
	buf[0] = byte(x >> 0)
	buf[1] = byte(x >> 8)
	buf[2] = byte(x >> 16)
	buf[3] = byte(x >> 24)
	return buf[:]
}

func int32ToBytes(x int32) []byte {
	var buf [4]byte
	buf[0] = byte(x >> 0)
	buf[1] = byte(x >> 8)
	buf[2] = byte(x >> 16)
	buf[3] = byte(x >> 24)
	return buf[:]
}

func uint64ToBytes(x uint64) []byte {
	var buf [8]byte
	buf[0] = byte(x >> 0)
	buf[1] = byte(x >> 8)
	buf[2] = byte(x >> 16)
	buf[3] = byte(x >> 24)
	buf[4] = byte(x >> 32)
	buf[5] = byte(x >> 40)
	buf[6] = byte(x >> 48)
	buf[7] = byte(x >> 56)

	return buf[:]
}

func int64ToBytes(x uint64) []byte {
	var buf [8]byte
	buf[0] = byte(x >> 0)
	buf[1] = byte(x >> 8)
	buf[2] = byte(x >> 16)
	buf[3] = byte(x >> 24)
	buf[4] = byte(x >> 32)
	buf[5] = byte(x >> 40)
	buf[6] = byte(x >> 48)
	buf[7] = byte(x >> 56)

	return buf[:]
}

func bytesToInt32le(b []byte) int32 {
	return int32(uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24)

}

func bytesToInt32be(b []byte) int32 {
	return int32(uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24)
}

func bytesToUint64le(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 | uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}
