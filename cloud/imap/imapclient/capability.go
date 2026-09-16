package imapclient

import (
	"fmt"

	"github.com/unix-world/smartgoext/cloud/imap"
	"github.com/unix-world/smartgoext/cloud/imap/internal"
	"github.com/unix-world/smartgoext/cloud/imap/internal/imapwire"
)

// Capability sends a CAPABILITY command.
func (c *Client) Capability() *CapabilityCommand {
	cmd := &CapabilityCommand{}
	c.beginCommand("CAPABILITY", cmd).end()
	return cmd
}

func (c *Client) handleCapability() error {
	caps, err := readCapabilities(c.dec)
	if err != nil {
		return err
	}
	c.setCaps(caps)
	if cmd := findPendingCmdByType[*CapabilityCommand](c); cmd != nil {
		cmd.caps = caps
	}
	return nil
}

// CapabilityCommand is a CAPABILITY command.
type CapabilityCommand struct {
	commandBase
	caps imap.CapSet
}

func (cmd *CapabilityCommand) Wait() (imap.CapSet, error) {
	err := cmd.wait()
	return cmd.caps, err
}

func readCapabilities(dec *imapwire.Decoder) (imap.CapSet, error) {
	caps := make(imap.CapSet)
	for dec.SP() {
		// Some IMAP servers send multiple SP between caps:
		// github.com / emersion / go-imap / pull / 652
		for dec.SP() {
		}

		cap, err := internal.ExpectCap(dec)
		if err != nil {
			return caps, fmt.Errorf("in capability-data: %w", err)
		}
		caps[cap] = struct{}{}
	}
	return caps, nil
}
