package rpc

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"time"
)

// FilePayload type
type FilePayload struct {
	Name string
	Data []byte
}

// FileAPI type
type FileAPI struct {
	dirPath string
}

// NewFileAPI creates a new FileAPI
func NewFileAPI(dirPath string) *FileAPI {
	return &FileAPI{
		dirPath: dirPath,
	}
}

// Upload a file
func (api *FileAPI) Upload(fp *FilePayload, reply *struct{}) error {
	log.Printf("File upload: %s", fp.Name)
	filename := fmt.Sprintf("%d-%s", time.Now().Unix(), fp.Name)
	err := ioutil.WriteFile(path.Join(api.dirPath, filename), fp.Data, 0644)
	return err
}
