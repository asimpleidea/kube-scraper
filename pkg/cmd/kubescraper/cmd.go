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
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	websitepoller "github.com/SunSince90/website-poller"
	redis "github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	defaultRedisChannel string = "poll-result"
)

var (
	log         zerolog.Logger
	pages       []websitepoller.Page
	respHandler ResponseHandler
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger().Level(zerolog.InfoLevel)
}

// NewCommand returns the cobra command
func NewCommand(h ResponseHandler, opts ...Option) *cobra.Command {
	if h == nil {
		log.Fatal().Msg("no response handler provided")
		return nil
	}
	respHandler = h

	conf := &options{}

	// -- The command
	cmd := &cobra.Command{
		Use:     `scrape <path> [--debug]`,
		Example: `scrape ./path/to/the/yaml/file --debug`,
		Short:   "scrape websites defined in the path file",
		Long: `The path file must be a valid path containing the websites and the polling options as
defined in scrape a webiste and notify users of a certain result.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if conf.debug {
				log.Level(zerolog.DebugLevel)
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

			// TODO: store redis channe somewhere
			_ = conf.redisChannel
		},
		Run: func(_ *cobra.Command, _ []string) {
			run(conf)
		},
	}

	// -- Flags
	cmd.PersistentFlags().BoolVar(&conf.debug, "debug", false, "whether to enable debug log lines")
	cmd.Flags().StringVar(&conf.redisAddr, "redis-address", "", "the address where redis is running")
	cmd.Flags().StringVar(&conf.redisChannel, "redis-channel", defaultRedisChannel, "the channel where to publish events on redis")

	// Required
	cmd.MarkFlagRequired("redis-address")

	return cmd
}

func run(opts *options) {
	// -- Init
	log.Info().Msg("starting...")
	ctx, canc := context.WithCancel(context.Background())
	exitChan := make(chan struct{})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	go func() {
		defer close(exitChan)

		// -- Get redis client
		rdb, err := func() (*redis.Client, error) {
			_rdb := redis.NewClient(&redis.Options{Addr: opts.redisAddr})
			rdCtx, rdCanc := context.WithTimeout(ctx, 15*time.Second)
			defer rdCanc()

			if _, err := _rdb.Ping(rdCtx).Result(); err != nil {
				log.Err(err).Msg("could not receive ping from redis, exiting...")
				return nil, err
			}

			return _rdb, nil
		}()
		if err != nil {
			signalChan <- os.Interrupt
			return
		}

		// TODO: store redis client somewhere
		_ = rdb
		log.Info().Msg("connected to redis")
		defer rdb.Close()

		// -- Create the pollers
		pollers := map[string]websitepoller.Poller{}
		for i := range pages {
			p, err := websitepoller.New(&pages[i])
			if err != nil {
				log.Err(err).Int("index", i).Msg("error while creating a poller for page with this index, exiting...")
				signalChan <- os.Interrupt
				return
			}

			p.SetHandlerFunc(func(id string, resp *http.Response, err error) {
				respHandler(id, resp, err)
			})
			pollers[p.GetID()] = p
		}
		log.Info().Msg("all pollers set up")

		// -- Start the pollers
		var wg sync.WaitGroup
		wg.Add(len(pollers))
		defer wg.Wait()

		for _, poller := range pollers {
			go func(p websitepoller.Poller) {
				p.Start(ctx, true)
				wg.Done()
			}(poller)
		}

		log.Info().Msg("all pollers started, waiting for shutdown command")
	}()

	<-signalChan
	log.Info().Msg("exit requested")

	// -- Close all connections and shut down
	canc()
	<-exitChan

	log.Info().Msg("goodbye!")
}
