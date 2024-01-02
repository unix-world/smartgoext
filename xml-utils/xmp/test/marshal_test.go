// Copyright (c) 2017-2018 Alexander Eichhorn
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/unix-world/smartgoext/xml-utils/xmp/models"
	"github.com/unix-world/smartgoext/xml-utils/xmp/xmp"
)

var (
	buf       []byte
	testfiles []string
)

// Testcases
//

func init() {
	testfiles, _ = filepath.Glob("../samples/*.xmp")
	xmp.SetLogLevel(xmp.LogLevelError)
}

func TestUnmarshalXMP(T *testing.T) {
	for _, v := range testfiles {
		f, err := os.Open(v)
		if err != nil {
			T.Logf("Cannot open sample '%s': %v", v, err)
			continue
		}
		d := xmp.NewDecoder(f)
		doc := &xmp.Document{}
		if err := d.Decode(doc); err != nil {
			T.Errorf("%s: %v", v, err)
		}
		f.Close()
		doc.Close()
	}
}

func TestMarshalXMP(T *testing.T) {
	for _, v := range testfiles {
		f, err := os.Open(v)
		if err != nil {
			T.Logf("Cannot open sample '%s': %v", v, err)
			continue
		}
		d := xmp.NewDecoder(f)
		doc := &xmp.Document{}
		if err := d.Decode(doc); err != nil {
			T.Errorf("%s: %v", v, err)
		}
		f.Close()
		buf, err := xmp.Marshal(doc)
		if err != nil {
			T.Errorf("%s: %v", v, err)
		} else if len(buf) == 0 {
			T.Errorf("%s: empty result", v)
		}
		doc.Close()
	}
}

func TestMarshalJSON(T *testing.T) {
	for _, v := range testfiles {
		f, err := os.Open(v)
		if err != nil {
			T.Logf("Cannot open sample '%s': %v", v, err)
			continue
		}
		d := xmp.NewDecoder(f)
		doc := &xmp.Document{}
		if err := d.Decode(doc); err != nil {
			T.Errorf("%s: %v", v, err)
		}
		f.Close()
		buf, err = json.MarshalIndent(d, "", "  ")
		if err != nil {
			T.Errorf("%s: %v", v, err)
		} else if len(buf) == 0 {
			T.Errorf("%s: empty result", v)
		}
		doc.Close()
	}
}

func TestXmpRoundtrip(T *testing.T) {
	for _, v := range testfiles {
		f, err := os.Open(v)
		if err != nil {
			T.Logf("Cannot open sample '%s': %v", v, err)
			continue
		}
		d := xmp.NewDecoder(f)
		doc := &xmp.Document{}
		if err := d.Decode(doc); err != nil {
			T.Errorf("%s: %v", v, err)
		}
		f.Close()
		buf, err := xmp.Marshal(doc)
		if err != nil {
			T.Errorf("%s: %v", v, err)
			continue
		} else if len(buf) == 0 {
			T.Errorf("%s: empty result", v)
			continue
		}
		doc2 := &xmp.Document{}
		if err := xmp.Unmarshal(buf, doc2); err != nil {
			T.Errorf("XMP Roundtrip %s: %v", v, err)
		}
		doc.Close()
		doc2.Close()
	}
}

func TestJsonRoundtrip(T *testing.T) {
	for _, v := range testfiles {
		f, err := os.Open(v)
		if err != nil {
			T.Logf("Cannot open sample '%s': %v", v, err)
			continue
		}
		d := xmp.NewDecoder(f)
		doc := &xmp.Document{}
		if err := d.Decode(doc); err != nil {
			T.Errorf("%s: %v", v, err)
		}
		f.Close()
		buf, err = json.MarshalIndent(d, "", "  ")
		if err != nil {
			T.Errorf("%s: %v", v, err)
			continue
		} else if len(buf) == 0 {
			T.Errorf("%s: empty result", v)
			continue
		}
		doc2 := &xmp.Document{}
		if err = json.Unmarshal(buf, doc2); err != nil {
			T.Errorf("JSON Roundtrip %s: %v", v, err)
		}
		doc.Close()
		doc2.Close()
	}
}

// Benchmarks
//
// func BenchmarkUnmarshalXMP(B *testing.B) {
// 	xmp.SetLogLevel(xmp.LogLevelError)
// 	b := []byte(data)
// 	for i := 0; i < B.N; i++ {
// 		d := &xmp.Document{}
// 		if err := xmp.Unmarshal(b, d); err != nil {
// 			B.Fatal(err)
// 		}
// 		d.Close()
// 	}
// }

