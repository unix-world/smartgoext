package main

import (
	"testing"

	"github.com/unix-world/smartgo/net/sasl"

	"github.com/unix-world/smartgoext/cloud/imap"
)

func TestClient_Authenticate(t *testing.T) {
	client, server := newClientServerPair(t, imap.ConnStateNotAuthenticated)
	defer client.Close()
	defer server.Close()

	saslClient := sasl.NewPlainClient("", testUsername, testPassword)
	if err := client.Authenticate(saslClient); err != nil {
		t.Fatalf("Authenticate() = %v", err)
	}

	if state := client.State(); state != imap.ConnStateAuthenticated {
		t.Errorf("State() = %v, want %v", state, imap.ConnStateAuthenticated)
	}
}
