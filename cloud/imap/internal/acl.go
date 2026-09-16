package internal

import (
	"github.com/unix-world/smartgoext/cloud/imap"
)

func FormatRights(rm imap.RightModification, rs imap.RightSet) string {
	s := ""
	if rm != imap.RightModificationReplace {
		s = string(rm)
	}
	return s + string(rs)
}
