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
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"cloud.google.com/go/pubsub"
	bpb "github.com/SunSince90/kube-scraper-backend/pkg/pb"
	websitepoller "github.com/SunSince90/website-poller"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"gopkg.in/yaml.v3"
)

var (
	log         zerolog.Logger
	pages       []websitepoller.Page
	opts        *HandlerOptions
	conn        *grpc.ClientConn
	pubsubcli   *pubsub.Client
	respHandler ResponseHandler
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	opts = &HandlerOptions{}
}

// NewCommand returns the cobra command
func NewCommand(h ResponseHandler) *cobra.Command {
	if h == nil {
		log.Fatal().Msg("no response handler provided")
		return nil
	}
	respHandler = h

	// -- The command
	cmd := &cobra.Command{
		Use:     `scrape <path> [--telegram-token=<token> --backend-address <address> --backend-port <port> --admin-chat-id <chat-id> --gcp-service-account <service-account-path> --gcp-project-id <project-id> --debug]`,
		Example: `scrape <path> --telegram-token agfb09w7x --backend-address another-example.org --backend-port 8989 --admin-chat-id sdsd8fesbp --debug`,
		Short:   "scrape websites defined in the path file",
		Long: `The path file must be a valid path containing the websites and the polling options as
defined in scrape a webiste and notify users of a certain result.`,
		PreRun: preRun,
		Run:    run,
	}

	// -- Flags
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "whether to enable debug log lines")
	cmd.Flags().StringVar(&telegramToken, "telegram-token", "", "the telegram token")
	cmd.Flags().StringVar(&backendAddress, "backend-address", "", "the backend address")
	cmd.Flags().IntVar(&backendPort, "backend-port", 0, "the backend port")
	cmd.Flags().Int64Var(&adminChatID, "admin-chat-id", 0, "the admin chat id")
	cmd.Flags().StringVar(&topicName, "pubsub-topic-name", "", "the name of the pubsub topic")
	cmd.Flags().StringVar(&gcpServAcc, "gcp-service-account", "", "the path to the gcp service account")
	cmd.Flags().StringVar(&projectID, "gcp-project-id", "", "the ID of the gcp/firestore project")

	return cmd
}

func preRun(cmd *cobra.Command, args []string) {
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// -- Get the path
	if len(args) == 0 {
		log.Fatal().Msg("no path set, exiting...")
		return // unnecessary but just for clarity
	}
	if len(args) > 1 {
		log.Warn().Int("args-len", len(args)).Msg("multiple paths are not supported yet: only the first one will be used")
	}
	yamlPath := args[0]
	var _pages []websitepoller.Page
	pagesByte, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse yaml file, exiting")
		return
	}
	if err := yaml.Unmarshal(pagesByte, &_pages); err != nil {
		log.Fatal().Err(err).Msg("could not unmarshal yaml file, exiting")
		return
	}
	pages = _pages

	// -- Is telegram set?
	if telegramToken != "" {
		_bot, err := tgbotapi.NewBotAPI(telegramToken)
		if err != nil {
			log.Fatal().Err(err).Str("telegram-token", telegramToken).Msg("could not start telegram bot, exiting...")
			return
		}

		log.Debug().Str("account", _bot.Self.UserName).Msg("authorized")
		_bot.Debug = debug
		opts.TelegramBotClient = _bot

		// -- Get the admin chat
		if adminChatID > 0 {
			opts.AdminChatID = adminChatID
			log.Debug().Int64("chat-id", adminChatID).Msg("parsed admin chat ID")
		}
	}

	// -- Get the backend
	if backendAddress != "" && backendPort > 0 {
		endpoint := fmt.Sprintf("%s:%d", backendAddress, backendPort)
		_conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
		if err != nil {
			log.Fatal().Err(err).Str("endpoint", endpoint).Msg("could not establish a gRPC connection")
		}
		conn = _conn
		log.Debug().Str("endpoint", endpoint).Msg("parsed backend flags")
		opts.BackendClient = bpb.NewBackendClient(conn)
	} else {
		log.Info().Msg("backend flags skipped because either address or port are not provided")
	}

	// -- Get pubsub client
	if topicName != "" && projectID != "" && gcpServAcc != "" {
		pscli, err := pubsub.NewClient(context.Background(), projectID, option.WithServiceAccountFile(gcpServAcc))
		if err != nil {
			if conn != nil {
				conn.Close()
			}
			log.Fatal().Err(err).Str("topic-name", topicName).Str("gcp-service-account", gcpServAcc).Str("project-id", projectID).Msg("could not establish a gRPC connection")
			return
		}
		pubsubcli = pscli
		opts.PubSubTopic = pubsubcli.Topic(topicName)
	} else {
		log.Info().Msg("pubsub flags skipped because either topic or service account are not provided")
	}
}

func run(cmd *cobra.Command, args []string) {
	// -- Init
	log.Info().Msg("starting...")
	ctx, canc := context.WithCancel(context.Background())
	pollers := map[string]websitepoller.Poller{}
	var wg sync.WaitGroup

	// -- Create the pollers
	for i := range pages {
		p, err := websitepoller.New(&pages[i])
		if err != nil {
			if conn != nil {
				conn.Close()
			}
			if pubsubcli != nil {
				pubsubcli.Close()
			}
			log.Fatal().Err(err).Int("index", i).Msg("error while creating a poller for page with this index, exiting...")
		}

		p.SetHandlerFunc(func(id string, resp *http.Response, err error) {
			respHandler(opts, id, resp, err)
		})
		pollers[p.GetID()] = p
	}
	log.Debug().Msg("all pollers set up")

	// -- Start the pollers
	wg.Add(len(pollers))
	for _, poller := range pollers {
		go func(p websitepoller.Poller) {
			p.Start(ctx, true)
			wg.Done()
		}(poller)
	}
	log.Info().Msg("all pollers started, waiting for shutdown command")

	// -- Graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	<-signalChan
	log.Info().Msg("exit requested")

	// -- Close all connections and shut down
	canc()
	if conn != nil {
		conn.Close()
	}
	if pubsubcli != nil {
		pubsubcli.Close()
	}
	log.Info().Msg("goodbye!")
}
