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

func (n *Node) HandlePing(arguments *Args, reply *Reply) error {
	if arguments.Command == "Ping" {
		reply.Reply = "Ping reply"
	}
	return nil
}

func NewNode(Address string) *Node {

	node := Node{NodeAddress(Address), []NodeAddress{}, "", []NodeAddress{}, make(map[Key]string)}

	return &node

}

func hashString(elt string) *big.Int {

	hasher := sha1.New()
	hasher.Write([]byte(elt))

	return new(big.Int).SetBytes(hasher.Sum(nil))

}

func (n *Node) jump() {

}

func (n *Node) between() {

}

func (n *Node) getLocalAddress() {

}

func (n *Node) findSuccessor() {

}

func (n *Node) create() {

	/*
		Create a new ring if none already exist
		return error if a ring already exists


	*/

}

func (n *Node) join(address NodeAddress) {

}

func (n *Node) stabilize() {

}

func (n *Node) notify() {

}

func (n *Node) fix_fingers() {

}

func (n *Node) checkPredecessor() {

}

func run() {

	//start an RPC server over http
	// define request and reply structs
	// handle requests appropriatley

}
