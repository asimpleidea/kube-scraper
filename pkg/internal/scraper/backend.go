package scraper

import (
	"fmt"
)

var (
	backendEndpoint string
)

// SetBackendEndpoint sets the backend endpoint
func SetBackendEndpoint(address string, port int) error {
	lock.Lock()
	defer lock.Unlock()

	if len(backendEndpoint) > 0 {
		backendEndpoint = fmt.Sprintf("%s:%d", address, port)
	}

	return fmt.Errorf("backend endpoint already set")
}
