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

func between(start *big.Int, elt *big.Int, end *big.Int, inclusive bool) bool {

	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}

}

func (n *Node) Get_predecessor(args *Args, reply *Reply) error {

	reply.Reply = string(node.Predecessor)
	return nil

}

func (n *Node) Find_successor(args *Args, reply *Reply) error {

	addressH := hashAddress(NodeAddress(args.Address))

	addH := hashAddress(node.Address)

	succH := hashAddress(NodeAddress(n.Successors[0]))

	if between(addH, addressH, succH, true) {
		reply.Found = true
		reply.Reply = string(n.Successors[0])
		reply.Forward = ""
	} else {
		reply.Found = false
		reply.Reply = ""
		reply.Forward = string(n.Successors[0])
		//fmt.Println(reply.Reply)
	}

	return nil
}

func hashAddress(elt NodeAddress) *big.Int {

	hasher := sha1.New()
	hasher.Write([]byte(elt))

	return new(big.Int).SetBytes(hasher.Sum(nil))

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

func (n *Node) findSuccessor(m *Node) NodeAddress {

	return ""

}

func (n *Node) create() {

	n.Predecessor = ""
	n.Successors = append(n.Successors, n.Address)

	/*
		Create a new ring if none already exist
		return error if a ring already exists


	*/

}

func (n *Node) join(address NodeAddress) {

}

func (n *Node) stabilize() {

}

func (n *Node) Notify(args *Args, reply *Reply) error {

	addH := hashAddress(NodeAddress(args.Address))

	addressH := hashAddress(n.Address)

	preH := hashAddress(NodeAddress(n.Predecessor))

	if n.Predecessor == "" || (between(preH, addH, addressH, false)) {
		n.Predecessor = NodeAddress(args.Address)
		reply.Reply = "Success"
	} else {
		reply.Reply = "Fail"
	}

	return nil
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
