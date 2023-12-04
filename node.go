package main

import (
	"crypto/sha1"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

type Key string
type NodeAddress string

type Node struct {
	Address     NodeAddress
	FingerTable []NodeAddress
	Predecessor NodeAddress
	Successors  []NodeAddress

	Bucket map[Key]string

	port string
}

func (n *Node) server() {

	rpc.Register(n)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", n.port)
	if err != nil {
		panic("Error starting RPC")
	}

	go http.Serve(l, nil)

}

func NewNode(Address string, port string) *Node {

	node := Node{NodeAddress(Address), []NodeAddress{}, "", []NodeAddress{}, make(map[Key]string), port}

	node.server()
	return &node

}

func hashString(elt string) *big.Int {

	hasher := sha1.New()
	hasher.Write([]byte(elt))

	return new(big.Int).SetBytes(hasher.Sum(nil))

}

func (n *Node) setPort(port int) {
	n.port = ":" + strconv.Itoa(port)
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

//RPC call
// define request and reply interfaces in a separate RPC file
func call(address string, method string, request interface{}, reply interface{}) error {

	return nil
}
