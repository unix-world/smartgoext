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

type subscriptDelimiterProcessor struct {
}

func (p *subscriptDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '~'
}

func (p *subscriptDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *subscriptDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return ast.NewSubscript()
}

var defaultSubscriptDelimiterProcessor = &subscriptDelimiterProcessor{}

type subscriptParser struct {
}

var defaultSubscriptParser = &subscriptParser{}

// NewSubscriptParser return a new InlineParser that parses
// subscript expressions.
func NewSubscriptParser() parser.InlineParser {
	return defaultSubscriptParser
}

func (s *subscriptParser) Trigger() []byte {
	return []byte{'~'}
}

func (s *subscriptParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultSubscriptDelimiterProcessor)
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

func (s *subscriptParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// SubscriptHTMLRenderer is a renderer.NodeRenderer implementation that
// renders Subscript nodes.
type SubscriptHTMLRenderer struct {
	html.Config
}

// NewSubscriptHTMLRenderer returns a new SubscriptHTMLRenderer.
func NewSubscriptHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &SubscriptHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *SubscriptHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindSubscript, r.renderSubscript)
}

// SubscriptAttributeFilter defines attribute names which dd elements can have.
var SubscriptAttributeFilter = html.GlobalAttributeFilter

func (r *SubscriptHTMLRenderer) renderSubscript(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<sub")
			html.RenderAttributes(w, n, SubscriptAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<sub>")
		}
	} else {
		_, _ = w.WriteString("</sub>")
	}
	return gast.WalkContinue, nil
}

type subscript struct {
}

// Subscript is an extension that allows you to use a subscript expression like 'x~0~'.
var Subscript = &subscript{}

func (e *subscript) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewSubscriptParser(), 600),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewSubscriptHTMLRenderer(), 600),
	))
}