// func BenchmarkMarshalXMP(B *testing.B) {
// 	xmp.SetLogLevel(xmp.LogLevelError)
// 	d := &xmp.Document{}
// 	if err := xmp.Unmarshal([]byte(data), d); err != nil {
// 		B.Fatal(err)
// 	}
// 	B.ResetTimer()
// 	var (
// 		err error
// 	)
// 	for i := 0; i < B.N; i++ {
// 		buf, err = xmp.Marshal(d)
// 		if err != nil {
// 			B.Fatal(err)
// 		}
// 	}
// 	d.Close()
// }

// func BenchmarkMarshalJSON(B *testing.B) {
// 	xmp.SetLogLevel(xmp.LogLevelError)
// 	d := &xmp.Document{}
// 	if err := xmp.Unmarshal([]byte(data), d); err != nil {
// 		B.Fatal(err)
// 	}
// 	B.ResetTimer()
// 	var (
// 		err error
// 	)
// 	for i := 0; i < B.N; i++ {
// 		buf, err = json.MarshalIndent(d, "", "  ")
// 		if err != nil {
// 			B.Fatal(err)
// 		}
// 	}
// 	d.Close()
// }

// func BenchmarkUnmarshalJSON(B *testing.B) {
// 	xmp.SetLogLevel(xmp.LogLevelError)
// 	d := &xmp.Document{}
// 	if err := xmp.Unmarshal([]byte(data), d); err != nil {
// 		B.Fatal(err)
// 	}
// 	buf, err := json.MarshalIndent(d, "", "  ")
// 	if err != nil {
// 		B.Fatal(err)
// 	}
// 	B.ResetTimer()
// 	for i := 0; i < B.N; i++ {
// 		dx := &xmp.Document{}
// 		if err = json.Unmarshal(buf, dx); err != nil {
// 			B.Fatal(err)
// 		}
// 	}
// 	d.Close()
// }

