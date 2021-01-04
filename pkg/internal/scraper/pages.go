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
