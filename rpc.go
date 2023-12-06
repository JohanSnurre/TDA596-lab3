package main

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
