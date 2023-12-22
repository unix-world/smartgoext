package httpimg_test

import (
	"github.com/unix-world/smartgoext/fpdf"
	"github.com/unix-world/smartgoext/fpdf/contrib/httpimg"
	"github.com/unix-world/smartgoext/fpdf/internal/example"
)

func ExampleRegister() {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetFillColor(200, 200, 220)
	pdf.AddPage()

	url := "https://github.com/unix-world/smartgoext/fpdf/raw/main/image/logo_gofpdf.jpg"
	httpimg.Register(pdf, url, "")
	pdf.Image(url, 15, 15, 267, 0, false, "", 0, "")
	fileStr := example.Filename("contrib_httpimg_Register")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated ../../pdf/contrib_httpimg_Register.pdf
}
