package main

import (
	"bufio"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
)

type Key string
type NodeAddress string

type Node struct {
	Id          *big.Int
	Address     NodeAddress
	FingerTable []NodeAddress
	Predecessor NodeAddress
	Successors  []NodeAddress
	Ts          int
	Tff         int
	Tcp         int

	Bucket map[Key]string
}

func (n Node) jump() {

}

func (n Node) between() {

}

func (n *Node) FindSuccessor(args *Args, reply *ReplyFindSuc) error {
	reply.IsFound = true
	reply.Reply = fmt.Sprint(n.Address)

	// if iSuccessor {
	// 	reply.IsFound = true
	// 	reply.Reply = fmt.Sprint(n.Address)
	// } else {
	// 	reply.IsFound = false
	// 	reply.Reply = newAddress
	// }
	return nil
}

func (n *Node) create() {
	n.Predecessor = ""
	n.Successors[0] = n.Address
	rpc.Register(n)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", string(n.Address))
	if err != nil {
		fmt.Println(err)
		panic("Error with Rpc listen side ")
	}

	go http.Serve(l, nil)

}

func (n *Node) join(toJoinAddr NodeAddress) {
	n.Predecessor = ""

	rpc.Register(n)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", string(n.Address))
	if err != nil {
		fmt.Println(err)
		panic("Error with Rpc listen side ")
	}

	go http.Serve(l, nil)

	reply := ReplyFindSuc{}
	arg := Args{n.Id}
	noError := n.call(toJoinAddr, "Node.FindSuccessor", &arg, &reply)
	for noError && !reply.IsFound {
		arg.Id, _ = new(big.Int).SetString(reply.Reply, 10)
		noError = n.call(toJoinAddr, "Node.FindSuccessor", &arg, &reply)
	}
	if noError {
		n.Successors[0] = NodeAddress(reply.Reply)
	}

}

func (n Node) stabilize() {

}

func (n Node) notify() {

}

func (n Node) fix_fingers() {

}

func (n Node) checkPredecessor() {

}

func (n Node) run() {
	input := bufio.NewReader(os.Stdin)

	reading := true

	for reading {
		fmt.Print("> ")
		text, _ := input.ReadString('\n')
		text = strings.TrimSpace(text)
		temp := strings.Split(text, " ")

		switch temp[0] {
		case "Lookup":
			fmt.Println("Lookup")
		case "StoreFile":
			fmt.Println("StoreFile")
		case "PrintState":
			fmt.Println("PrintState :")
			fmt.Println("Successor :", n.Successors)
		case "quit":
			reading = false
			fmt.Println("Quitting...")
		default:
			fmt.Println("Unknown command :", text)

		}
	}

	return

}

// RPC call
func (n Node) call(address NodeAddress, method string, args interface{}, reply interface{}) bool {
	client, err := rpc.DialHTTP("tcp", string(address))
	if err != nil {
		fmt.Println(err)
		fmt.Println("Error rpc dialing side")
		return false
	}
	defer client.Close()
	err = client.Call(method, args, reply)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Error rpc call")
		return false
	}
	return true
}
