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

// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/PuerkitoBio/goquery"
	"github.com/SunSince90/kube-scraper-backend/pkg/pb"
	"github.com/SunSince90/kube-scraper/pkg/cmd/kubescraper"
	"github.com/rs/zerolog"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

const (
	targetPrice float64 = 300
	websiteName string  = "MamazonE"
)

var (
	log      zerolog.Logger
	products map[string]string
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	products = map[string]string{"poller-id-1": "iPhone 12", "poller-id-2": "iPhone 12 Pro"}
}

func scrape(opts *kubescraper.HandlerOptions, id string, resp *http.Response, err error) {
	if err != nil {
		log.Err(err).Msg("error on response")
		return
	}

	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != 200 {
		if opts.AdminChatID > 0 && opts.TelegramBotClient != nil {
			message := fmt.Sprintf("poller with id %s returned status %s", id, resp.Status)
			conf := tgbotapi.NewMessage(opts.AdminChatID, message)
			if _, err := opts.TelegramBotClient.Send(conf); err != nil {
				log.Err(err).Msg("error while notify admin about the error")
			}
		}
		log.Info().Str("status", resp.Status).Str("id", id).Msg("got response")
		return
	}

	// Parse the document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		if opts.AdminChatID > 0 && opts.TelegramBotClient != nil {
			message := fmt.Sprintf("could not scrape website: poller %s returned error %s", id, err.Error())
			conf := tgbotapi.NewMessage(opts.AdminChatID, message)
			if _, err := opts.TelegramBotClient.Send(conf); err != nil {
				log.Err(err).Msg("error while notify admin about the error")
			}
		}
		log.Err(err).Str("id", id).Msg("could not scrape website")
		return
	}

	// Get the price of the product
	price := doc.Find("span#price").First()
	f, _ := strconv.ParseFloat(price.Text(), 64)
	if f >= targetPrice {
		if opts.PubSubTopic != nil {
			// Send a pub sub message
			m := map[string]string{
				"price":   price.Text(),
				"message": "price is higher",
			}
			byteMessage, _ := json.Marshal(m)
			pubMsg := &pubsub.Message{
				Data: byteMessage,
			}
			ctx, canc := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
			defer canc()
			pubResult := opts.PubSubTopic.Publish(ctx, pubMsg)
			_ = pubResult
		}
		return
	}

	// Send a notification to all users subscribed to the bot
	if opts.TelegramBotClient == nil {
		log.Error().Str("id", id).Msg("bot was not set: no message will be sent")
		return
	}
	if opts.BackendClient == nil {
		log.Error().Str("id", id).Msg("no backend endpoint is set, no message will be sent")
		return
	}

	ctx, canc := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer canc()

	chats, err := opts.BackendClient.GetChatsList(ctx, &pb.ChatRequest{})
	if err != nil {
		log.Err(err).Str("id", id).Msg("error while getting chats list")
		return
	}

	for _, chat := range chats.Chats {
		message := fmt.Sprintf("%s is now priced %f at %s! Go buy at %s", products[id], f, websiteName, resp.Request.URL.String())
		conf := tgbotapi.NewMessage(opts.AdminChatID, message)
		if _, err := opts.TelegramBotClient.Send(conf); err != nil {
			log.Err(err).Int64("chat-id", chat.Id).Msg("error while sending telegram message to this chat, skipping...")
			continue
		}
	}
}
