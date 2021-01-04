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
	"net/http"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// Scrape the website
func Scrape(id string, resp *http.Response, err error) {
	if err != nil {
		log.Err(err).Msg("error on response")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {

		if len(adminChatID) > 0 {
			message := fmt.Sprintf("poller with id %s returned status %s", id, resp.Status)
			tgbotapi.NewMessage(adminChatID, message)
			bot.Send(&tgbot)
		}

	}
}
