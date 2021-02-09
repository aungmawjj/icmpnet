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
	filename := fmt.Sprintf("%s-%d", fp.Name, time.Now().Unix())
	err := ioutil.WriteFile(path.Join(api.dirPath, filename), fp.Data, 0644)
	if err != nil {
		log.Printf("File upload error: %s : %s\n", fp.Name, err)
	}
	return err
}

// Download a file
func (api *FileAPI) Download(fp *FilePayload, reply *FilePayload) error {
	log.Printf("File download request: %s", fp.Name)
	reply.Name = fp.Name
	data, err := ioutil.ReadFile(path.Join(api.dirPath, fp.Name))
	reply.Data = data
	if err != nil {
		log.Printf("File download error: %s : %s\n", fp.Name, err)
	}
	return err
}
