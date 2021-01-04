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

	poller "github.com/SunSince90/website-poller"
)

var (
	pages = map[string]*poller.Page{}
)

// SetPages sets the pages from the poller
func SetPages(polledPages []poller.Page) error {
	lock.Lock()
	defer lock.Unlock()

	if len(pages) > 0 {
		return fmt.Errorf("pages already set")
	}

	for i, page := range polledPages {
		if page.ID == nil {
			log.Warn().Str("url", page.URL).Msg("no id found, skipping...")
			continue
		}
		id := *page.ID
		pages[id] = &polledPages[i]
	}

	return nil
}
