
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2021-present unix-world.org
// r.20260216.2358 :: STABLE
// [ ARCHIVERS ]

// REQUIRE: go 1.19 or later
package smartgoext

import (
	"log"

	"bytes"
	"io/ioutil"

	"archive/zip"

	smart "github.com/unix-world/smartgo"
)


//-----


func zipArchiveReadFileFromMemory(zf *zip.File) ([]byte, error) {
	//--
	defer smart.PanicHandler() // req. for unzip operations
	//--
	f, errOpen := zf.Open()
	if(errOpen != nil) {
		return nil, errOpen
	} //end if
	//--
	defer f.Close()
	//--
	data, errRead := ioutil.ReadAll(f)
	if(errRead != nil) {
		return nil, errRead
	} //end if
	//--
	return data, nil
	//--
} //END FUNCTION


func UnzipArchive(zipData []byte) (map[string]string, error) {
	//--
	defer smart.PanicHandler() // req. for unzip operations
	//--
	var noFiles map[string]string = map[string]string{}
	//--
	if(zipData == nil) {
		return noFiles, smart.NewError("Zip Archive: Content is Empty")
	} //end if
	//--
	zipReader, errReader := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if(errReader != nil) {
		return noFiles, smart.NewError("Zip Archive: Reader ERR: " + errReader.Error())
	} //end if
	if(len(zipReader.File) <= 0) {
		return noFiles, smart.NewError("Zip Archive: Contains No Readable Files")
	} //end if
	//-- read all the files from zip archive
	files := noFiles
	var fName string = ""
	for key, zipFile := range zipReader.File {
		//--
		fName = smart.StrTrimWhitespaces(zipFile.Name)
		//--
		if(fName == "") {
			//--
			log.Println("[NOTICE]", smart.CurrentFunctionName(), "Zip Archive: Skip Reading a File, has an Empty Name, at Key:", key)
			//--
		} else {
			//--
			if(DEBUG) {
				log.Println("[DEBUG]", smart.CurrentFunctionName(), "Zip Archive: Reading a File:", fName, "at Key:", key)
			} //end if
			//--
			unzippedBytes, err := zipArchiveReadFileFromMemory(zipFile)
			if(err != nil) {
				log.Println("[NOTICE]", smart.CurrentFunctionName(), "Zip Archive: Failed to Unzip a File from:", fName, "at Key:", key)
			} else if(unzippedBytes == nil) {
				log.Println("[NOTICE]", smart.CurrentFunctionName(), "Zip Archive: Unziped File is Empty:", fName, "at Key:", key)
			} else {
				files[fName] = string(unzippedBytes) // this is unzipped file bytes
			} //end if
			//--
		} //end if else
		//--
	} //end for
	//--
	return files, nil
	//--
} //END FUNCTION


//-----


// #END
