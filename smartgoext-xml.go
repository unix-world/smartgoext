
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2020-present unix-world.org
// r.20260111.2358 :: STABLE
// [ XML ]

// REQUIRE: go 1.19 or later
package smartgoext

import (
	"strings"

	smart "github.com/unix-world/smartgo"

	"github.com/unix-world/smartgoext/xml-utils/xml2json"
	"github.com/unix-world/smartgoext/xml-utils/etree"
//	"github.com/unix-world/smartgoext/xml-utils/c14n"
)


//-----


func XmlConvertToJson(xmlData string) (string, error) {
	//--
	defer smart.PanicHandler() // for YAML Parser
	//--
	xml := strings.NewReader(xmlData) // xml is an io.Reader
	json, err := xml2json.Convert(xml)
	if(err != nil) {
		return "", err // returns empty string and the conversion error
	} //end if
	//--
	return json.String(), nil // returns the json as string, no error
	//--
} //END FUNCTION


//-----


func XmlC14NCanonize(xmlData string, subPath string, subNs string, withComments bool, transformMode string, oneLine bool) (error, string) {
	//--
	// transform modes: "C14N10Exc"
	//--
	defer smart.PanicHandler() // for XML Parser
	//--
	if(transformMode != "C14N10Exc") {
		return smart.NewError("Unsupported mode: `" + transformMode + "`"), xmlData
	} //end if
	//--
	if(xmlData == "") {
		return nil, ""
	} //end if
	//--
	var numSpaces int = 4
	if(oneLine == true) {
		numSpaces = 0
	} //end if
	//--
	settingsRd := etree.ReadSettings{
		Permissive: 				true, 	// default is FALSE
		PreserveCData: 				true, 	// default is FALSE
		PreserveDuplicateAttrs: 	true, 	// default is FALSE
		ValidateInput: 				true, 	// default is FALSE ; if set to TRUE there are performance issues, but will ensure a well-formed XML before processing it
	}
	settingsWr := etree.WriteSettings{
		CanonicalEndTags: 			true, 	// default is FALSE
		CanonicalText:    			true, 	// default is FALSE
		CanonicalAttrVal: 			true, 	// default is FALSE
	}
	//--
	iS := etree.NewIndentSettings()
	iS.Spaces =						numSpaces 	// default is 4
	iS.UseTabs = 					false 		// default is FALSE
	iS.UseCRLF = 					false 		// default is FALSE
	iS.PreserveLeafWhitespace = 	false 		// default is FALSE
	iS.SuppressTrailingWhitespace = true 		// default is FALSE
	//--
	doc := etree.NewDocument()
	doc.ReadSettings = settingsRd
	//--
	errRd := doc.ReadFromString(xmlData)
	if(errRd != nil) {
		return smart.NewError("eTree Read XML Failed: " + errRd.Error()), xmlData
	} //end if
	//--
	doc.IndentWithSettings(iS)
	if(oneLine == true) {
		doc.IndentTabs() // safety
		doc.Indent(0)
	} //end if
	//--
	var xmlCanonical string = ""
	var errWr error = nil
	//--
	subPath = smart.StrTrimWhitespaces(subPath)
	if(subPath != "") {
		subPathObj, errSubPathObj := etree.CompilePath(subPath)
		if(errSubPathObj != nil) {
			return smart.NewError("eTree Failed to Compile the XML SubPath[" + subPath + "]: " + errSubPathObj.Error()), xmlData
		} //end if
		elSubPath := doc.FindElementsPath(subPathObj)
		if(len(elSubPath) < 1) {
			return smart.NewError("eTree Failed to Find the XML SubPath[" + subPath + "]"), xmlData
		} //end if
		if(subNs != "") {
			elSubPath[0].CreateAttr("xmlns", subNs)
		} //end if
		subDoc := etree.NewDocumentWithRoot(elSubPath[0])
		subDoc.WriteSettings = settingsWr
		xmlCanonical, errWr = subDoc.WriteToString()
	} else {
		doc.WriteSettings = settingsWr
		xmlCanonical, errWr = doc.WriteToString()
	} //end if else
	if(errWr != nil) {
		return smart.NewError("eTree Write XML Failed: " + errWr.Error()), xmlData
	} //end if
	if(xmlCanonical == "") {
		return smart.NewError("eTree Write XML Failed, is Null"), xmlData
	} //end if
	//--
	if(withComments == false) {
		xmlCanonical = smart.StrRegexReplaceAll(`<!--(.*?)-->`, xmlCanonical, "")
	} //end if
	//--
	if(oneLine == true) {
		xmlCanonical = smart.StrTr(xmlCanonical, map[string]string{"\t":"", "\r":"", "\n":""})
	} //end if
	//--
	return nil, xmlCanonical
	//--
} //END FUNCTION


