
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2021-present, unix-world.org
// r.20260805.2358 :: STABLE
// [ ARCHIVERS / TAR ]

// REQUIRE: go 1.22 or later
package smartgoext

import (
	"io"
	"strings"
	"path/filepath"
	"os"

	smart "github.com/unix-world/smartgo"
	"github.com/unix-world/smartgo/utils/iox"

	"github.com/unix-world/smartgoext/compress/tar"
)


//-----


func TarStreamCompressDir(srcDir string, allowAbsolutePath bool, wrStreams ...io.WriteCloser) error {
	//--
	// based on: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07 # license MIT
	// Tar takes a source and variable writers and walks 'source' writing each file found to the tar writer
	//--
	// the purpose for accepting multiple writers is to allow for multiple outputs (for example a file and a hash)
	//--
	defer smart.PanicHandler()
	//--
	if(smart.StrTrimWhitespaces(srcDir) == "") {
		return smart.NewError("WARNING: Dir Path is Empty")
	} //end if
	srcDir = smart.SafePathFixClean(srcDir)
	if(smart.PathIsEmptyOrRoot(srcDir) == true) {
		return smart.NewError("WARNING: Dir Path is Empty/Root")
	} //end if
	if(smart.PathIsSafeValidPath(srcDir) != true) {
		return smart.NewError("WARNING: Dir Path is Invalid Unsafe")
	} //end if
	if(smart.PathIsBackwardUnsafe(srcDir) == true) {
		return smart.NewError("WARNING: Dir Path is Backward Unsafe")
	} //end if
	if(allowAbsolutePath != true) {
		if(smart.PathIsAbsolute(srcDir) == true) {
			return smart.NewError("NOTICE: Dir Path is Absolute but not allowed to be absolute by the calling parameters")
		} //end if
	} //end if
	//--
	if(wrStreams == nil) {
		return smart.NewError("Output Streams is Null")
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
	fd, err := os.Stat(srcDir)
	if(err != nil) { // ensure the srcDir actually exists before trying to tar it
		return smart.NewError("Tar Base Directory Stat Failed: " + err.Error())
	} //end if
	if(fd.IsDir() != true) {
		return smart.NewError("Tar Base Directory Stat Failed, Not A Directory")
	} //end if
	//--
	tw := tar.NewWriter(mw)
	defer tw.Close()
	//--
	return filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error { // walk path
		//--
		if(err != nil) {
			return err // return on any error
		} //end if
		//--
		if(!fi.Mode().IsRegular()) {
			return nil // return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		} //end if
		//--
		header, errInfHdr := tar.FileInfoHeader(fi, fi.Name()) // create a new dir/file header
		if(errInfHdr != nil) {
			return errInfHdr
		} //end if
		header.Format = tar.FormatPAX // as in OpenBSD
		//--
		if(smart.StrTrimWhitespaces(file) == "") {
			return smart.NewError("WARNING: Target Tar Path is Empty")
		} //end if
		if(allowAbsolutePath != true) {
			if(smart.PathIsAbsolute(file) == true) {
				return smart.NewError("WARNING: Target Tar Path is Absolute: `" + file + "`")
			} //end if
		} //end if
		file = smart.SafePathFixClean(file)
		if(smart.PathIsEmptyOrRoot(file) == true) {
			return smart.NewError("WARNING: Target Tar Path is Empty/Root: `" + file + "`")
		} //end if
		if(smart.PathIsSafeValidPath(file) != true) {
			return smart.NewError("WARNING: Target Tar Path is Invalid Unsafe: `" + file + "`")
		} //end if
		if(smart.PathIsBackwardUnsafe(file) == true) {
			return smart.NewError("WARNING: Target Tar Path is Backward Unsafe: `" + file + "`")
		} //end if
		//--
		header.Name = strings.TrimPrefix(strings.Replace(file, srcDir, "", -1), string(filepath.Separator)) // update the name to correctly reflect the desired destination when untaring
		errHdr := tw.WriteHeader(header)
		if(errHdr != nil) { // write the header
			return errHdr
		} //end if
		//--
		f, errFOpen := os.Open(file) // open files for taring (read-only mode)
		if(errFOpen != nil) {
			return errFOpen
		} //end if
		//--
		_, errCopy := io.Copy(tw, f)
		if(errCopy != nil) { // copy file data into tar writer
			return errCopy
		} //end if
		//--
		f.Close() // manually close here after each file operation ; defering would cause each file close to wait until all operations have completed.
		//--
		return nil
		//--
	}) //end fx
	//--
} //END FUNCTION


func TarStreamUncompressDir(dstDir string, allowAbsolutePath bool, rdStream io.ReadCloser, preserveFileChmod bool) error {
	//--
	// based on: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07 # license MIT
	// Untar takes a destination path and a reader; a tar reader loops over the tarfile
	// creating the file structure at 'dst' along the way, and writing any files
	//--
	defer smart.PanicHandler()
	//--
	if(smart.StrTrimWhitespaces(dstDir) == "") {
		return smart.NewError("WARNING: Dir Path is Empty")
	} //end if
	dstDir = smart.SafePathFixClean(dstDir)
	if(smart.PathIsEmptyOrRoot(dstDir) == true) {
		return smart.NewError("WARNING: Dir Path is Empty/Root")
	} //end if
	if(smart.PathIsSafeValidPath(dstDir) != true) {
		return smart.NewError("WARNING: Dir Path is Invalid Unsafe")
	} //end if
	if(smart.PathIsBackwardUnsafe(dstDir) == true) {
		return smart.NewError("WARNING: Dir Path is Backward Unsafe")
	} //end if
	if(allowAbsolutePath != true) {
		if(smart.PathIsAbsolute(dstDir) == true) {
			return smart.NewError("NOTICE: Dir Path is Absolute but not allowed to be absolute by the calling parameters")
		} //end if
	} //end if
	//--
	if(smart.PathExists(dstDir) != true) {
		isSuccess, errMsg := smart.SafePathDirCreate(dstDir, true, allowAbsolutePath) // recursive
		if(errMsg != nil) {
			return smart.NewError("WARNING: Dir Create Failed: " + errMsg.Error())
		} //end if
		if(isSuccess != true) {
			return smart.NewError("WARNING: Dir Create Error")
		} //end if
	} //end if
	if(smart.PathIsDir(dstDir) != true) {
		return smart.NewError("WARNING: Dir Path does Not Exists")
	} //end if
	//--
	if(rdStream == nil) {
		return smart.NewError("Input Stream is Null")
	} //end if
	//--
	defer rdStream.Close()
	//--
	tr := tar.NewReader(rdStream)
	//--
	for {
		//--
		header, err := tr.Next()
		//--
		switch {
			case err == io.EOF: // if no more files are found return
				return nil
			case err != nil: // return any other error
				return err
			case header == nil: // if the header is nil, just skip it (not sure how this happens)
				continue
		} //end switch
		//--
		var target string = header.Name
		if(smart.StrTrimWhitespaces(target) == "") {
			return smart.NewError("WARNING: Target Path is Empty")
		} //end if
		if(smart.PathIsAbsolute(target) == true) {
			return smart.NewError("WARNING: Target Path is Absolute: `" + target + "`")
		} //end if
		target = filepath.Join(dstDir, target) // the target location where the dir/file should be created
		target = smart.SafePathFixClean(target)
		if(smart.PathIsEmptyOrRoot(target) == true) {
			return smart.NewError("WARNING: Target Path is Empty/Root: `" + target + "`")
		} //end if
		if(smart.PathIsSafeValidPath(target) != true) {
			return smart.NewError("WARNING: Target Path is Invalid Unsafe: `" + target + "`")
		} //end if
		if(smart.PathIsBackwardUnsafe(target) == true) {
			return smart.NewError("WARNING: Target Path is Backward Unsafe: `" + target + "`")
		} //end if
		//--
		//fi := header.FileInfo() // this switch could also be done using fi.Mode(), not sure if there a benefit of using one vs. the other.
		//--
		switch(header.Typeflag) { // check the file type
			case tar.TypeDir: // if its a dir and it doesn't exist create it
				trFd, err := os.Stat(target)
				if(err != nil) { // destination does not exists
					err := os.MkdirAll(target, smart.CHMOD_DIRS)
					if(err != nil) {
						return err
					} //end if
				} else if(trFd.IsDir() != true) {
					return smart.NewError("Tar Failed: Dir Destination is Not a Directory: `" + target + "`")
				} //end if
			case tar.TypeReg: // if it's a file create it
				trFd, err := os.Stat(target)
				if(err == nil) {
					if(trFd.IsDir() == true) {
						return smart.NewError("Tar Failed: File Destination is a Directory: `" + target + "`")
					} //end if
				} //end if
				dirOfFile := smart.PathDirName(target)
				if(smart.PathExists(dirOfFile) != true) {
					isSuccess, errMsg := smart.SafePathDirCreate(dirOfFile, true, allowAbsolutePath) // recursive
					if(errMsg != nil) {
						return smart.NewError("Tar Failed: Dir Destination Failed to be Created for File: `" + target + "` as `" + dirOfFile + "`: " + errMsg.Error())
					} //end if
					if(isSuccess != true) {
						return smart.NewError("Tar Failed: Dir Destination Failed to be Created for File: `" + target + "` as `" + dirOfFile + "`: ERROR")
					} //end if
				} //end if
				fileChmod := os.FileMode(header.Mode)
				if(preserveFileChmod == false) {
					fileChmod = smart.CHMOD_FILES
				} //end if
			//	f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, fileChmod)
				f, err := os.OpenFile(target, os.O_EXCL|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileChmod) // os.O_EXCL: ensure file does not exists !
				if(err != nil) {
					return err
				} //end if
				//--
				_, errCopy := io.Copy(f, tr)
				if(errCopy != nil) { // copy over contents
					return errCopy
				} //end if
				//--
				f.Close() // manually close here after each file operation ; defering would cause each file close to wait until all operations have completed.
				//--
		} //end switch
		//--
	} //end for
	//--
} //END FUNCTION


//-----


func TarGzStreamCompressDir(srcDir string, allowAbsolutePath bool, level int, wrStreams ...io.WriteCloser) error {
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
	errGz := smart.GzStreamEncode(pipeR, level, wrStreams...)
	if(errGz != nil) {
		return smart.NewError("Gz Stream Compress Failed: " + errGz.Error())
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


func TarGzStreamUncompressDir(dstDir string, allowAbsolutePath bool, rdStream io.ReadCloser, preserveFileChmod bool) error {
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
		errGz := smart.GzStreamDecode(rdStream, pipeW)
		if(errGz != nil) {
			errChan <- smart.NewError("Gz Stream Uncompress Failed: " + errGz.Error()) // #1
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
