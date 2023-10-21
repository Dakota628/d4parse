package ribbit2

import (
	"github.com/Dakota628/d4parse/pkg/bnet/bpsv"
	"io"
	"net"
)

const (
	DefaultPort = "1119"
)

// Request ...
type Request struct {
	Command []byte
}

func (r Request) Write(c net.Conn) error {
	if _, err := c.Write(r.Command); err != nil {
		return err
	}

	if _, err := c.Write([]byte{'\r', '\n'}); err != nil {
		return err
	}

	return nil
}

// Response ...
type Response struct {
	Data []byte
}

func (r *Response) Read(c net.Conn) (err error) {
	r.Data, err = io.ReadAll(c)
	return
}

func (r *Response) BPSV() (bpsv.Document, error) {
	return bpsv.ParseDocument(r.Data)
}

// Client ...
type Client struct {
	addr string
}

func NewClient(host string, port ...string) (*Client, error) {
	// Set port to default if not provided
	if len(port) == 0 {
		port = []string{DefaultPort}
	}

	// Create Ribbit addr
	addr := net.JoinHostPort(host, port[0])

	return &Client{
		addr: addr,
	}, nil
}

func (c *Client) getConn() (net.Conn, error) {
	return net.Dial("tcp", c.addr)
}

func (c *Client) Do(req Request) (resp Response, err error) {
	conn, err := c.getConn()
	if err != nil {
		return Response{}, err
	}
	defer conn.Close()

	err = req.Write(conn)
	if err != nil {
		return
	}

	err = resp.Read(conn)
	return
}
