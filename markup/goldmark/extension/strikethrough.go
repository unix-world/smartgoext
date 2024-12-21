package extension

// modified by unixman

import (
	"github.com/unix-world/smartgoext/markup/goldmark"
	gast "github.com/unix-world/smartgoext/markup/goldmark/ast"
	"github.com/unix-world/smartgoext/markup/goldmark/extension/ast"
	"github.com/unix-world/smartgoext/markup/goldmark/parser"
	"github.com/unix-world/smartgoext/markup/goldmark/renderer"
	"github.com/unix-world/smartgoext/markup/goldmark/renderer/html"
	"github.com/unix-world/smartgoext/markup/goldmark/text"
	"github.com/unix-world/smartgoext/markup/goldmark/util"
)

type strikethroughDelimiterProcessor struct {
}

func (p *strikethroughDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '~'
}

func (p *strikethroughDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *strikethroughDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return ast.NewStrikethrough()
}

var defaultStrikethroughDelimiterProcessor = &strikethroughDelimiterProcessor{}

type strikethroughParser struct {
}

var defaultStrikethroughParser = &strikethroughParser{}

// NewStrikethroughParser return a new InlineParser that parses
// strikethrough expressions.
func NewStrikethroughParser() parser.InlineParser {
	return defaultStrikethroughParser
}

func (s *strikethroughParser) Trigger() []byte {
	return []byte{'~'}
}

func (s *strikethroughParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
//	node := parser.ScanDelimiter(line, before, 1, defaultStrikethroughDelimiterProcessor)
	node := parser.ScanDelimiter(line, before, 2, defaultStrikethroughDelimiterProcessor) // fix by unixman: avoid colission with: ~subscript~ syntax
	if node == nil || node.OriginalLength > 2 || before == '~' {
		return nil
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *strikethroughParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// StrikethroughHTMLRenderer is a renderer.NodeRenderer implementation that
// renders Strikethrough nodes.
type StrikethroughHTMLRenderer struct {
	html.Config
}

// NewStrikethroughHTMLRenderer returns a new StrikethroughHTMLRenderer.
func NewStrikethroughHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &StrikethroughHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *StrikethroughHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindStrikethrough, r.renderStrikethrough)
}

// StrikethroughAttributeFilter defines attribute names which dd elements can have.
var StrikethroughAttributeFilter = html.GlobalAttributeFilter

func (r *StrikethroughHTMLRenderer) renderStrikethrough(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<del")
			html.RenderAttributes(w, n, StrikethroughAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<del>")
		}
	} else {
		_, _ = w.WriteString("</del>")
	}
	return gast.WalkContinue, nil
}

type strikethrough struct {
}

// Strikethrough is an extension that allow you to use strikethrough expression like '~~text~~' .
var Strikethrough = &strikethrough{}

func (e *strikethrough) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewStrikethroughParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewStrikethroughHTMLRenderer(), 500),
	))
}