
// smartgo ext :: BarCode Extras
// (c) 2026-present unix-world.org
// License: BSD
// r.20260202.2358 :: STABLE

// by unixman

package extras

import (
	"fmt"
	"image"
	"image/color"

	smart "github.com/unix-world/smartgo"
)


func ExtractBarcodeImgDimensionsAndFilledRegions(barcodeImg image.Image, inversed bool) (error, []bool, int, int) {
	//--
	defer smart.PanicHandler() // various
	//--
	if(barcodeImg == nil) {
		return smart.NewError("Image is Null"), nil, 0, 0
	} //end if
	//--
	var w int = barcodeImg.Bounds().Max.X
	var h int = barcodeImg.Bounds().Max.Y
	//--
	if(w <= 0) {
		return smart.NewError("Image Width is Zero"), nil, w, h
	} //end if
	if(h <= 0) {
		return smart.NewError("Image Height is Zero"), nil, w, h
	} //end if
	//--
	var fillMap []bool = []bool{}
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			isFilled, errFilled := isColorRegionFilled(barcodeImg.At(x, y))
			if(errFilled != nil) {
				return smart.NewError("ERR: " + errFilled.Error()), nil, w, h
			} //end if
			fillMap = append(fillMap, isFilled)
		} //end for
	} //end for
	//--
	if(inversed == true) { // by example, DataMatrix needs to be inversed ...
		var invFillMap []bool = []bool{}
		for i:=len(fillMap)-1; i>=0; i-- {
			invFillMap = append(invFillMap, fillMap[i])
		} //end for
		fillMap = invFillMap
		invFillMap = nil // free mem
	} //end if
	//--
	if(len(fillMap) <= 0) {
		return smart.NewError("Fill Map is Empty"), nil, w, h
	} //end if
	if(len(fillMap) != (w * h)) {
		return smart.NewError("Fill Map have an Invalid Dimension"), nil, w, h
	} //end if
	//--
	return nil, fillMap, w, h
	//--
} //END FUNCTION


func isColorRegionFilled(c color.Color) (bool, error) {
	//--
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	//--
	var clr string = smart.StrToLower(smart.StrTrimWhitespaces(fmt.Sprintf("#%.2x%.2x%.2x", rgba.R, rgba.G, rgba.B)))
	//--
	var isFilled bool = false
	var errFilled error = nil
	switch(clr) {
		case "#ffffff":
			isFilled = false
			break
		case "#000000":
			isFilled = true
			break
		default:
			errFilled = smart.NewError("Unknown Color Region: `" + clr + "`")
	} //end switch
	//--
	return isFilled, errFilled
	//--
} //END FUNCTION


// #end
