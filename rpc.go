package main

import "math/big"

type Args struct {
	Command   string
	Address   string
	Offset    int64
	Filename  string
	PublicKey int64
	Prime     int64
	Generator int64
}

type Reply struct {
	Found      bool
	Reply      string
	Forward    string
	Successors []NodeAddress
	Content    string
	PublicKey  int64
	EncKey     string
}

type s struct {
	ID *big.Int
}

type r struct {
	Address NodeAddress
}