/*
func XmlC14NCanonize(xmlData string, subPath string, subNs string, withComments bool, transformMode string, oneLine bool) (error, string) {
	//--
	// transform modes: "C14N10Exc", "C14N10", "C14N11", "" (aka none, just standardize)
	//--
	defer smart.PanicHandler() // for XML Parser
	//--
	if(xmlData == "") {
		return nil, ""
	} //end if
	//--
	var numSpaces int = 4
	if(oneLine == true) {
		numSpaces = 0
	} //end if
	//--
	settingsRd := etree.ReadSettings{
		Permissive: 				true, 	// default is FALSE
		PreserveCData: 				true, 	// default is FALSE
		PreserveDuplicateAttrs: 	true, 	// default is FALSE
		ValidateInput: 				true, 	// default is FALSE ; if set to TRUE there are performance issues, but will ensure a well-formed XML before processing it
	}
	settingsWr := etree.WriteSettings{
		CanonicalEndTags: 			false, 	// default is FALSE
		CanonicalText:    			false, 	// default is FALSE
		CanonicalAttrVal: 			false, 	// default is FALSE
	}
	//--
	iS := etree.NewIndentSettings()
	iS.Spaces =						numSpaces 	// default is 4
	iS.UseTabs = 					false 		// default is FALSE
	iS.UseCRLF = 					false 		// default is FALSE
	iS.PreserveLeafWhitespace = 	false 		// default is FALSE
	iS.SuppressTrailingWhitespace = true 		// default is FALSE
	//--
	doc := etree.NewDocument()
	doc.ReadSettings = settingsRd
	doc.WriteSettings = settingsWr
	//--
	errRd := doc.ReadFromString(xmlData)
	if(errRd != nil) {
		return smart.NewError("eTree Read XML Failed: " + errRd.Error()), xmlData
	} //end if
	//--
	if(oneLine != true) {
		doc.IndentWithSettings(iS)
	} //end if
	//--
	var xmlCanonical string = ""
	var errWr error = nil
	//--
	subPath = smart.StrTrimWhitespaces(subPath)
	if(subPath != "") {
		subPathObj, errSubPathObj := etree.CompilePath(subPath)
		if(errSubPathObj != nil) {
			return smart.NewError("eTree Failed to Compile the XML SubPath[" + subPath + "]: " + errSubPathObj.Error()), xmlData
		} //end if
		elSubPath := doc.FindElementsPath(subPathObj)
		if(len(elSubPath) < 1) {
			return smart.NewError("eTree Failed to Find the XML SubPath[" + subPath + "]"), xmlData
		} //end if
		if(subNs != "") {
			elSubPath[0].CreateAttr("xmlns", subNs)
		} //end if
		subDoc := etree.NewDocumentWithRoot(elSubPath[0])
		xmlCanonical, errWr = subDoc.WriteToString()
	} else {
		xmlCanonical, errWr = doc.WriteToString()
	} //end if else
	if(errWr != nil) {
		return smart.NewError("eTree Write XML Failed: " + errWr.Error()), xmlData
	} //end if
	if(xmlCanonical == "") {
		return smart.NewError("eTree Write XML Failed, is Null"), xmlData
	} //end if
	//--
	var canonicalizer c14n.Canonicalizer = nil
	switch(transformMode) {
		case "":
			return nil, xmlCanonical
			break
		case "C14N10Exc":
			if(withComments == true) {
				canonicalizer = c14n.MakeC14N10ExclusiveWithCommentsCanonicalizerWithPrefixList("")
			} else {
				canonicalizer = c14n.MakeC14N10ExclusiveCanonicalizerWithPrefixList("")
			} //end if else
			break
		case "C14N10":
			if(withComments == true) {
				canonicalizer = c14n.MakeC14N10WithCommentsCanonicalizer()
			} else {
				canonicalizer = c14n.MakeC14N10RecCanonicalizer()
			} //end if else
			break
		case "C14N11":
			if(withComments == true) {
				canonicalizer = c14n.MakeC14N11WithCommentsCanonicalizer()
			} else {
				canonicalizer = c14n.MakeC14N11Canonicalizer()
			} //end if else
			break
		default:
			return smart.NewError("Invalid C14N Canonicalizer Transform Mode: `" + transformMode + "`"), xmlCanonical
	} //end switch
	if(canonicalizer == nil) {
		return smart.NewError("C14N Canonicalizer is Null"), xmlCanonical
	} //end if
	//--
	raw := etree.NewDocument()
	err := raw.ReadFromString(xmlCanonical)
	if(err != nil) {
		return smart.NewError("C14N Canonicalizer Failed to Init: " + err.Error()), xmlCanonical
	} //end if
	//--
	xmlC14nCanonical, errC14Canonical := canonicalizer.Canonicalize(raw.Root())
	if(errC14Canonical != nil) {
		return smart.NewError("C14N Canonicalizer Failed to Canonicalize: " + errC14Canonical.Error()), xmlCanonical
	} //end if
	if(xmlC14nCanonical == nil) {
		return smart.NewError("C14N XML is Null"), xmlCanonical
	} //end if
	//--
	return nil, string(xmlC14nCanonical)
	//--
} //END FUNCTION
*/


//-----


// #END
