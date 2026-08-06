
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2021-present, unix-world.org
// r.20260805.2358 :: STABLE
// [ ARCHIVERS / LZ4 ]

// REQUIRE: go 1.22 or later
package smartgoext

import (
	"bytes"
	"io"
	"runtime"

	smart    "github.com/unix-world/smartgo"
	"github.com/unix-world/smartgo/utils/iox"

	"github.com/unix-world/smartgoext/utils/bytefmt"
	"github.com/unix-world/smartgoext/compress/lz4"
)


//----- IMPORTANT: Lz4 is not very efficient with small strings but it rocks with large data ; is the fastest in the area ...


func Lz4Compress(data []byte, concurrency int, level int, streamChecksum bool, blockChecksum bool, blockMaxSize string, verifyCompressed bool) ([]byte, error) { // compress lz4 compatible ; file extension: .lz4
	//--
	defer smart.PanicHandler()
	//--
	if(data == nil) {
		return nil, smart.NewError("Input Data is Empty")
	} //end if
	//--
	if((concurrency < 0) || (concurrency > runtime.GOMAXPROCS(0))) {
		concurrency = 0 // default (all CPUs)
	} //end if
	//--
	if((level < 0) || (level > 9)) {
		level = -1 // force fallback to default compression
	} //end if
	var lvl lz4.CompressionLevel = lz4.Level1 // LZ4 default level is 1 (lz4.Level1)
	switch level {
		case 0:
			lvl = lz4.Fast
			break
		case 1:
			lvl = lz4.Level1
			break
		case 2:
			lvl = lz4.Level2
			break
		case 3:
			lvl = lz4.Level3
			break
		case 4:
			lvl = lz4.Level4
			break
		case 5:
			lvl = lz4.Level5
			break
		case 6:
			lvl = lz4.Level6
			break
		case 7:
			lvl = lz4.Level7
			break
		case 8:
			lvl = lz4.Level8
			break
		case 9:
			lvl = lz4.Level9
			break
		default:
			// use default
	} //end switch
	//--
	blockMaxSize = smart.StrToUpper(smart.StrTrimWhitespaces(blockMaxSize))
	switch(blockMaxSize) {
		case "64K":  fallthrough
		case "256K": fallthrough
		case "1M":   fallthrough
		case "4M": // maximum value
			break
		case "":
			blockMaxSize = "4M" // maximum value for block size
		default:
			return nil, smart.NewError("Unsupported Block Max Size: `" + blockMaxSize + "`")
	} //end switch
	//--
	safeBlockMaxSz, errBlockMaxSz := bytefmt.ToBytes(blockMaxSize)
	if(errBlockMaxSz != nil) {
		return nil, smart.NewError("Invalid Settings: Max Block Size")
	} //end if
	//--
	var buf bytes.Buffer
	//--
	w := lz4.NewWriter(&buf)
	options := []lz4.Option{
		lz4.CompressionLevelOption(lvl), 					// compression level: 0..9 ; default is 1
		lz4.ConcurrencyOption(concurrency), 				// concurrency, 0 = default (all CPUs)
		lz4.ChecksumOption(streamChecksum), 				// false / true ; default is false
		lz4.BlockChecksumOption(blockChecksum), 			// false / true ; default is false
		lz4.BlockSizeOption(lz4.BlockSize(safeBlockMaxSz)), // block max size [ 64K, 256K, 1M, 4M ] ; default is: 4M
	}
	errOptions := w.Apply(options...)
	if(errOptions != nil) {
		return nil, smart.NewError("Compress Options Error: " + errOptions.Error())
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
		unarchData, unarchErr := Lz4Uncompress(byts, concurrency)
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


func Lz4Uncompress(data []byte, concurrency int) ([]byte, error) { // uncompress lz4 compatible ; file extension: .lz4
	//--
	defer smart.PanicHandler()
	//--
	if(data == nil) {
		return nil, smart.NewError("Input Data is Empty")
	} //end if
	//--
	if((concurrency < 0) || (concurrency > runtime.GOMAXPROCS(0))) {
		concurrency = 0 // default (all CPUs)
	} //end if
	//--
	b := bytes.NewReader(data)
	//--
	r := lz4.NewReader(b)
	options := []lz4.Option{
		lz4.ConcurrencyOption(concurrency), // concurrency, 0 = default (all CPUs)
	}
	errOptions := r.Apply(options...)
	if(errOptions != nil) {
		return nil, smart.NewError("Uncompress Options Error: " + errOptions.Error())
	} //end if
	//--
	byts, errRd := io.ReadAll(r)
	if(errRd != nil) {
		return nil, smart.NewError("Uncompress Read Failed: " + errRd.Error())
	} //end if
	//--
	if(byts == nil) {
		return nil, smart.NewError("Uncompress Data is Empty")
	} //end if
	//--
	return byts, nil
	//--
} //END FUNCTION


//-----


func Lz4StreamCompress(rdStream io.ReadCloser, concurrency int, level int, streamChecksum bool, blockChecksum bool, blockMaxSize string, wrStreams ...io.WriteCloser) error { // compress lz4 compatible ; file extension: .lz4
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
	if((concurrency < 0) || (concurrency > runtime.GOMAXPROCS(0))) {
		concurrency = 0 // default (all CPUs)
	} //end if
	//--
	if((level < 0) || (level > 9)) {
		level = -1 // force fallback to default compression
	} //end if
	var lvl lz4.CompressionLevel = lz4.Level1 // LZ4 default level is 1 (lz4.Level1)
	switch level {
		case 0:
			lvl = lz4.Fast
			break
		case 1:
			lvl = lz4.Level1
			break
		case 2:
			lvl = lz4.Level2
			break
		case 3:
			lvl = lz4.Level3
			break
		case 4:
			lvl = lz4.Level4
			break
		case 5:
			lvl = lz4.Level5
			break
		case 6:
			lvl = lz4.Level6
			break
		case 7:
			lvl = lz4.Level7
			break
		case 8:
			lvl = lz4.Level8
			break
		case 9:
			lvl = lz4.Level9
			break
		default:
			// use default
	} //end switch
	//--
	blockMaxSize = smart.StrToUpper(smart.StrTrimWhitespaces(blockMaxSize))
	switch(blockMaxSize) {
		case "64K":  fallthrough
		case "256K": fallthrough
		case "1M":   fallthrough
		case "4M": // maximum value
			break
		case "":
			blockMaxSize = "4M" // maximum value for block size
		default:
			return smart.NewError("Unsupported Block Max Size: `" + blockMaxSize + "`")
	} //end switch
	//--
	safeBlockMaxSz, errBlockMaxSz := bytefmt.ToBytes(blockMaxSize)
	if(errBlockMaxSz != nil) {
		return smart.NewError("Invalid Settings: Max Block Size")
	} //end if
	//--
	w := lz4.NewWriter(mw)
	options := []lz4.Option{
		lz4.CompressionLevelOption(lvl), 					// compression level: 0..9 ; default is 1
		lz4.ConcurrencyOption(concurrency), 				// concurrency, 0 = default (all CPUs)
		lz4.ChecksumOption(streamChecksum), 				// false / true ; default is false
		lz4.BlockChecksumOption(blockChecksum), 			// false / true ; default is false
		lz4.BlockSizeOption(lz4.BlockSize(safeBlockMaxSz)), // block max size [ 64K, 256K, 1M, 4M ] ; default is: 4M
	}
	errOptions := w.Apply(options...)
	if(errOptions != nil) {
		return smart.NewError("Compress Options Error: " + errOptions.Error())
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


func Lz4StreamUncompress(rdStream io.ReadCloser, concurrency int, wrStreams ...io.WriteCloser) error { // uncompress lz4 compatible ; file extension: .lz4
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
	if((concurrency < 0) || (concurrency > runtime.GOMAXPROCS(0))) {
		concurrency = 0 // default (all CPUs)
	} //end if
	//--
	r := lz4.NewReader(rdStream)
	options := []lz4.Option{
		lz4.ConcurrencyOption(concurrency), // concurrency, 0 = default (all CPUs)
	}
	errOptions := r.Apply(options...)
	if(errOptions != nil) {
		return smart.NewError("Uncompress Options Error: " + errOptions.Error())
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


func TarLz4StreamCompressDir(srcDir string, allowAbsolutePath bool, concurrency int, level int, streamChecksum bool, blockChecksum bool, blockMaxSize string, wrStreams ...io.WriteCloser) error {
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
	errLz4 := Lz4StreamCompress(pipeR, concurrency, level, streamChecksum, blockChecksum, blockMaxSize, wrStreams...)
	if(errLz4 != nil) {
		return smart.NewError("Lz4 Stream Compress Failed: " + errLz4.Error())
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


func TarLz4StreamUncompressDir(dstDir string, allowAbsolutePath bool, rdStream io.ReadCloser, concurrency int, preserveFileChmod bool) error {
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
		errLz4 := Lz4StreamUncompress(rdStream, concurrency, pipeW)
		if(errLz4 != nil) {
			errChan <- smart.NewError("Lz4 Stream Uncompress Failed: " + errLz4.Error()) // #1
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
