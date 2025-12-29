// Copyright ©2023 The go-pdf Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// v.20250226.2358
// (c) 2023-present unix-world.org
// license: BSD

package fpdf

import (
//	"log"
	"fmt"

	"crypto/md5"
	"encoding/hex"
//	"strings"

	smart "github.com/unix-world/smartgo"
)

const DEFAULT_MINME_TYPE string = "application/octet-stream" // fix by unixman

// Attachment defines a content to be included in the pdf, in one
// of the following ways :
//   - associated with the document as a whole : see SetAttachments()
//   - accessible via a link localized on a page : see AddAttachmentAnnotation()
type Attachment struct {
	Content []byte

	// Filename is the displayed name of the attachment
	Filename string

	//-- Mime Type for PDF/A compliance ; by unixman
	MimeType string
	MarkAF uint8
	//-- #

	// Description is only displayed when using AddAttachmentAnnotation(),
	// and might be modified by the pdf reader.
	Description string

	objectNumber int // filled when content is included
}

// return the hex encoded checksum of `data`
func checksum(data []byte) string {
	sl := md5.Sum(data)
	return hex.EncodeToString(sl[:])
}

// Writes a compressed file like object as "/EmbeddedFile". Compressing is
// done with deflate. Includes length, compressed length and MD5 checksum.
//func (f *Fpdf) writeCompressedFileObject(content []byte) {
func (f *Fpdf) writeCompressedFileObject(content []byte, mimeType string) { // extend by unixman: PDF/A compliance
	lenUncompressed := len(content)
	sum := checksum(content)
	mem := xmem.compress(content)
	defer mem.release()
	compressed := mem.bytes()
	lenCompressed := len(compressed)
	f.newobj()
	//-- fix by unixman: add mime type ; PDF/A compliance
	mimeType = smart.StrTrimWhitespaces(mimeType)
	if(mimeType == "") {
		mimeType = DEFAULT_MINME_TYPE
	} //end if
	mimeType = f.escapeSmarter(mimeType) // ex: application#2Foctet-stream
	mimeType = smart.StrToLower(mimeType) // conformance

	mod := timeOrNow(f.modDate)
//	f.outf("<< /Type /EmbeddedFile /Length %d /Filter /FlateDecode /Params << /CheckSum <%s> /Size %d >> >>\n", lenCompressed, sum, lenUncompressed)
	f.outf("<< /Type /EmbeddedFile /Subtype /%s /Length %d /Filter /FlateDecode /Params << /ModDate %s /CheckSum <%s> /Size %d >> >>", mimeType, lenCompressed, f.textstring("D:"+mod.Format("20060102150405")), sum, lenUncompressed) // fix by unixman to comply with PDF/A standards
	//--
	f.putstream(compressed)
	f.out("endobj")
}

// Embed includes the content of `a`, and update its internal reference.
func (f *Fpdf) embed(a *Attachment) {
	if a.objectNumber != 0 { // already embedded (objectNumber start at 2)
		return
	}
	oldState := f.state
	f.state = 1 // we write file content in the main buffer
	f.writeCompressedFileObject(a.Content, a.MimeType)
	streamID := f.n
	f.newobj()
	//-- unixman: possible values for AFRelationship: Source, Data, Alternative, Supplement, and Unspecified
//	f.outf("<< /Type /Filespec /F () /UF %s /EF << /F %d 0 R >> /Desc %s\n>>", f.textstring(utf8toutf16(a.Filename)), streamID, f.textstring(utf8toutf16(a.Description)))
	var afRelatedStr string = "/Unspecified"
	switch(a.MarkAF) {
		case 1:
			afRelatedStr = "/Source"
			break
		case 2:
			afRelatedStr = "/Data"
			break
		case 3:
			afRelatedStr = "/Alternative"
			break
		case 4:
			afRelatedStr = "/Supplement"
			break
		default: // unspecified, unknown, unused
			// future use: EncryptedPayload, FormData, Schema
	}
	f.outf("<< /Type /Filespec /F %s /UF %s /AFRelationship " + afRelatedStr + " /EF << /F %d 0 R >> /Desc %s\n>>", f.textstring(a.Filename), f.textstring(utf8toutf16(a.Filename)), streamID, f.textstring(utf8toutf16(a.Description)))
	//--
	f.out("endobj")
	a.objectNumber = f.n
	f.state = oldState
}

