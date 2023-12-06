package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

var running, created bool
var localPort string
var node *Node
var add NodeAddress

var address *string
var addressPort *int
var joinAddress *string
var joinPort *int
var timeStablize *int
var timeFixFingers *int
var timeCheckPredecessor *int
var successorAmount *int
var identifier *string

func main() {

	address = flag.String("a", "", "Address")
	addressPort = flag.Int("p", -1, "Port")
	joinAddress = flag.String("ja", "", "Join address")
	joinPort = flag.Int("jp", -1, "Join port")
	timeStablize = flag.Int("ts", -1, "Time before stabilize call")
	timeFixFingers = flag.Int("tff", -1, "Time before fix fingers call")
	timeCheckPredecessor = flag.Int("tcp", -1, "Time before check predecessor is called")
	successorAmount = flag.Int("r", -1, "The amount of immediate successor stored")
	identifier = flag.String("i", "", "The string identifier of a node")

	flag.Parse()
	*address = strings.TrimSpace(*address)
	*joinAddress = strings.TrimSpace(*joinAddress)
	*identifier = strings.TrimSpace(*identifier)
	localPort = ":" + strconv.Itoa(*addressPort)

	//fmt.Printf("%s, %d", *joinAddress, *joinPort)

	if (len(*joinAddress) == 0 && *joinPort >= 0) || (len(*joinAddress) > 0 && *joinPort < 0) {
		fmt.Printf("You have to provide both -ja and -jp flags")
		return
	}
	if (*addressPort < 0 || *timeStablize < 1 || *timeCheckPredecessor < 1 || *timeFixFingers < 1) ||
		(*timeStablize > 60000 || *timeCheckPredecessor > 60000 || *timeFixFingers > 60000) {
		fmt.Println("Invalid arguments")
		return
	}

	add = NodeAddress(*address + localPort)
	node = &Node{Address: NodeAddress(add), Successors: []NodeAddress{}, Predecessor: "", FingerTable: []NodeAddress{}, Bucket: make(map[Key]string)}

	server(*address, ":"+strconv.Itoa(*addressPort))

	//if not joining a ring

	//if joining a ring

	//fmt.Printf("%v %v %v %v %v %v %v %v %v ", *address, *addressPort, *joinAddress, *joinPort, *timeStablize, *timeFixFingers, *timeCheckPredecessor, *successorAmount, *identifier)

	if len(*joinAddress) > 0 && *joinPort > 0 {
		//Joining an existing ring
		add := NodeAddress(*joinAddress + ":" + strconv.Itoa(*joinPort))
		join(add)
	} else {
		//Creating a new ring
		args := []string{*address + ":" + strconv.Itoa(*addressPort)}
		create(args)

	}

	go loopCP(time.Duration(*timeCheckPredecessor))
	go loopStab(time.Duration(*timeStablize))

	res := bufio.NewReader(os.Stdin)
	var s string
	running = true
	created = false

	m := make(map[string]func([]string))
	m["help"] = help
	m["quit"] = quit
	//m["port"] = port
	//m["create"] = create
	//m["ping"] = ping
	m["dump"] = dump
	m["notify"] = notify
	m["stab"] = stabilize
	m["cp"] = cp
	m["c"] = c
	for running {

		fmt.Print("::> ")
		s, _ = res.ReadString('\n')

		s = strings.TrimSpace(s)
		args := strings.Split(s, " ")

		f, ok := m[args[0]]
		if ok {
			f(args)
		}

	}

	return
}

func loopCP(t time.Duration) {
	for {
		cp([]string{})
		time.Sleep(t * time.Millisecond)
	}

}

func loopFF(t time.Duration) {

	time.Sleep(t * time.Millisecond)
}

func loopStab(t time.Duration) {

	for {
		//fmt.Println("LOL")
		stabilize([]string{})
		time.Sleep(t * time.Millisecond)

	}

}

func help(args []string) {
	fmt.Print("Help arrived!\n")
}

func quit(args []string) {
	running = false
	fmt.Print("Quitting!\n")
}

func cp(args []string) {

	arguments := Args{"CP", string(node.Address)}
	reply := Reply{}

	if string(node.Predecessor) == "" {
		return
	}

	ok := call(string(node.Predecessor), "Node.HandlePing", &arguments, &reply)
	if !ok {
		node.mu.Lock()
		fmt.Println("Can not connect to predecessor")
		node.Predecessor = NodeAddress("")
		node.mu.Unlock()
		return
	}

}

func c(args []string) {
	node.Predecessor = ""
}

