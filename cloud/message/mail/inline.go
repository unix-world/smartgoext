package mail

import (
	"github.com/unix-world/smartgoext/cloud/message"
)

// A InlineHeader represents a message text header.
type InlineHeader struct {
	message.Header
}

var _ PartHeader = (*InlineHeader)(nil)

func (*InlineHeader) partHeader() {}
