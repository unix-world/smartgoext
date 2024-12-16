/*
Package fpdi wraps the fpdi PDF library to import existing PDFs as templates. See github.com/unix-world/smartgoext/pdf/fpdi
for further information and examples.

Users should call NewImporter() to obtain their own Importer instance to work with.
To retain backwards compatibility, the package offers a default Importer that may be used via global functions. Note
however that use of the default Importer is not thread safe.
*/

// v.20241215.1258
// (c) unix-world.org
// license: BSD

package fpdi

import (
	"io"

	gofpdi "github.com/unix-world/smartgoext/pdf/fpdi"
)

// fpdiPdf is a partial interface that only implements the functions we need
// from the PDF generator to put the imported PDF templates on the PDF.
type fpdiPdf interface {
	ImportObjects(objs map[string][]byte)
	ImportObjPos(objs map[string]map[int]string)
	ImportTemplates(tpls map[string]string)
	UseImportedTemplate(tplName string, x float64, y float64, w float64, h float64)
	SetError(err error)
}

// Importer wraps an Importer from the fpdi library.
type Importer struct {
	fpdi *gofpdi.Importer
}

// NewImporter creates a new Importer wrapping functionality from the fpdi library.
func NewImporter() *Importer {
	return &Importer{
		fpdi: gofpdi.NewImporter(),
	}
}

// ImportPage imports a page of a PDF file with the specified box (/MediaBox,
// /TrimBox, /ArtBox, /CropBox, or /BleedBox). Returns a template id that can
// be used with UseImportedTemplate to draw the template onto the page.
func (i *Importer) ImportPage(f fpdiPdf, sourceFile string, pageno int, box string) int {
	// Set source file for fpdi
	i.fpdi.SetSourceFile(sourceFile)
	// return template id
	return i.getTemplateID(f, pageno, box)
}

// ImportPageFromStream imports a page of a PDF with the specified box
// (/MediaBox, TrimBox, /ArtBox, /CropBox, or /BleedBox). Returns a template id
// that can be used with UseImportedTemplate to draw the template onto the
// page.
func (i *Importer) ImportPageFromStream(f fpdiPdf, rs *io.ReadSeeker, pageno int, box string) int {
	// Set source stream for fpdi
	i.fpdi.SetSourceStream(rs)
	// return template id
	return i.getTemplateID(f, pageno, box)
}

func (i *Importer) getTemplateID(f fpdiPdf, pageno int, box string) int {
	// Import page
	tpl := i.fpdi.ImportPage(pageno, box)

	// Import objects into current pdf document
	// Unordered means that the objects will be returned with a sha1 hash instead of an integer
	// The objects themselves may have references to other hashes which will be replaced in ImportObjects()
	tplObjIDs := i.fpdi.PutFormXobjectsUnordered()

	// Set template names and ids (hashes) in gofpdf
	f.ImportTemplates(tplObjIDs)

	// Get a map[string]string of the imported objects.
	// The map keys will be the ID of each object.
	imported := i.fpdi.GetImportedObjectsUnordered()

	// Import fpdi objects into gofpdf
	f.ImportObjects(imported)

	// Get a map[string]map[int]string of the object hashes and their positions within each object,
	// to be replaced with object ids (integers).
	importedObjPos := i.fpdi.GetImportedObjHashPos()

	// Import fpdi object hashes and their positions into gopdf
	f.ImportObjPos(importedObjPos)

	return tpl
}

// UseImportedTemplate draws the template onto the page at x,y. If w is 0, the
// template will be scaled to fit based on h. If h is 0, the template will be
// scaled to fit based on w.
func (i *Importer) UseImportedTemplate(f fpdiPdf, tplid int, x float64, y float64, w float64, h float64) {
	// Get values from fpdi
	tplName, scaleX, scaleY, tX, tY := i.fpdi.UseTemplate(tplid, x, y, w, h)

	f.UseImportedTemplate(tplName, scaleX, scaleY, tX, tY)
}

// GetPageSizes returns page dimensions for all pages of the imported pdf.
// Result consists of map[<page number>]map[<box>]map[<dimension>]<value>.
// <page number>: page number, note that page numbers start at 1
// <box>: box identifier, e.g. "/MediaBox"
// <dimension>: dimension string, either "w" or "h"
func (i *Importer) GetPageSizes() map[int]map[string]map[string]float64 {
	return i.fpdi.GetPageSizes()
}

// Default Importer used by global functions
var fpdi = NewImporter()

// ImportPage imports a page of a PDF file with the specified box (/MediaBox,
// /TrimBox, /ArtBox, /CropBox, or /BleedBox). Returns a template id that can
// be used with UseImportedTemplate to draw the template onto the page.
// Note: This uses the default Importer. Call NewImporter() to obtain a custom Importer.
func ImportPage(f fpdiPdf, sourceFile string, pageno int, box string) int {
	return fpdi.ImportPage(f, sourceFile, pageno, box)
}

// ImportPageFromStream imports a page of a PDF with the specified box
// (/MediaBox, TrimBox, /ArtBox, /CropBox, or /BleedBox). Returns a template id
// that can be used with UseImportedTemplate to draw the template onto the
// page.
// Note: This uses the default Importer. Call NewImporter() to obtain a custom Importer.
func ImportPageFromStream(f fpdiPdf, rs *io.ReadSeeker, pageno int, box string) int {
	return fpdi.ImportPageFromStream(f, rs, pageno, box)
}

// UseImportedTemplate draws the template onto the page at x,y. If w is 0, the
// template will be scaled to fit based on h. If h is 0, the template will be
// scaled to fit based on w.
// Note: This uses the default Importer. Call NewImporter() to obtain a custom Importer.
func UseImportedTemplate(f fpdiPdf, tplid int, x float64, y float64, w float64, h float64) {
	fpdi.UseImportedTemplate(f, tplid, x, y, w, h)
}

// GetPageSizes returns page dimensions for all pages of the imported pdf.
// Result consists of map[<page number>]map[<box>]map[<dimension>]<value>.
// <page number>: page number, note that page numbers start at 1
// <box>: box identifier, e.g. "/MediaBox"
// <dimension>: dimension string, either "w" or "h"
// Note: This uses the default Importer. Call NewImporter() to obtain a custom Importer.
func GetPageSizes() map[int]map[string]map[string]float64 {
	return fpdi.GetPageSizes()
}
