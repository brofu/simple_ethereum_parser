package main

import (
	"fmt"
	"os"

	"github.com/brofu/simple_ethereum_parser/packages/ethereum"
	"github.com/brofu/simple_ethereum_parser/packages/logging"
	"github.com/brofu/simple_ethereum_parser/packages/parser"
	"github.com/spf13/cobra"
)

var (
	entryPoint     = "https://cloudflare-eth.com/"
	testEntryPoint = "http://localhost:8080/rpc"
)

func main() {

	logger := logging.NewDefaultLogger(logging.LevelDebug)
	chainAccesser := ethereum.NewEthJsonRpcClient(entryPoint, logger)
	parser := parser.NewToolParser(logger, chainAccesser)

	var rootCmd = &cobra.Command{Use: "cmd-tool"}

	var blockNumCmd = &cobra.Command{
		Use:   "get-block-number",
		Short: "Get current block number",
		Run: func(cmd *cobra.Command, args []string) {
			bn := parser.GetCurrentBlock()
			fmt.Printf("%d\n", bn)
		},
	}

	var trxCmd = &cobra.Command{
		Use:   "get-transactions [address]",
		Short: "Get transactions of an address",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			address := args[0]
			trx := parser.GetTransactions(address)
			fmt.Printf("%+v\n", trx)
		},
	}

	rootCmd.AddCommand(blockNumCmd)
	rootCmd.AddCommand(trxCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
