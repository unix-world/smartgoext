
// AES256CBC/SHA1
// (c) 2023-2024 unix-world.org
// r.20241117.2358

package aes256cbc

import (
	"errors"

	"crypto/sha1"
	"crypto/cipher"
	"crypto/aes"

	smart "github.com/unix-world/smartgo"
	pbkdf2 "github.com/unix-world/smartgo/crypto/pbkdf2"

	cryptoutils "github.com/unix-world/smartgoext/crypto/utils"
)

const (
	VERSION string = "20241117.2358"
	// Algo Params
	aes256Iterations int = 2000
	aes256LengthKey  int = 32
	aes256LengthSalt int = 16
)


func GenerateRandomSalt() (string, error) {
	//--
	return smart.GenerateRandomString(aes256LengthSalt)
	//--
} //END FUNCTION


func GenerateRandomIv() ([]byte, error) {
	//--
	return smart.GenerateRandomBytes(aes.BlockSize)
	//--
} //END FUNCTION


func EncryptBase64AES256CBC(key string, text string, iv []byte, salt string) (string, error) { // b64EncData, errEnc
	//--
	defer smart.PanicHandler() // for aes ...
	//--
	if(smart.StrTrimWhitespaces(text) == "") {
		return "", nil
	} //end if
	//--
	if((iv == nil) || (len(iv) != aes.BlockSize)) {
		return "", smart.NewError("Invalid iV Length")
	} //end if
	//--
	salt = smart.StrTrimWhitespaces(salt)
	if(len(salt) != aes256LengthSalt) {
		return "", smart.NewError("Invalid Salt Length")
	} //end if
	//--
	dkey := pbkdf2.Key([]byte(key), []byte(salt), aes256Iterations, aes256LengthKey, sha1.New)
	block, err := aes.NewCipher(dkey)
	if(err != nil) {
		return "", err
	} //end if
	//--
	data, errPad := cryptoutils.Pkcs7Pad([]byte(text), aes.BlockSize)
	if(errPad != nil) {
		return "", errPad
	} //end if
	//--
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(data, data)
	//--
	return smart.Base64Encode(string(data)), nil
	//--
} //END FUNCTION


func DecryptBase64AES256CBC(key string, b64Data string, iv []byte, salt string) (string, error) { // decStr, decErr
	//--
	defer smart.PanicHandler() // base64 decode may panic
	//--
	if(smart.StrTrimWhitespaces(b64Data) == "") {
		return "", nil
	} //end if
	//--
	if((iv == nil) || (len(iv) != aes.BlockSize)) {
		return "", smart.NewError("Invalid iV Length")
	} //end if
	//--
	salt = smart.StrTrimWhitespaces(salt)
	if(len(salt) != aes256LengthSalt) {
		return "", smart.NewError("Invalid Salt Length")
	} //end if
	//--
	b64Data = smart.StrTrimWhitespaces(b64Data)
	if(b64Data == "") {
		return "", nil // empty b64Data after trim
	} //end if
	//--
	dataRaw := smart.Base64Decode(b64Data)
	b64Data = "" // free mem
	if(dataRaw == "") { // if the b64Data is not empty but decoding is empty ... this is an error !
		return "", errors.New("ERR: Base64 Data Decoding Failed")
	} //end if
	var data []byte = []byte(dataRaw)
	dataRaw = "" // free mem
	//--
	dkey := pbkdf2.Key([]byte(key), []byte(salt), aes256Iterations, aes256LengthKey, sha1.New)
	block, err := aes.NewCipher(dkey)
	if(err != nil) {
		return "", err
	} //end if
	//--
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(data, data)
	//--
	text, errUnpad := cryptoutils.Pkcs7Unpad(data, aes.BlockSize)
	//--
	return string(text), errUnpad // if errUnpad still have to return de decrypted data to try recovering (post decrypt) from extra padding if the case
	//--
} //END FUNCTION


// #END
