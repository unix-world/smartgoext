package imapserver

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"

	"github.com/unix-world/smartgoext/cloud/imap"
	"github.com/unix-world/smartgoext/cloud/imap/internal/imapwire"
)

func (c *Conn) canStartTLS() bool {
	_, isTLS := c.conn.(*tls.Conn)
	return c.server.options.TLSConfig != nil && c.state == imap.ConnStateNotAuthenticated && !isTLS
}

func (c *Conn) handleStartTLS(tag string, dec *imapwire.Decoder) error {
	if !dec.ExpectCRLF() {
		return dec.Err()
	}

	if c.server.options.TLSConfig == nil {
		return &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: "STARTTLS not supported",
		}
	}
	if !c.canStartTLS() {
		return &imap.Error{
			Type: imap.StatusResponseTypeBad,
			Text: "STARTTLS not available",
		}
	}

	// Do not allow to write cleartext data past this point: keep c.encMutex
	// locked until the end
	enc := newResponseEncoder(c)
	defer enc.end()

	err := writeStatusResp(enc.Encoder, tag, &imap.StatusResponse{
		Type: imap.StatusResponseTypeOK,
		Text: "Begin TLS negotiation now",
	})
	if err != nil {
		return err
	}

	// Drain buffered data from our bufio.Reader
	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, c.br, int64(c.br.Buffered())); err != nil {
		panic(err) // unreachable
	}

	var cleartextConn net.Conn
	if buf.Len() > 0 {
		r := io.MultiReader(&buf, c.conn)
		cleartextConn = startTLSConn{c.conn, r}
	} else {
		cleartextConn = c.conn
	}

	tlsConn := tls.Server(cleartextConn, c.server.options.TLSConfig)

	c.mutex.Lock()
	c.conn = tlsConn
	c.mutex.Unlock()

	rw := c.server.options.wrapReadWriter(tlsConn)
	c.br.Reset(rw)
	c.bw.Reset(rw)

	return nil
}

type startTLSConn struct {
	net.Conn
	r io.Reader
}

func (conn startTLSConn) Read(b []byte) (int, error) {
	return conn.r.Read(b)
}
