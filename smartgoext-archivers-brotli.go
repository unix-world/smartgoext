
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2021-present, unix-world.org
// r.20260823.2358 :: STABLE
// [ ARCHIVERS / BROTLI ]

// REQUIRE: go 1.22 or later
package smartgoext

import (
	"bytes"
	"io"

	smart    "github.com/unix-world/smartgo"
	"github.com/unix-world/smartgo/utils/iox"

	"github.com/unix-world/smartgoext/compress/brotli"
)


//----- IMPORTANT: Brotli have no CRC checksums ... it is intended mainly for streams and very fast speeds ; very slow with large data at max quality, use 5 (half quality) for large data, ex: tar


func BrotliCompress(data []byte, mode int, quality int, windowSize int, verifyCompressed bool) ([]byte, error) { // compress brotli compatible ; file extension: .br
	//--
	defer smart.PanicHandler()
	//--
	if(data == nil) {
		return nil, smart.NewError("Input Data is Empty")
	} //end if
	//--
	if((mode < 0) || (mode > 2)) {
		mode = 0 // default to generic mode ; mode=0 (generic mode) ; mode=1 (text mode) ; mode=2 (font mode)
	} //end if
	if((quality < 0) || (quality > 11)) {
		quality = 11 // default quality
	} //end if
	if((windowSize < 0) || (windowSize > 22)) {
		windowSize = 22 // default windows size
	} //end if
	//--
	var buf bytes.Buffer
	//--
	w := brotli.NewWriterOptions(&buf, brotli.WriterOptions{Mode: mode, Quality:quality, LGWin:windowSize})
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
		unarchData, unarchErr := BrotliUncompress(byts)
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


func BrotliUncompress(data []byte) ([]byte, error) { // compress brotli compatible ; file extension: .br
	//--
	defer smart.PanicHandler()
	//--
	if(data == nil) {
		return nil, smart.NewError("Input Data is Empty")
	} //end if
	//--
	b := bytes.NewReader(data)
	//--
	r := brotli.NewReader(b)
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


func BrotliStreamCompress(rdStream io.ReadCloser, mode int, quality int, windowSize int, wrStreams ...io.WriteCloser) error { // compress brotli compatible ; file extension: .br
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
	if((mode < 0) || (mode > 2)) {
		mode = 0 // default to generic mode ; mode=0 (generic mode) ; mode=1 (text mode) ; mode=2 (font mode)
	} //end if
	if((quality < 0) || (quality > 11)) {
		quality = 11 // default quality
	} //end if
	if((windowSize < 0) || (windowSize > 22)) {
		windowSize = 22 // default windows size
	} //end if
	//--
	w := brotli.NewWriterOptions(mw, brotli.WriterOptions{Mode: mode, Quality:quality, LGWin:windowSize})
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


func BrotliStreamUncompress(rdStream io.ReadCloser, wrStreams ...io.WriteCloser) error { // uncompress brotli compatible ; file extension: .br
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
	r := brotli.NewReader(rdStream)
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


// #END
