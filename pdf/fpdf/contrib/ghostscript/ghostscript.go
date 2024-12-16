package main

// v.20241215.1258
// (c) unix-world.org
// license: BSD

// This command demonstrates the use of ghotscript to reduce the size
// of generated PDFs. This is based on a comment made by farkerhaiku:
// https://github.com/phpdave11/gofpdf/issues/57#issuecomment-185843315

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/unix-world/smartgoext/pdf/fpdf"
)

func report(fileStr string, err error) {
	if err == nil {
		var info os.FileInfo
		info, err = os.Stat(fileStr)
		if err == nil {
			fmt.Printf("%s: OK, size %d\n", fileStr, info.Size())
		} else {
			fmt.Printf("%s: bad stat\n", fileStr)
		}
	} else {
		fmt.Printf("%s: %s\n", fileStr, err)
	}
}

func newPdf() (pdf *fpdf.Fpdf) {
	pdf = fpdf.New("P", "mm", "A4", "../../font")
	pdf.SetCompression(false)
	pdf.AddFont("Calligrapher", "", "calligra.json")
	pdf.AddPage()
	pdf.SetFont("Calligrapher", "", 35)
	pdf.Cell(0, 10, "Enjoy new fonts with FPDF!")
	return
}

func full(name string) {
	report(name, newPdf().OutputFileAndClose(name))
}

func min(name string) {
	// gs -sDEVICE=pdfwrite -dCompatibilityLevel=1.5 -dPDFSETTINGS=/printer -dPDFA-3 -dNOPAUSE -dQUIET -dBATCH -sOutputFile=out-gs.pdf out.pdf
	cmd := exec.Command("gs", "-sDEVICE=pdfwrite", "-dCompatibilityLevel=1.4", "-dPDFSETTINGS=/screen", "-dNOPAUSE", "-dQUIET", "-dBATCH", "-sOutputFile="+name, "-")
	inPipe, err := cmd.StdinPipe()
	if err == nil {
		errChan := make(chan error, 1)
		go func() {
			errChan <- cmd.Start()
		}()
		err = newPdf().Output(inPipe)
		if err == nil {
			report(name, <-errChan)
		} else {
			report(name, err)
		}
	} else {
		report(name, err)
	}
}

func main() {
	full("full.pdf")
	min("min.pdf")
}
