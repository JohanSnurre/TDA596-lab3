package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"math/big"
)

func main() {
	address := flag.String("a", "", "The IP address that the Chord client will bind to, as well as advertise to other nodes")
	port := flag.Int("p", 0, "The port that the Chord client will bind to and listen on")

	joinAddress := flag.String("ja", "", "The IP address of the machine running a Chord node")
	joinPort := flag.Int("jp", 0, "The port that an existing Chord node is bound to and listening on")

	ts := flag.Int("ts", 0, "The time in milliseconds between invocations of 'stabilize'")
	tff := flag.Int("tff", 0, "The time in milliseconds between invocations of 'fix fingers'")
	tcp := flag.Int("tcp", 0, "The time in milliseconds between invocations of 'check predecessor'")

	r := flag.Int("r", 0, "The number of successors maintained by the Chord client")

	id := flag.String("i", "", "The identifier (ID) assigned to the Chord client which will override the ID computed by the SHA1 sum of the client’s IP address and port number")

	flag.Parse()

	newNode := Node{
		Address:    NodeAddress(*address + ":" + fmt.Sprint(*port)),
		Ts:         *ts,
		Tff:        *tff,
		Tcp:        *tcp,
		Successors: make([]NodeAddress, *r),
	}

	if isFlagPassed("i") {
		newNode.Id, _ = new(big.Int).SetString(*id, 10)
	} else {
		newNode.Id = hashString(string(newNode.Address))
		fmt.Println("my id:", newNode.Id)
	}

	if isFlagPassed("ja") && isFlagPassed("jp") {
		go newNode.join(NodeAddress(*joinAddress + ":" + fmt.Sprint(*joinPort)))
	} else {
		go newNode.create()
	}
	newNode.run()

}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func hashString(elt string) *big.Int {

	hasher := sha1.New()
	hasher.Write([]byte(elt))

	return new(big.Int).SetBytes(hasher.Sum(nil))

}

