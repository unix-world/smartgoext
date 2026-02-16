package sfnt

import (
	"strconv"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
//	"golang.org/x/text/encoding/japanese"
//	"golang.org/x/text/encoding/korean"
//	"golang.org/x/text/encoding/simplifiedchinese"
//	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"

//	"github.com/go-sw/text-codec/apple"
//	johab "github.com/go-sw/text-codec/korean"

)

// PlatformID represents the platform id used to specify a particular character encoding.
type PlatformID uint16

const (
	PlatformUnicode   = PlatformID(0)
	PlatformMac       = PlatformID(1)
	PlatformISO       = PlatformID(2) // Deprecated: as per TrueType specification
	PlatformMicrosoft = PlatformID(3)
	PlatformCustom    = PlatformID(4)
)

// String returns an idenfying string for each platform or "Platform X" for unknown values.
func (p PlatformID) String() string {
	switch p {
	case PlatformUnicode:
		return "Unicode"
	case PlatformMac:
		return "Mac"
	case PlatformISO:
		return "ISO"
	case PlatformMicrosoft:
		return "Microsoft"
	case PlatformCustom:
		return "Custom"
	default:
		return "Platform " + strconv.Itoa(int(p))
	}
}

// PlatformEncodingID represents the platform specific id used to
// specify a particular character encoding.
type PlatformEncodingID uint16

const (
	// Unicode platform

	PlatformEncodingUnicode1_0            = PlatformEncodingID(0) // Deprecated: Unicode 1.0 semantics
	PlatformEncodingUnicode1_1            = PlatformEncodingID(1) // Deprecated: Unicode 1.1 semantics
	PlatformEncodingUnicodeISO10646       = PlatformEncodingID(2) // Deprecated: ISO/IEC 10646 semantics
	PlatformEncodingUnicodeBMP            = PlatformEncodingID(3)
	PlatformEncodingUnicodeFullRepertoire = PlatformEncodingID(4)

	// Macintosh platform (only those actually used)
	// See https://github.com/fonttools/fonttools/issues/236

	PlatformEncodingMacRoman              = PlatformEncodingID(0)
	PlatformEncodingMacJapanese           = PlatformEncodingID(1)
	PlatformEncodingMacChineseTraditional = PlatformEncodingID(2)
	PlatformEncodingMacKorean             = PlatformEncodingID(3)
	PlatformEncodingMacGreek              = PlatformEncodingID(6)
	PlatformEncodingMacRussian            = PlatformEncodingID(7)
	PlatformEncodingMacChineseSimplified  = PlatformEncodingID(25)
	PlatformEncodingMacSlavic             = PlatformEncodingID(29)
	PlatformEncodingMacTurkish            = PlatformEncodingID(35)
	PlatformEncodingMacIceland            = PlatformEncodingID(37)

	// Deprecated: ISO platform
	PlatformEncodingISOASCII  = PlatformEncodingID(0)
	PlatformEncodingISO10646  = PlatformEncodingID(1) // Deprecated: Unicode
	PlatformEncodingISO8859_1 = PlatformEncodingID(2) // Deprecated: ISO 8859-1

	// Microsoft platform

	PlatformEncodingMicrosoftSymbol                = PlatformEncodingID(0)
	PlatformEncodingMicrosoftUnicode               = PlatformEncodingID(1) // Unicode BMP
	PlatformEncodingMicrosoftShiftJIS              = PlatformEncodingID(2)
	PlatformEncodingMicrosoftPRC                   = PlatformEncodingID(3)
	PlatformEncodingMicrosoftBig5                  = PlatformEncodingID(4)
	PlatformEncodingMicrosoftWansung               = PlatformEncodingID(5)
//	PlatformEncodingMicrosoftJohab                 = PlatformEncodingID(6)
	PlatformEncodingMicrosoftUnicodeFullRepertoire = PlatformEncodingID(10)
)

// PlatformLanguageID represents the platform specific language id used to
// specify a particular character encoding.
type PlatformLanguageID uint16

