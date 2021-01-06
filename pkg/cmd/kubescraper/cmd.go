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

import "github.com/spf13/cobra"

// NewCommand returns the cobra command
func NewCommand( /*handlerFunction*/ ) *cobra.Command {

	// -- The command
	cmd := &cobra.Command{
		Use:     `scrape <path> [--telegram-token=<token> --backend-address <address> --backend-port <port> --admin-chat-id <chat-id> --gcp-service-account <service-account-path> --gcp-project-id <project-id> --debug]`,
		Example: `scrape <path> --telegram-token agfb09w7x --backend-address another-example.org --backend-port 8989 --admin-chat-id sdsd8fesbp --debug`,
		Short:   "scrape websites defined in the path file",
		Long: `The path file must be a valid path containing the websites and the polling options as
defined in scrape a webiste and notify users of a certain result.`,
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