func stabilize(args []string) {

	arguments := Args{"", string(node.Address)}
	reply := Reply{}

	ok := call(string(node.Successors[0]), "Node.Get_predecessor", &arguments, &reply)
	if !ok {
		fmt.Println("Could not connect to successor")
		dump([]string{})
		node.mu.Lock()
		node.Successors = node.Successors[1:]
		node.mu.Unlock()
		return
	}
	node.mu.Lock()
	addH := hashAddress(node.Address)
	addressH := hashAddress(NodeAddress(reply.Reply))
	succH := hashAddress(node.Successors[0])

	/*

		1. Ask successor for their predecessor
		2. See if the predecessor is within the range(self, successor)
		3. If it is within range, self.sucessor = sucessor.predecessor
		4. Ask sucessor for their sucessor table
		5. Append the successor table onto self.successor table.
		6. If the table exceeds size r, remove the last element


	*/

	if between(addH, addressH, succH, false) && reply.Reply != "" {

		//fmt.Println(addH)
		//fmt.Println(addressH)
		//fmt.Println(succH)
		//fmt.Println(between(addH, addressH, succH, false))

		//node.Successors = []NodeAddress{NodeAddress(reply.Reply)}
		//fmt.Println(reply.Successors)
		node.Successors = []NodeAddress{NodeAddress(reply.Reply)}
		//node.Successors = append(node.Successors, reply.Successors...)

	}

	node.mu.Unlock()
	arguments = Args{"", string(node.Address)}
	reply = Reply{}
	ok = call(string(node.Successors[0]), "Node.Get_successors", &arguments, &reply)
	if !ok {
		//fmt.Println("Call failed to successor in stabilize <2>")
	}
	node.mu.Lock()
	//fmt.Println(reply.Successors)
	node.Successors = []NodeAddress{node.Successors[0]}
	node.Successors = append(node.Successors, reply.Successors[:len(reply.Successors)]...)
	if len(node.Successors) > *successorAmount {
		node.Successors = node.Successors[:*successorAmount]
	}
	node.mu.Unlock()

	//dump([]string{})

	arguments = Args{"Stabilize", string(node.Address)}
	reply = Reply{}
	notify([]string{})

}

func notify(args []string) {

	arguments := Args{"Notify", string(node.Address)}
	reply := Reply{}

	ok := call(string(node.Successors[0]), "Node.Notify", &arguments, &reply)
	if !ok {
		//fmt.Println("Call failed in notify")
	}

	/*
		arguments := Args{"Notify", string(node.Address)}
		reply := Reply{}

		mu.Lock()
		call(string(node.Successors[0]), "Node.Notify", &arguments, &reply)
		mu.Unlock()
	*/
	//fmt.Println(reply.Reply)
}

func port(args []string) {
	if len(args) < 2 {
		fmt.Println("Not enough arguments for port!")
	}
	if created {
		fmt.Println("Node already created")
		return
	}

	localPort = ":" + args[1]
	fmt.Println(args[1])
}

func server(address string, port string) {

	rpc.Register(node)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", port)
	if err != nil {
		panic("Error starting RPC")
	}

	go http.Serve(l, nil)
	fmt.Println("Created node at address: " + address + localPort)

}

func create(args []string) {
	if created {
		fmt.Println("Node already created")
		return
	}

	node.create()

	//notify([]string{})
}

func join(address NodeAddress) {

	reply := Reply{}
	args := Args{"", string(node.Address)}

	add := address

	flag := false

	for !flag {

		call(string(add), "Node.Find_successor", &args, &reply)
		//fmt.Println(reply.Reply)

		switch found := reply.Found; found {
		case true:
			node.join(NodeAddress(reply.Reply))
			//fmt.Println("True")
			flag = true
			break
		case false:
			add = NodeAddress(reply.Forward)
			//fmt.Println("False")
			break

		}

	}

	//notify([]string{})
}

func call(address string, method string, args interface{}, reply interface{}) bool {

	//then := time.Now().Nanosecond()
	c, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		//fmt.Println("Could not connect, return false")
		return false
	}
	defer c.Close()

	err = c.Call(method, args, reply)

	//now := time.Now().Nanosecond()
	//fmt.Println(now - then)
	if err == nil {
		return true
	}

	fmt.Println("CALL ERROR: ", err)
	return false

}

func find_successor(id int) {

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

	fmt.Println("Address: " + node.Address)
	fmt.Print("Finger table: ")
	fmt.Println(node.FingerTable)
	fmt.Println("Predecessor: " + node.Predecessor)
	fmt.Print("Successors: ")
	fmt.Println(node.Successors)
	fmt.Print("Bucket: ")
	fmt.Println(node.Bucket)

}

func ping(args []string) {

	reply := Reply{}
	arguments := Args{"Ping", ""}
	fmt.Println(args[1])

	ok := call(args[1], "Node.HandlePing", &arguments, &reply)
	if ok {
		fmt.Printf("reply: %v\n", reply.Reply)
	} else {
		fmt.Printf("Call failed\n")
	}
	r := reply.Reply
	fmt.Println(r)
}

func getLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	fmt.Println(conn.LocalAddr().(*net.UDPAddr).IP.String())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
