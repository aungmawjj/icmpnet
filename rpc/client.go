package rpc

import (
	"io"
	"io/ioutil"
	"net/rpc"
	"path"
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
func (c *Client) FileUpload(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	fp := &FilePayload{
		Name: filepath.Base(filename),
		Data: data,
	}
	return c.rpcClient.Call(MethodFile+".Upload", fp, &struct{}{})
}

// FileDownload downloads a file
func (c *Client) FileDownload(filename string, downloadDir string) error {
	req := &FilePayload{
		Name: filename,
	}
	resp := new(FilePayload)
	err := c.rpcClient.Call(MethodFile+".Download", req, resp)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(downloadDir, filename), resp.Data, 0644)
}

// InfoVersion requests server version info
func (c *Client) InfoVersion() (string, error) {
	var vInfo string
	return vInfo, c.rpcClient.Call(MethodInfo+".Version", &struct{}{}, &vInfo)
}
