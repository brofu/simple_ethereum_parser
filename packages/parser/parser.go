package parser

import "github.com/brofu/simple_ethereum_parser/packages/ethereum"

type Parser interface {
	GetCurrentBlock() int
	Subscribe(string) bool
	GetTransactions(address string) []ethereum.Transaction
}
