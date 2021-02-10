package rpc

import (
	"log"
)

// InfoAPI type
type InfoAPI struct {
	versionInfo string
}

// NewInfoAPI creates a new InfoAPI
func NewInfoAPI(versionInfo string) *InfoAPI {
	return &InfoAPI{
		versionInfo: versionInfo,
	}
}

// Version handler
func (api *InfoAPI) Version(req *struct{}, reply *string) error {
	log.Println("Version request")
	*reply = api.versionInfo
	return nil
}
