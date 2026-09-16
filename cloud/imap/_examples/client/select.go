package main

import (
	"testing"

	"github.com/unix-world/smartgoext/cloud/imap"
)

func TestSelect(t *testing.T) {
	client, server := newClientServerPair(t, imap.ConnStateAuthenticated)
	defer client.Close()
	defer server.Close()

	data, err := client.Select("INBOX", nil).Wait()
	if err != nil {
		t.Fatalf("Select() = %v", err)
	} else if data.NumMessages != 1 {
		t.Errorf("SelectData.NumMessages = %v, want %v", data.NumMessages, 1)
	}
}
