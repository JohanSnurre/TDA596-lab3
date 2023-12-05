package main

type Args struct {
	Command string
	Address string
}

type Reply struct {
	Found   bool
	Reply   string
	Forward string
}
