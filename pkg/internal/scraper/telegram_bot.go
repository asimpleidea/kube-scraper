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
	"sync"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var (
	bot         *tgbotapi.BotAPI
	adminChatID int64
	tLock       sync.Locker
)

// SetTelegramBot sets up the telegram bot
func SetTelegramBot(token string) (err error) {
	lock.Lock()
	defer lock.Unlock()

	if len(token) == 0 {
		return fmt.Errorf("no token provided")
	}
	if bot != nil {
		return fmt.Errorf("bot already started")
	}

	_bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	log.Debug().Str("account", bot.Self.UserName).Msg("authorized")
	bot = _bot

	return
}

// SetAdminChatID sets the chat id with the admin. This is used
func SetAdminChatID(id int64) error {
	lock.Lock()
	defer lock.Unlock()

	if adminChatID > 0 {
		adminChatID = id
	}

	return fmt.Errorf("admin chat alredy set")
}
