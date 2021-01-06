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

package kubescraper

import (
	"net/http"

	"cloud.google.com/go/pubsub"
	bpb "github.com/SunSince90/kube-scraper-backend/pkg/pb"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// HandlerOptions contains data, such as variables and clients, that
// are passed to a handler so that the function can use.
type HandlerOptions struct {
	// TelegramBotClient is the bot client and can send messages
	TelegramBotClient *tgbotapi.BotAPI
	// AdminChatID is the ID of the chat with the admin
	AdminChatID int64
	// BackendClient is the client that comunicates with the backend pod
	BackendClient bpb.BackendClient
	// PubSubTopic where to publish messages
	PubSubTopic *pubsub.Topic
}

// ResponseHandler is a function is executed after each http call completed.
type ResponseHandler func(*HandlerOptions, string, *http.Response, error)