// const data = `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
// <x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="Adobe XMP Core 5.5-c021 79.155241, 2013/11/25-21:10:40        ">
//  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
//   <rdf:Description rdf:about=""
//     xmlns:xmp="http://ns.adobe.com/xap/1.0/"
//     xmlns:xmpDM="http://ns.adobe.com/xmp/1.0/DynamicMedia/"
//     xmlns:stDim="http://ns.adobe.com/xap/1.0/sType/Dimensions#"
//     xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/"
//     xmlns:stEvt="http://ns.adobe.com/xap/1.0/sType/ResourceEvent#"
//     xmlns:stRef="http://ns.adobe.com/xap/1.0/sType/ResourceRef#"
//     xmlns:dc="http://purl.org/dc/elements/1.1/"
//     xmlns:creatorAtom="http://ns.adobe.com/creatorAtom/1.0/"
//    xmp:CreateDate="2015-03-23T22:18:20-04:00"
//    xmp:ModifyDate="2015-03-23T22:18:29-04:00"
//    xmp:MetadataDate="2015-03-23T22:18:29-04:00"
//    xmp:CreatorTool="Adobe Premiere Pro CC (Macintosh)"
//    xmpDM:startTimeScale="24"
//    xmpDM:startTimeSampleSize="1"
//    xmpDM:videoFrameRate="24.000000"
//    xmpDM:videoFieldOrder="Progressive"
//    xmpDM:videoPixelAspectRatio="1/1"
//    xmpDM:audioSampleRate="48000"
//    xmpDM:audioSampleType="16Int"
//    xmpDM:audioChannelType="Stereo"
//    xmpMM:InstanceID="xmp.iid:5d95d24a-457a-44bd-a92c-0786456548fd"
//    xmpMM:DocumentID="009fd353-c6ed-1e3f-5ae1-c92200000053"
//    xmpMM:OriginalDocumentID="xmp.did:d1ae4cc5-a458-4eb7-9394-30ec546e2528"
//    dc:format="QuickTime">
//    <xmpDM:duration
//     xmpDM:value="203"
//     xmpDM:scale="1/24"/>
//    <xmpDM:altTimecode
//     xmpDM:timeValue="01:02:03:04"
//     xmpDM:timeFormat="24Timecode"/>
//    <xmpDM:videoFrameSize
//     stDim:w="1280"
//     stDim:h="720"
//     stDim:unit="pixel"/>
//    <xmpDM:startTimecode
//     xmpDM:timeFormat="24Timecode"
//     xmpDM:timeValue="00:00:00:00"/>
//    <xmpDM:projectRef
//     xmpDM:type="movie"/>
//    <xmpMM:History>
//     <rdf:Seq>
//      <rdf:li
//       stEvt:action="saved"
//       stEvt:instanceID="a985eefc-d371-6b9c-d554-0f0200000080"
//       stEvt:when="2015-03-23T22:18:29-04:00"
//       stEvt:softwareAgent="Adobe Premiere Pro CC (Macintosh)"
//       stEvt:changed="/"/>
//      <rdf:li
//       stEvt:action="created"
//       stEvt:instanceID="xmp.iid:8c22b83c-cfd3-43fe-b402-2d96d4a3d4d4"
//       stEvt:when="2015-03-23T22:18:20-04:00"
//       stEvt:softwareAgent="Adobe Premiere Pro CC (Macintosh)"/>
//      <rdf:li
//       stEvt:action="saved"
//       stEvt:instanceID="xmp.iid:3f3bd4e9-9370-4b25-bb36-ac7d7184bbd5"
//       stEvt:when="2015-03-23T22:18:29-04:00"
//       stEvt:softwareAgent="Adobe Premiere Pro CC (Macintosh)"
//       stEvt:changed="/"/>
//      <rdf:li
//       stEvt:action="saved"
//       stEvt:instanceID="xmp.iid:5d95d24a-457a-44bd-a92c-0786456548fd"
//       stEvt:when="2015-03-23T22:18:29-04:00"
//       stEvt:softwareAgent="Adobe Premiere Pro CC (Macintosh)"
//       stEvt:changed="/metadata"/>
//     </rdf:Seq>
//    </xmpMM:History>
//    <xmpMM:Ingredients>
//     <rdf:Bag>
//      <rdf:li
//       stRef:instanceID="47a9a75b-3a6b-0990-173c-d5a900000074"
//       stRef:documentID="aeffd160-efa3-71bf-610a-09bd00000047"
//       stRef:fromPart="time:0d2154055680000f254016000000"
//       stRef:toPart="time:0d2154055680000f254016000000"
//       stRef:filePath="CornerKick.mov"
//       stRef:maskMarkers="None"/>
//      <rdf:li
//       stRef:instanceID="47a9a75b-3a6b-0990-173c-d5a900000074"
//       stRef:documentID="aeffd160-efa3-71bf-610a-09bd00000047"
//       stRef:fromPart="time:0d2154055680000f254016000000"
//       stRef:toPart="time:0d2154055680000f254016000000"
//       stRef:filePath="CornerKick.mov"
//       stRef:maskMarkers="None"/>
//     </rdf:Bag>
//    </xmpMM:Ingredients>
//    <xmpMM:Pantry>
//     <rdf:Bag>
//      <rdf:li>
//       <rdf:Description
//        xmp:CreateDate="2014-03-20T21:34:41Z"
//        xmp:ModifyDate="2015-03-23T22:10:43-04:00"
//        xmp:MetadataDate="2015-03-23T22:10:43-04:00"
//        xmpDM:startTimeScale="5000"
//        xmpDM:startTimeSampleSize="200"
//        xmpMM:InstanceID="47a9a75b-3a6b-0990-173c-d5a900000074"
//        xmpMM:DocumentID="aeffd160-efa3-71bf-610a-09bd00000047"
//        xmpMM:OriginalDocumentID="xmp.did:cd1f24ab-d821-4bad-aa79-bdc9baaabe14">
//       <xmpDM:duration
//        xmpDM:value="21200"
//        xmpDM:scale="1/2500"/>
//       <xmpDM:altTimecode
//        xmpDM:timeValue="00:00:00:00"
//        xmpDM:timeFormat="25Timecode"/>
//       <xmpMM:History>
//        <rdf:Seq>
//         <rdf:li
//          stEvt:action="saved"
//          stEvt:instanceID="47a9a75b-3a6b-0990-173c-d5a900000074"
//          stEvt:when="2015-03-23T22:10:43-04:00"
//          stEvt:softwareAgent="Adobe Premiere Pro CC (Macintosh)"
//          stEvt:changed="/"/>
//        </rdf:Seq>
//       </xmpMM:History>
//       </rdf:Description>
//      </rdf:li>
//     </rdf:Bag>
//    </xmpMM:Pantry>
//    <xmpMM:DerivedFrom
//     stRef:instanceID="xmp.iid:8c22b83c-cfd3-43fe-b402-2d96d4a3d4d4"
//     stRef:documentID="xmp.did:8c22b83c-cfd3-43fe-b402-2d96d4a3d4d4"
//     stRef:originalDocumentID="xmp.did:8c22b83c-cfd3-43fe-b402-2d96d4a3d4d4"/>
//    <creatorAtom:windowsAtom
//     creatorAtom:extension=".prproj"
//     creatorAtom:invocationFlags="/L"/>
//    <creatorAtom:macAtom
//     creatorAtom:applicationCode="1347449455"
//     creatorAtom:invocationAppleEvent="1129468018"
//     creatorAtom:posixProjectPath="/Users/alandabul/Documents/Adobe/Premiere Pro/7.0/Untitled.prproj"/>
//   </rdf:Description>
//  </rdf:RDF>
// </x:xmpmeta>

// <?xpacket end="w"?>`
