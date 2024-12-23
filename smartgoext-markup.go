
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2021-present unix-world.org
// r.20241223.2358 :: STABLE
// [ MARKUP ]

// REQUIRE: go 1.22 or later
package smartgoext


import (
	"log"

	"bytes"

	smart "github.com/unix-world/smartgo"

	"github.com/unix-world/smartgoext/markup/goldmark"
	"github.com/unix-world/smartgoext/markup/goldmark/extension"
	"github.com/unix-world/smartgoext/markup/goldmark/parser"
	"github.com/unix-world/smartgoext/markup/goldmark/renderer/html"
	extsupersub "github.com/unix-world/smartgoext/markup/goldmark/extensions/super-sub-script"
)

//-----


func MarkdownGfToHTMLRender(mkdwDoc string) (string, error) {
	//--
	defer smart.PanicHandler() // just in case
	//--
	if(mkdwDoc == "") {
		return "<!-- Markdown:empty -->", nil
	} //end if
	if(uint64(len(mkdwDoc)) > smart.MAX_DOC_SIZE_MARKDOWN) {
		return "<!-- Markdown:oversized -->", nil
	} //end if
	//-- goldmark
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Footnote, extsupersub.Subscript, extsupersub.Superscript),
		goldmark.WithParserOptions(
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithEscapedHtml(),
		),
	)
	if(md == nil) {
		log.Println("[WARNING] Markdown Init Failed")
		return "<!-- Markdown:init.failed -->", smart.NewError("Markdown Init Failed")
	} //end if
	var buf bytes.Buffer
	errRender := md.Convert([]byte(mkdwDoc), &buf)
	if(errRender != nil) {
		log.Println("[WARNING] Markdown Render Failed:", errRender)
		return "<!-- Markdown:html.err-render -->", errRender
	} //end if
	var htmlCode string = `<div class="markdown" data-type="gfm">` + "\n" + buf.String() + "\n" + `</div>`
	buf.Reset() // free mem
	//-- #
	if(DEBUG == true) {
		log.Println("[DATA] Markdown HTML: ========", htmlCode)
	} //end if
	//-- sanitizer
	htmlCode, errHtmlSanitizer := smart.HTMLCodeFixSanitize(htmlCode)
	if(errHtmlSanitizer != nil) {
		log.Println("[WARNING] Markdown HTML Sanitized:", errHtmlSanitizer)
		return "<!-- Markdown:html.err-fix.sn -->", errHtmlSanitizer
	} //end if
	if(DEBUG == true) {
		log.Println("[DATA] Markdown HTML Sanitized: ========", htmlCode)
	} //end if
	//-- validator
	htmlCode, errFixHtml := smart.HTMLCodeFixValidate(htmlCode)
	if(errFixHtml != nil) {
		log.Println("[WARNING] Markdown HTML ValidateFixed:", errHtmlSanitizer)
		return "<!-- Markdown:html.err-fix.vd -->", errHtmlSanitizer
	} //end if
	if(DEBUG == true) {
		log.Println("[DATA] Markdown HTML Fixed (Sanitized + Validated): ========", htmlCode)
	} //end if
	//--
	return htmlCode + "<!-- Markdown:html.safe -->", errHtmlSanitizer
	//--
} //END FUNCTION


func SafePathMarkdownGfFileToHTMLRender(mdFilePath string, allowAbsolutePath bool) (string, error) {
	//--
	defer smart.PanicHandler()
	//--
	if(smart.StrTrimWhitespaces(mdFilePath) == "") {
		return "<!-- # Markdown.Err:1 -->", smart.NewError("Markdown File # File Path is Empty")
	} //end if
	//--
	mdFilePath = smart.SafePathFixClean(mdFilePath)
	//--
	if(smart.PathIsEmptyOrRoot(mdFilePath) == true) {
		return "<!-- # Markdown.Err:2 -->", smart.NewError("Markdown File # File Path is Empty/Root")
	} //end if
	//--
	if(!smart.StrEndsWith(mdFilePath, ".gf.md")) {
		return "<!-- # Markdown.Err:3 -->", smart.NewError("Markdown File # Invalid File Extension, accepted: .gf.md # `" + mdFilePath + "`")
	} //end if
	//--
	fileSize, errSize := smart.SafePathFileGetSize(mdFilePath, allowAbsolutePath)
	if(errSize != nil) {
		return "", errSize
	} //end if
	if(uint64(fileSize) > smart.MAX_DOC_SIZE_MARKDOWN) {
		return "<!-- # Markdown.Err:4 -->", smart.NewError("Markdown File # OverSized # `" + mdFilePath + "`")
	} //end if
	//--
	mdData, errRd := smart.SafePathFileRead(mdFilePath, allowAbsolutePath)
	if(errRd != nil) {
		return "<!-- # Markdown.Err:5 -->", smart.NewError("Markdown File # Read Failed `" + mdFilePath + "`: " + errRd.Error())
	} //end if
	if(smart.StrTrimWhitespaces(mdData) == "") {
		return "<!-- # Markdown.Err:6 -->", smart.NewError("Markdown File # Content is Empty `" + mdFilePath + "`")
	} //end if
	//--
	html, err := MarkdownGfToHTMLRender(mdData)
	if(err != nil) {
		return "<!-- # Markdown.Err:7 -->", smart.NewError("Markdown File # Parse ERR: " + err.Error() + " # `" + mdFilePath + "`")
	} //end if
	//--
	return html, nil
	//--
} //END FUNCTION


// #END
