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
	tLock.Lock()
	defer tLock.Unlock()

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
	tLock.Lock()
	defer tLock.Unlock()

	if adminChatID > 0 {
		adminChatID = id
	}

	return fmt.Errorf("admin chat alredy set")
}
