package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
)

var flag, created bool
var addressPort string = ":3410"

func main() {
	mode := os.Args[1]
	id, _ := strconv.Atoi(os.Args[2])

	node := NewNode(id, 4)

	if mode == "c" {
		create(node)
		for {

		}
	} else { // join
		//node.join()
		handleJoin(node)

	}
}

func handleJoin(node *Node) {
	// find successor
	reply := Reply{}
	arguments := Args{"Ping", node.id}

	found := false
	nextNode, _ := strconv.Atoi(addressPort)

	for {

		call(addressPort, "Node.HandleFind", &arguments, &reply)
		nextNode = reply.Successor
		found = reply.FoundSucc

		if found {
			break
		}

	}
	succ := nextNode
	fmt.Println("new succ is: " + strconv.Itoa(succ))

	node.successors[0] = node
}

func create(node *Node) {
	if created {
		fmt.Println("Node already created")
		return
	}

	//node := NewNode(getLocalAddress())
	node.create()

	rpc.Register(node)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", addressPort)
	if err != nil {
		panic("Error starting RPC")
	}

	go http.Serve(l, nil)
	created = true
	fmt.Println("Created node at address: " + getLocalAddress() + addressPort)

}

func call(address string, method string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		fmt.Println("Could not connect, return false")
		return false
	}
	defer c.Close()

	err = c.Call(method, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false

}

func put(args []string) {

}

func putrandom(args []string) {

}

func get(args []string) {

}

func delete(args []string) {

}

func dump(args []string) {

}

// func ping() {

// 	reply := Reply{}
// 	arguments := Args{"Ping", 9}

// 	ok := call(addressPort, "Node.HandlePing", &arguments, &reply)
// 	if ok {
// 		fmt.Printf("reply: %v\n", reply.Reply)
// 	} else {
// 		fmt.Printf("Call failed\n")
// 	}
// 	r := reply.Reply
// 	fmt.Println(r)
// }

func getLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
