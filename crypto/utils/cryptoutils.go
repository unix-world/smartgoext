
// Crypto Utils
// (c) 2023-2024 unix-world.org
// r.20241117.2358

package cryptoutils

import (
	"fmt"
	"errors"
	"bytes"
)


func Pkcs7Pad(data []byte, blockSize int) ([]byte, error) {
	//--
	if(blockSize <= 0) {
		return nil, fmt.Errorf("invalid blockSize: %d", blockSize)
	} //end if
	if(len(data) == 0) {
		return nil, errors.New("data is empty")
	} //end if
	padlen := 1
	for ((len(data) + padlen) % blockSize) != 0 {
		padlen = padlen + 1
	} //end for
	pad := bytes.Repeat([]byte{byte(padlen)}, padlen)
	//--
	return append(data, pad...), nil
	//--
} //END FUNCTION


func Pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	//--
	if(blockSize <= 0) {
		return nil, fmt.Errorf("invalid blockSize: %d", blockSize)
	} //end if
	if((len(data) % blockSize != 0) || (len(data) == 0)) {
		return nil, fmt.Errorf("invalid data length: %d", len(data))
	} //end if
	padlen := int(data[len(data)-1])
	if((padlen > blockSize) || (padlen == 0)) {
		return nil, errors.New("invalid padding")
	} //end if
	pad := data[len(data)-padlen:]
	for i := 0; i < padlen; i++ {
		if(pad[i] != byte(padlen)) {
			return nil, errors.New("invalid padding byte")
		} //end if
	} //end for
	//--
	return data[:len(data)-padlen], nil
	//--
} //END FUNCTION


// #END