const (
	// Unicode platform language ID

	PlatformLanguageUnicodeDefault = PlatformLanguageID(0)

	// Macintosh platform language IDs (only those actually used)

	PlatformLanguageMacEnglish    = PlatformLanguageID(0)
	PlatformLanguageMacIcelandic  = PlatformLanguageID(15)
	PlatformLanguageMacTurkish    = PlatformLanguageID(17)
	PlatformLanguageMacCroatian   = PlatformLanguageID(18)
	PlatformLanguageMacLithuanian = PlatformLanguageID(24)
	PlatformLanguageMacPolish     = PlatformLanguageID(25)
	PlatformLanguageMacHungarian  = PlatformLanguageID(26)
	PlatformLanguageMacEstonian   = PlatformLanguageID(27)
	PlatformLanguageMacLatvian    = PlatformLanguageID(28)
	PlatformLanguageMacAlbanian   = PlatformLanguageID(36)
	PlatformLanguageMacRomanian   = PlatformLanguageID(37)
	PlatformLanguageMacCzech      = PlatformLanguageID(38)
	PlatformLanguageMacSlovak     = PlatformLanguageID(39)
	PlatformLanguageMacSlovenian  = PlatformLanguageID(40)

	// Microsoft platform language ID

	PlatformLanguageMicrosoftEnglish = PlatformLanguageID(0x0409)
)

// GetEncoding is a best-effort attempt to return the text encoding for a given
// platformID/encodingID/langID, which might result in broken text.
// Returns nil if the encoding is already UTF-8 compatible (e.g. ASCII) or unsupported.
func GetEncoding(platformID PlatformID, encodingID PlatformEncodingID, langID PlatformLanguageID) encoding.Encoding {
	switch platformID {
	case PlatformUnicode:
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
//	case PlatformMac:
//		return getMacEncoding(encodingID, langID)
	case PlatformISO:
		return getISOEncoding(encodingID)
//	case PlatformMicrosoft:
//		return getMicrosoftEncoding(encodingID)
	}

	return nil
}

// getMacEncoding returns the encoding for Mac platform entries.
/*
func getMacEncoding(encodingID PlatformEncodingID, langID PlatformLanguageID) encoding.Encoding {
	switch encodingID {
	case PlatformEncodingMacRoman: // Mac Roman
		switch langID {
		case PlatformLanguageMacIcelandic:
			return apple.Iceland
		case PlatformLanguageMacTurkish:
			return apple.Turkish
		case PlatformLanguageMacCroatian:
			return apple.Croatian
		case PlatformLanguageMacLithuanian, PlatformLanguageMacPolish, PlatformLanguageMacHungarian,
			PlatformLanguageMacEstonian, PlatformLanguageMacLatvian, PlatformLanguageMacAlbanian,
			PlatformLanguageMacCzech, PlatformLanguageMacSlovak, PlatformLanguageMacSlovenian: // mac_latin2
			return apple.CentralEuropean
		case PlatformLanguageMacRomanian:
			return apple.Romanian
		default:
			return charmap.Macintosh
		}

	case PlatformEncodingMacJapanese:
		return apple.Japanese
	case PlatformEncodingMacChineseTraditional:
		return apple.ChineseTraditional
	case PlatformEncodingMacKorean:
		return apple.Korean
	case PlatformEncodingMacGreek:
		return apple.Greek
	case PlatformEncodingMacRussian:
		return charmap.MacintoshCyrillic
	case PlatformEncodingMacChineseSimplified:
		return apple.ChineseSimplified
	case PlatformEncodingMacSlavic: // mac_latin2
		return apple.CentralEuropean
	case PlatformEncodingMacTurkish:
		return apple.Turkish
	case PlatformEncodingMacIceland:
		return apple.Iceland
	}

	return nil
}
*/

// getISOEncoding returns the encoding for ISO platform entries.
func getISOEncoding(encodingID PlatformEncodingID) encoding.Encoding {
	switch encodingID {
	case PlatformEncodingISOASCII:
		return nil // ASCII is valid UTF-8
	case PlatformEncodingISO10646:
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case PlatformEncodingISO8859_1:
		return charmap.ISO8859_1
	}

	return nil
}

/*
// getMicrosoftEncoding returns the encoding for Microsoft platform entries.
func getMicrosoftEncoding(encodingID PlatformEncodingID) encoding.Encoding {
	switch encodingID {
	case PlatformEncodingMicrosoftSymbol:
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case PlatformEncodingMicrosoftUnicode:
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case PlatformEncodingMicrosoftShiftJIS:
		return japanese.ShiftJIS
	case PlatformEncodingMicrosoftPRC:
		return simplifiedchinese.GBK
	case PlatformEncodingMicrosoftBig5:
		return traditionalchinese.Big5
	case PlatformEncodingMicrosoftWansung:
		return korean.EUCKR
//	case PlatformEncodingMicrosoftJohab:
//		return johab.Johab
	case PlatformEncodingMicrosoftUnicodeFullRepertoire:
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	}

	return nil
}
*/
