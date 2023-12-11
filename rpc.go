package main

import "math/big"

type Args struct {
	Command string
	Address string
	Offset  int64
}

type Reply struct {
	Found      bool
	Reply      string
	Forward    string
	Successors []NodeAddress
}

type s struct {
	ID *big.Int
}

type r struct {
	Address NodeAddress
}

type FArgs struct {
	Address  string
	Filename string
	Content  string
}

type FReply struct {
}

type KArgs struct {
	Filename  string
	PublicKey int64
	Prime     int64
	Generator int64
	Secret    int64
}

type KReply struct {
	Content   string
	PublicKey int64
	EncKey    string
}
