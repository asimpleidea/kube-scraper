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
	"fmt"
	"os"

	"github.com/SunSince90/kube-scraper/pkg/cmd/kubescraper"
)

// Example of additional flags
// var (
// 	flag int
// )

func main() {
	cmd := kubescraper.NewCommand(scrape)

	// Add any additional flags here:
	// cmd.Flags().IntVar(&flag, "another-flag", 5, "the maximum times an error is tollerable")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
