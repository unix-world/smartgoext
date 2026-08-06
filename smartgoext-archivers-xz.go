
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2021-present, unix-world.org
// r.20260805.2358 :: STABLE
// [ ARCHIVERS / XZ ]

// REQUIRE: go 1.22 or later
package smartgoext

import (
	"io"
	"bytes"

	smart "github.com/unix-world/smartgo"
	"github.com/unix-world/smartgo/utils/iox"

	"github.com/unix-world/smartgoext/compress/xz"
)


//-----


func xzDictCapExps(level uint) uint {
	//--
	// private: handle the Xz Compression levels
	//--
	if((level < 0) || (level > 9)) { // {{{SYNC-XZ-LEVEL-CHECK}}}
		level = 6 // xz default compression level
	} //end if
	//--
	var lzmaDictCapExps []uint = []uint{18, 20, 21, 22, 22, 23, 23, 24, 25, 26} // this was get from gxz go utility, xz/cmd/gxz/file.go
	//--
	return lzmaDictCapExps[level]
	//--
} //END FUNCTION


//-----


func XzCompress(data []byte, level int, checksumMode string, verifyCompressed bool) ([]byte, error) { // compress xz ; file extension: .xz
	//--
	defer smart.PanicHandler()
	//--
	if(data == nil) {
		return nil, smart.NewError("Input Data is Empty")
	} //end if
	//--
	if((level < 0) || (level > 9)) { // {{{SYNC-XZ-LEVEL-CHECK}}}
		level = 6 // xz default compression
	} //end if
	//--
	var xzChecksumMode byte = xz.CRC64 // default checksum mode in xz
	switch(smart.StrToUpper(checksumMode)) {
		case "SHA256":
			xzChecksumMode = xz.SHA256
			break
		case "CRC32":
			xzChecksumMode = xz.CRC32
			break
		case "CRC64":
		//	xzChecksumMode = xz.CRC64
			break
		default:
			return nil, smart.NewError("Unsupported Checksum Mode: `" + checksumMode + "`")
	} //end switch
	//--
	var buf bytes.Buffer
	//--
	xzCfg := xz.WriterConfig{ // this was get from gxz go utility, xz/cmd/gxz/file.go
		DictCap: 1 << xzDictCapExps(uint(level)),
		CheckSum: xzChecksumMode,
	}
	w, errInit := xzCfg.NewWriter(&buf)
	if(errInit != nil) {
		return nil, smart.NewError("Compress Init Failed: " + errInit.Error())
	} //end if
	//--
	_, errWr := w.Write(data)
	if(errWr != nil) {
		return nil, smart.NewError("Compress Write Failed: " + errWr.Error())
	} //end if
	//--
	errClose := w.Close()
	if(errClose != nil) {
		return nil, smart.NewError("Compress Close Failed: " + errClose.Error())
	} //end if
	//--
	byts := buf.Bytes()
	if(byts == nil) {
		return nil, smart.NewError("Compressed Data is Empty")
	} //end if
	//--
	if(verifyCompressed == true) {
		unarchData, unarchErr := XzUncompress(byts)
		if(unarchErr != nil) {
			return nil, smart.NewError("Compressed Data Verification Failed: " + unarchErr.Error())
		} //end if
		if(unarchData == nil) {
			return nil, smart.NewError("Compressed Data Verification Failed, Empty")
		} //end if
		if(len(unarchData) != len(data)) {
			return nil, smart.NewError("Compressed Data Verification Failed, Length")
		} //end if
		if(string(smart.Sh3aByt512B64(unarchData)) != string(smart.Sh3aByt512B64(data))) {
			return nil, smart.NewError("Compressed Data Verification Failed, Checksum")
		} //end if
		if(smart.BytesEqual(unarchData, data) != true) {
			return nil, smart.NewError("Compressed Data Verification Failed, Data")
		} //end if
	} //end if
	//--
	return byts, nil
	//--
} //END FUNCTION


