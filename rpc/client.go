package rpc

import (
	"io"
	"io/ioutil"
	"net/rpc"
	"path/filepath"
)

// Client type
type Client struct {
	rpcClient *rpc.Client
}

// NewClient creates a new Client
func NewClient(conn io.ReadWriteCloser) *Client {
	return &Client{
		rpcClient: rpc.NewClient(conn),
	}
}

// FileUpload uploads a file
func (c *Client) FileUpload(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	fp := &FilePayload{
		Name: filepath.Base(filePath),
		Data: data,
	}
	return c.rpcClient.Call(MethodFile+".Upload", fp, &struct{}{})
}
