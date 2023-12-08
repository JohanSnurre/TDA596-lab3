package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"os"
	"sync"
)

type Key string
type NodeAddress string

type Node struct {
	mu          sync.Mutex
	Address     NodeAddress
	FingerTable []NodeAddress
	Predecessor NodeAddress
	Successors  []NodeAddress

	Bucket map[Key]string
}

func (n *Node) HandlePing(arguments *Args, reply *Reply) error {
	n.mu.Lock()
	//fmt.Println("In HandlePing")
	if arguments.Offset == 69 {
		fmt.Println("good ping")
	}
	if arguments.Command == "CP" {
		reply.Reply = "CP reply"
	}
	n.mu.Unlock()
	return nil
}

/*func NewNode(Address string) *Node {

	node := Node{NodeAddress(Address), []NodeAddress{}, "", []NodeAddress{}, make(map[Key]string)}

	return &node

}*/

func between(start *big.Int, elt *big.Int, end *big.Int, inclusive bool) bool {

	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}

}

func (n *Node) Get_predecessor(args *Args, reply *Reply) error {
	n.mu.Lock()
	reply.Reply = string(node.Predecessor)
	n.mu.Unlock()
	return nil

}

/*
func (n *Node) Find_successor(args *Args, reply *Reply) error {

		n.mu.Lock()
		addressH := hashAddress(NodeAddress(args.Address))

		addressH = addressH.Add(addressH, big.NewInt(args.Offset))

		found, address := n.findSuccessor(addressH)

		if found {
			reply.Found = true
			reply.Reply = string(address)
			reply.Successors = n.Successors
			//fmt.Println(n.Successors)
			reply.Forward = ""
		} else {
			reply.Found = false
			reply.Reply = ""
			reply.Forward = string(n.Successors[0])
			//fmt.Println(reply.Reply)
		}

		n.mu.Unlock()
		return nil
	}
*/
func (n *Node) closest_predecing_node(id *big.Int) NodeAddress {

	for i := len(n.FingerTable) - 1; i >= 0; i-- {

		addH := hashAddress(n.Address)
		fingerH := hashAddress(n.FingerTable[i])

		if between(addH, fingerH, id, false) {
			return n.FingerTable[i]
		}

	}
	return n.Address

}

func hashAddress(elt NodeAddress) *big.Int {

	hasher := sha1.New()
	hasher.Write([]byte(elt))

	t := new(big.Int).SetBytes(hasher.Sum(nil))

	return new(big.Int).Mod(t, big.NewInt(int64(1024)))
	//return new(big.Int).SetBytes(hasher.Sum(nil))

}

func hashString(elt string) *big.Int {

	hasher := sha1.New()
	hasher.Write([]byte(elt))

	t := new(big.Int).SetBytes(hasher.Sum(nil))

	return new(big.Int).Mod(t, big.NewInt(int64(1024)))
	//return new(big.Int).SetBytes(hasher.Sum(nil))

}

func (n *Node) jump() {

}

func (n *Node) between() {

}

func (n *Node) getLocalAddress() {

}

func (n *Node) FindSuccessor(args *Args, reply *Reply) error {

	n.mu.Lock()
	addH := hashAddress(n.Address)

	ID := hashAddress(NodeAddress(args.Address))

	succH := hashAddress(NodeAddress(n.Successors[0]))

	//If the ID is between self and immediate successor
	if between(addH, ID, succH, true) {
		reply.Found = true
		reply.Reply = string(n.Successors[0])
		//reply.Successors = n.Successors
	} else {
		//if the file is outside. Should return the closest preceding node before ID. Have to implement fix_fingers for this to work.
		//Right now it will return the next successor, jumping only 1 step on the ring. Search time is O(N), we want O(log(N))
		reply.Found = false
		reply.Reply = ""
		reply.Forward = string(n.Successors[0])
	}

	n.mu.Unlock()
	return nil

}

func (n *Node) Get_successors(args *Args, reply *Reply) error {

	n.mu.Lock()
	reply.Successors = node.Successors
	n.mu.Unlock()
	return nil

}

func (n *Node) create() {

	n.mu.Lock()
	n.Predecessor = ""
	n.Successors = append(n.Successors, n.Address)
	n.mu.Unlock()

}

func (n *Node) join(address NodeAddress) {
	n.mu.Lock()
	node.Predecessor = ""
	node.Successors = []NodeAddress{address}

	n.mu.Unlock()

}

func (n *Node) stabilize() {

}

func (n *Node) Notify(args *Args, reply *Reply) error {

	n.mu.Lock()
	addH := hashAddress(NodeAddress(args.Address))

	addressH := hashAddress(n.Address)

	preH := hashAddress(NodeAddress(n.Predecessor))

	if n.Predecessor == "" || (between(preH, addH, addressH, false)) {
		n.Predecessor = NodeAddress(args.Address)
		reply.Reply = "Success"
	} else {
		reply.Reply = "Fail"
	}
	n.mu.Unlock()
	return nil
}

func (n *Node) fix_fingers() {
	/*
		next := 1

		addressH := hashAddress(n.Address)

		for {
			if next > 5 {
				break
			}

			n.FingerTable[next] =

			next = next + 1

		}
	*/

}

func (n *Node) checkPredecessor() {

}

func (n *Node) Store(args *Args, reply *Reply) error {
	f := args.Address
	content := []byte(args.Command)

	filename := f + string(n.Address)

	err := os.WriteFile(filename, content, 0777)
	if err != nil {
		fmt.Println("problem writing file")
	}

	return nil
}

func run() {

	//start an RPC server over http
	// define request and reply structs
	// handle requests appropriatley

}