func XzUncompress(data []byte) ([]byte, error) { // uncompress xz ; file extension: .xz
	//--
	defer smart.PanicHandler()
	//--
	if(data == nil) {
		return nil, smart.NewError("Input Data is Empty")
	} //end if
	//--
	b := bytes.NewReader(data)
	//--
	var level int = 6 // xz default compression
	//--
	xzCfg := xz.ReaderConfig{ // this was get from gxz go utility, xz/cmd/gxz/file.go
		DictCap: 1 << xzDictCapExps(uint(level)),
	}
	r, errInit := xzCfg.NewReader(b)
	if(errInit != nil) {
		return nil, smart.NewError("Uncompress Init Failed: " + errInit.Error())
	} //end if
	//--
	byts, errRd := io.ReadAll(r)
	if(errRd != nil) {
		return nil, smart.NewError("Uncompress Read Failed: " + errRd.Error())
	} //end if
	//--
	if(byts == nil) {
		return nil, smart.NewError("Uncompressed Data is Empty")
	} //end if
	//--
	return byts, nil
	//--
} //END FUNCTION


//-----


func XzStreamCompress(rdStream io.ReadCloser, level int, checksumMode string, wrStreams ...io.WriteCloser) error { // compress xz ; file extension: .xz
	//--
	// the purpose for accepting multiple writers is to allow for multiple outputs (for example a file and a hash)
	//--
	defer smart.PanicHandler()
	//--
	if(rdStream == nil) {
		return smart.NewError("Input Stream is Null")
	} //end if
	if(wrStreams == nil) {
		return smart.NewError("Output Stream is Null")
	} //end if
	if(len(wrStreams) < 1) {
		return smart.NewError("Output Streams is Empty")
	} //end if
	//--
	for i:=0; i<len(wrStreams); i++ {
		if(wrStreams[i] == nil) {
			return smart.NewError("Output Stream[" + smart.ConvertIntToStr(i) + "] is Null")
		} //end if
	} //end for
	mw := iox.MultiWriteCloser(wrStreams...)
	defer mw.Close()
	//--
	defer rdStream.Close()
	//--
	if((level < 0) || (level > 9)) { // {{{SYNC-XZ-LEVEL-CHECK}}}
		level = 6 // xz default compression
	} //end if
	//--
	var xzChecksumMode byte = xz.CRC64 // default checksum mode in xz
	switch(smart.StrToUpper(checksumMode)) {
		case "SHA256":
			xzChecksumMode = xz.SHA256
			break
		case "CRC32":
			xzChecksumMode = xz.CRC32
			break
		case "CRC64":
		//	xzChecksumMode = xz.CRC64
			break
		default:
			return smart.NewError("Unsupported Checksum Mode: `" + checksumMode + "`")
	} //end switch
	//--
	xzCfg := xz.WriterConfig{ // this was get from gxz go utility, xz/cmd/gxz/file.go
		DictCap: 1 << xzDictCapExps(uint(level)),
		CheckSum: xzChecksumMode,
	}
	w, errInit := xzCfg.NewWriter(mw)
	if(errInit != nil) {
		return smart.NewError("Compress Init Failed: " + errInit.Error())
	} //end if
	//--
	_, errWr := io.Copy(w, rdStream)
	if(errWr != nil) {
		return smart.NewError("Compress Write Failed: " + errWr.Error())
	} //end if
	//--
	errClose := w.Close()
	if(errClose != nil) {
		return smart.NewError("Compress Close Failed: " + errClose.Error())
	} //end if
	//--
	return nil
	//--
} //END FUNCTION


func XzStreamUncompress(rdStream io.ReadCloser, wrStreams ...io.WriteCloser) error { // uncompress xz ; file extension: .xz
	//--
	// the purpose for accepting multiple writers is to allow for multiple outputs (for example a file and a hash)
	//--
	defer smart.PanicHandler()
	//--
	if(rdStream == nil) {
		return smart.NewError("Input Stream is Null")
	} //end if
	if(wrStreams == nil) {
		return smart.NewError("Output Streams are Null")
	} //end if
	if(len(wrStreams) < 1) {
		return smart.NewError("Output Streams are Empty")
	} //end if
	//--
	for i:=0; i<len(wrStreams); i++ {
		if(wrStreams[i] == nil) {
			return smart.NewError("Output Stream[" + smart.ConvertIntToStr(i) + "] is Null")
		} //end if
	} //end for
	mw := iox.MultiWriteCloser(wrStreams...)
	defer mw.Close()
	//--
	defer rdStream.Close()
	//--
	var level int = 6 // xz default compression
	//--
	xzCfg := xz.ReaderConfig{ // this was get from gxz go utility, xz/cmd/gxz/file.go
		DictCap: 1 << xzDictCapExps(uint(level)),
	}
	r, errInit := xzCfg.NewReader(rdStream)
	if(errInit != nil) {
		return smart.NewError("Uncompress Init Failed: " + errInit.Error())
	} //end if
	//--
	_, errWr := io.Copy(mw, r)
	if(errWr != nil) {
		return smart.NewError("Uncompress Write Failed: " + errWr.Error())
	} //end if
	//--
	return nil
	//--
} //END FUNCTION


