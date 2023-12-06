package main

import (
	"math/big"
)

type Args struct {
	Id *big.Int
}

type ReplyFindSuc struct {
	IsFound bool
	Reply   string
}
