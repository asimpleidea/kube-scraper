// Copyright © 2021 Elis Lulja
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

package main

import (
	"context"
	"net/http"

	"github.com/SunSince90/kube-scraper/pkg/cmd/kubescraper"
)

func scrape(id string, resp *http.Response, err error) {
	// Do your scraping here

	// Example:
	rdb := kubescraper.GetRedisClient()
	rdb.Publish(context.Background(), kubescraper.GetRedisPubChannel(), struct {
		message string
	}{message: "Hello, world!"})
}
