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

	websitepoller "github.com/SunSince90/website-poller"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

	// -- The command
	cmd := &cobra.Command{
		Use:     `scrape <path> [--debug]`,
		Example: `scrape ./path/to/the/yaml/file --debug`,
		Short:   "scrape websites defined in the path file",
		Long: `The path file must be a valid path containing the websites and the polling options as
defined in scrape a webiste and notify users of a certain result.`,
		PreRun: preRun,
		Run:    run,
	}

	// -- Flags
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "whether to enable debug log lines")

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
			log.Fatal().Err(err).Int("index", i).Msg("error while creating a poller for page with this index, exiting...")
		}

		p.SetHandlerFunc(func(id string, resp *http.Response, err error) {
			respHandler(id, resp, err)
		})
		pollers[p.GetID()] = p
	}
	log.Info().Msg("all pollers set up")

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
	wg.Wait()
	log.Info().Msg("goodbye!")
}
