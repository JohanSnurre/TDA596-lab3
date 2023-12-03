package main

import (
	"crypto/sha1"
	"math/big"
)

type Key string
type NodeAddress string

type Node struct {
	Address     NodeAddress
	FingerTable []NodeAddress
	Predecessor NodeAddress
	Successors  []NodeAddress

	Bucket map[Key]string
}

func main() {

	return
}

func hashString(elt string) *big.Int {

	hasher := sha1.New()
	hasher.Write([]byte(elt))

	return new(big.Int).SetBytes(hasher.Sum(nil))

}

func jump() {

}

func between() {

}

func getLocalAddress() {

}

func findSuccessor() {

}

func create() {

}

func join() {

}

func stabilize() {

}

func notify() {

}

func fix_fingers() {

}

func checkPredecessor() {

}

func run() {

	//start an RPC server over http
	// define request and reply structs
	// handle requests appropriatley

}

//RPC call
func call() error {

	return nil
}
