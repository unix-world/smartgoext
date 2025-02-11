package extension

// modified by unixman

import (
	ast "github.com/unix-world/smartgoext/markup/goldmark/extensions/super-sub-script/ast" // unixman
	"github.com/unix-world/smartgoext/markup/goldmark"
	gast "github.com/unix-world/smartgoext/markup/goldmark/ast"
	"github.com/unix-world/smartgoext/markup/goldmark/parser"
	"github.com/unix-world/smartgoext/markup/goldmark/renderer"
	"github.com/unix-world/smartgoext/markup/goldmark/renderer/html"
	"github.com/unix-world/smartgoext/markup/goldmark/text"
	"github.com/unix-world/smartgoext/markup/goldmark/util"
)

type superscriptDelimiterProcessor struct {
}

func (p *superscriptDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '^'
}

func (p *superscriptDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *superscriptDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return ast.NewSuperscript()
}

var defaultSuperscriptDelimiterProcessor = &superscriptDelimiterProcessor{}

type superscriptParser struct {
}

var defaultSuperscriptParser = &superscriptParser{}

// NewSuperscriptParser return a new InlineParser that parses
// superscript expressions.
func NewSuperscriptParser() parser.InlineParser {
	return defaultSuperscriptParser
}

func (s *superscriptParser) Trigger() []byte {
	return []byte{'^'}
}

func (s *superscriptParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultSuperscriptDelimiterProcessor)
	if node == nil {
		return nil
	}
	if node.CanOpen {
		for i := 1; i < len(line); i++ {
			c := line[i]
			if c == line[0] {
				break
			}
			if util.IsSpace(c) {
				return nil
			}
		}
	}
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *superscriptParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// SuperscriptHTMLRenderer is a renderer.NodeRenderer implementation that
// renders Superscript nodes.
type SuperscriptHTMLRenderer struct {
	html.Config
}

// NewSuperscriptHTMLRenderer returns a new SuperscriptHTMLRenderer.
func NewSuperscriptHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &SuperscriptHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *SuperscriptHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindSuperscript, r.renderSuperscript)
}

// SuperscriptAttributeFilter defines attribute names which dd elements can have.
var SuperscriptAttributeFilter = html.GlobalAttributeFilter

func (r *SuperscriptHTMLRenderer) renderSuperscript(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<sup")
			html.RenderAttributes(w, n, SuperscriptAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<sup>")
		}
	} else {
		_, _ = w.WriteString("</sup>")
	}
	return gast.WalkContinue, nil
}

type superscript struct {
}

// Superscript is an extension that allows you to use a superscript expression like 'x^2^'.
var Superscript = &superscript{}

func (e *superscript) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewSuperscriptParser(), 600),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewSuperscriptHTMLRenderer(), 600),
	))
}
