// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/coinbase/rosetta-ethereum/ethereum"

	"github.com/spf13/cobra"
)

var (
	utilsBootstrapCmd = &cobra.Command{
		Use:   "utils:generate-bootstrap",
		Short: "Generate a bootstrap balance configuration file",
		Long: `For rosetta-cli testing, it can be useful to generate
a bootstrap balances file for balances that were created
at genesis. This command creates such a file given the
path of an Ethereum genesis file.

When calling this command, you must provide 2 arguments:
[1] the location of the genesis file
[2] the location of where to write bootstrap balances file`,
		RunE: runUtilsBootstrapCmd,
		Args: cobra.ExactArgs(2), //nolint:gomnd
	}
)

func runUtilsBootstrapCmd(cmd *cobra.Command, args []string) error {
	return ethereum.GenerateBootstrapFile(args[0], args[1])
}