// SetAttachments writes attachments as embedded files (document attachment).
// These attachments are global, see AddAttachmentAnnotation() for a link
// anchored in a page. Note that only the last call of SetAttachments is
// useful, previous calls are discarded. Be aware that not all PDF readers
// support document attachments. See the SetAttachment example for a
// demonstration of this method.
func (f *Fpdf) SetAttachments(as []Attachment) {
	f.attachments = as
}

// embed current attachments. store object numbers
// for later use by getEmbeddedFiles()
func (f *Fpdf) putAttachments() {
	for i, a := range f.attachments {
		f.embed(&a)
		f.attachments[i] = a
	}
}

// return /EmbeddedFiles tree name catalog entry.
func (f Fpdf) getEmbeddedFiles() string {
	names := make([]string, len(f.attachments))
	for i, as := range f.attachments {
	//	names[i] = fmt.Sprintf("(Attachement%d) %d 0 R", i+1, as.objectNumber)
		names[i] = fmt.Sprintf("%s %d 0 R", f.textstring(as.Filename), as.objectNumber) // fix by unixman, better compliancy
	}
//	nameTree := fmt.Sprintf("<< /Names [\n%s\n] >>", smart.Implode("\n", names))
	nameTree := fmt.Sprintf("<< /Names [%s] >>", smart.StrTrimWhitespaces(smart.Implode(" ", names))) // unixman
	return smart.StrTrimWhitespaces(nameTree)
}

//-- by unixman, PDF/A
// return /AFEntry
func (f Fpdf) getAFEntries() string {
	var afEntry = ""
	for _, as := range f.attachments {
	//	if(afEntry == "") { // only get the 1st one
	//		if(as.MarkAF > 0) {
				afEntry += fmt.Sprintf("%d 0 R", as.objectNumber) + " " // fix by unixman, better compliancy
	//			break
	//		}
	//	} //end if
	}
	return smart.StrTrimWhitespaces(afEntry)
}
//-- #

// ---------------------------------- Annotations ----------------------------------

type annotationAttach struct {
	*Attachment

	x, y, w, h float64 // fpdf coordinates (y diff and scaling done)
}

// AddAttachmentAnnotation puts a link on the current page, on the rectangle
// defined by `x`, `y`, `w`, `h`. This link points towards the content defined
// in `a`, which is embedded in the document. Note than no drawing is done by
// this method : a method like `Cell()` or `Rect()` should be called to
// indicate to the reader that there is a link here. Requiring a pointer to an
// Attachment avoids useless copies in the resulting pdf: attachment pointing
// to the same data will have their content only be included once, and be
// shared amongst all links. Be aware that not all PDF readers support
// annotated attachments. See the AddAttachmentAnnotation example for a
// demonstration of this method.
func (f *Fpdf) AddAttachmentAnnotation(a *Attachment, x, y, w, h float64) {
	if a == nil {
		return
	}
	f.pageAttachments[f.page] = append(f.pageAttachments[f.page], annotationAttach{
		Attachment: a,
		x:          x * f.k, y: f.hPt - y*f.k, w: w * f.k, h: h * f.k,
	})
}

// embed current annotations attachments. store object numbers
// for later use by putAttachmentAnnotationLinks(), which is
// called for each page.
func (f *Fpdf) putAnnotationsAttachments() {
	// avoid duplication
	m := map[*Attachment]bool{}
	for _, l := range f.pageAttachments {
		for _, an := range l {
			if m[an.Attachment] { // already embedded
				continue
			}
			f.embed(an.Attachment)
		}
	}
}

func (f *Fpdf) putAttachmentAnnotationLinks(out *fmtBuffer, page int) {
	for _, an := range f.pageAttachments[page] {
		x1, y1, x2, y2 := an.x, an.y, an.x+an.w, an.y-an.h
		as := fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [%.2f %.2f %.2f %.2f] /Length 0 >>",
			x1, y1, x2, y2)
		as += "\nstream\nendstream"

		out.printf("<< /Type /Annot /Subtype /FileAttachment /Rect [%.2f %.2f %.2f %.2f] /Border [0 0 0]\n",
			x1, y1, x2, y2)
		out.printf("/Contents %s ", f.textstring(utf8toutf16(an.Description)))
		out.printf("/T %s ", f.textstring(utf8toutf16(an.Filename)))
		out.printf("/AP << /N %s>>", as)
		out.printf("/FS %d 0 R >>\n", an.objectNumber)
	}
}
