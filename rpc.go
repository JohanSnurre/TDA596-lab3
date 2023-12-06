package main

type Args struct {
	Command string
	NodeID  int
}

type Reply struct {
	Reply     string
	Successor int
	FoundSucc bool
}