//-----


func TarXzStreamCompressDir(srcDir string, allowAbsolutePath bool, level int, checksumMode string, wrStreams ...io.WriteCloser) error {
	//--
	// the purpose for accepting multiple writers is to allow for multiple outputs (for example a file and a hash)
	//--
	defer smart.PanicHandler()
	//--
	// use a buffered channel below, maybe there are cases when the sender and receiver are not ready simultaneously, to avoid fatal error as: `fatal error: all goroutines are asleep - deadlock!`
	// it works with boths on tests (buffered or unbuffered), but perhaps in other situations with different readers/writeres there are lags ...
	// if a Go channel is created without the second parameter, it is an unbuffered channel (zero capacity)
	// unbuffered channels require both the sender and the receiver to be ready simultaneously
	// for buffered channels the capacity should be set to the number of workers ; when capacity is full the channel will block, being fatal error
//	errChan := make(chan error)    // unbuffered channel, synchronous
	errChan := make(chan error, 1) // buffered channel, asynchronous with capacity=1 (capacity should be set to number of expected messages from workers, only have 1 async go routine below, with 1 message)
	defer close(errChan)
	//--
	pipeR, pipeW := io.Pipe()
	//--
	go func() {
		errTar := TarStreamCompressDir(srcDir, allowAbsolutePath, pipeW)
		if(errTar != nil) {
			errChan <- smart.NewError("Tar Stream Compress Dir Failed: " + errTar.Error()) // #1
			return
		} //end if
		errChan <- nil // #1
		return
	}()
	//--
	errXz := XzStreamCompress(pipeR, level, checksumMode, wrStreams...)
	if(errXz != nil) {
		return smart.NewError("Xz Stream Compress Failed: " + errXz.Error())
	} //end if
	//--
	for i:=0; i<len(errChan)+1; i++ { // needs len(errChan) + 1, unstandard situation ; this is because also an unbuffered channel (having capacity=0) have to be avle to read ...
		if err := <-errChan; err != nil {
			return err
		} //end if
	} //end for
	//--
	return nil
	//--
} //END FUNCTION


func TarXzStreamUncompressDir(dstDir string, allowAbsolutePath bool, rdStream io.ReadCloser, preserveFileChmod bool) error {
	//--
	defer smart.PanicHandler()
	//--
	// use a buffered channel below, maybe there are cases when the sender and receiver are not ready simultaneously, to avoid fatal error as: `fatal error: all goroutines are asleep - deadlock!`
	// it works with boths on tests (buffered or unbuffered), but perhaps in other situations with different readers/writeres there are lags ...
	// if a Go channel is created without the second parameter, it is an unbuffered channel (zero capacity)
	// unbuffered channels require both the sender and the receiver to be ready simultaneously
	// for buffered channels the capacity should be set to the number of workers ; when capacity is full the channel will block, being fatal error
//	errChan := make(chan error)    // unbuffered channel, synchronous
	errChan := make(chan error, 1) // buffered channel, asynchronous with capacity=1 (capacity should be set to number of expected messages from workers, only have 1 async go routine below, with 1 message)
	defer close(errChan)
	//--
	pipeR, pipeW := io.Pipe()
	//--
	go func() {
		errXz := XzStreamUncompress(rdStream, pipeW)
		if(errXz != nil) {
			errChan <- smart.NewError("Xz Stream Uncompress Failed: " + errXz.Error()) // #1
			return
		} //end if
		errChan <- nil // #1
		return
	}()
	//--
	errTar := TarStreamUncompressDir(dstDir, allowAbsolutePath, pipeR, preserveFileChmod)
	if(errTar != nil) {
		return smart.NewError("Tar Stream Uncompress Dir Failed: " + errTar.Error())
	} //end if
	//--
	for i:=0; i<len(errChan)+1; i++ { // needs len(errChan) + 1, unstandard situation ; this is because also an unbuffered channel (having capacity=0) have to be avle to read ...
		if err := <-errChan; err != nil {
			return err
		} //end if
	} //end for
	//--
	return nil
	//--
} //END FUNCTION


//-----


// #END
