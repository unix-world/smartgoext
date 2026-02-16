package qr

import (
	"fmt"

	"github.com/unix-world/smartgoext/pdf/barcode/utils"
)

func encodeAuto(content string, ecl ErrorCorrectionLevel) (*utils.BitList, *versionInfo, error) {
	bits, vi, _ := Numeric.getEncoder()(content, ecl)
	if bits != nil && vi != nil {
		return bits, vi, nil
	}
	bits, vi, _ = AlphaNumeric.getEncoder()(content, ecl)
	if bits != nil && vi != nil {
		return bits, vi, nil
	}
	bits, vi, _ = Unicode.getEncoder()(content, ecl)
	if bits != nil && vi != nil {
		return bits, vi, nil
	}
	//-- fix by unixman, from upstream, 20250723 @ ea5ac7e13561f6334938261321e13a725d1c0180
//	return nil, nil, fmt.Errorf("No encoding found to encode \"%s\"", content)
	return nil, nil, fmt.Errorf("no encoding found to encode %q", content)
	//-- #
}
