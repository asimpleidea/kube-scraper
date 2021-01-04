// Copyright © 2020 Elis Lulja
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
