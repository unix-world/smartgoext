package extension

import (
	"github.com/unix-world/smartgoext/markup/goldmark"
)

type gfm struct {
}

// GFM is an extension that provides Github Flavored markdown functionalities.
var GFM = &gfm{}

func (e *gfm) Extend(m goldmark.Markdown) {
	Linkify.Extend(m)
	Table.Extend(m)
	Strikethrough.Extend(m)
	TaskList.Extend(m)
}
