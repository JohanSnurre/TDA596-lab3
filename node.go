package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
)

type Key string
type NodeAddress string

type Node struct {
	id          int
	successors  []int
	predecessor *Node
	fingers     []int
}

func (n *Node) HandlePing(arguments *Args, reply *Reply) error {
	fmt.Println("inged")
	if arguments.Command == "Ping" {
		reply.Reply = "Ping reply"
	}
	return nil
}

func (n *Node) HandleFind(arguments *Args, reply *Reply) error {
	id := arguments.NodeID
	succ := n.find(id)
	reply.Successor = succ.id
	reply.FoundSucc = true
	return nil

}

func NewNode(id int, maxSucc int) *Node {

	//node := Node{NodeAddress(Address), []NodeAddress{}, "", []NodeAddress{}, make(map[Key]string)}
	node := Node{}
	node.id = id
	node.successors = make([]int, maxSucc)
	node.predecessor = nil

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

func (n *Node) closestPrecedingNode(id int) *Node {
	return n
}

func (n *Node) findSuccessor(id int) (bool, *Node) {
	if n.id == n.successors[0] {
		return true, n
	}

	if id > n.id && id <= n.successors[0] {
		return true, n
	}

	return false, n.closestPrecedingNode(id)

}

func (n *Node) create() {
	n.predecessor = nil
	n.successors[0] = n.id

}

func (n *Node) find(id int) *Node {
	found := false
	nextNode := n

	for {
		if found {
			break
		}

		found, nextNode = nextNode.findSuccessor(id)

	}

	return nextNode

}

func (n *Node) join(existing *Node) {
	n.predecessor = nil
	// succ := existing.find(n)
	// n.successors[0] = succ
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
