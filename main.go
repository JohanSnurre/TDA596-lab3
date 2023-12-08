package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
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
	//go loopFF(time.Duration(*timeFixFingers))

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
	m["StoreFile"] = StoreFile
	m["LookUp"] = LookUp
	m["PrintState"] = PrintState
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

func LookUp(args []string) {
	add := findFile(args)
	fmt.Println(add)
}

func findFile(args []string) string {
	filename := args[1]

	reply := Reply{}
	arguments := Args{"", filename, 0}

	add := node.Address
	flag := false

	for !flag {
		ok := call(string(add), "Node.FindSuccessor", &arguments, &reply)
		if !ok {
			fmt.Println("Failed to fix fingers")
			return ""
		}
		switch found := reply.Found; found {

		//if the file maps between self and successor then reply.Reply = node.Successor[0]
		case true:

			flag = true
			break
		//if the file maps somewhere else then we have to forward the request to a better node
		case false:
			add = NodeAddress(reply.Forward)
			break
		}

	}

	//Print out the correct address to store the file in. Dependent on hashString(filename)
	//fmt.Println(reply.Reply)
	return reply.Reply

}

func StoreFile(args []string) {

	filename := args[1]

	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("file open error: " + err.Error())
	}

	content := string(file)

	id := findFile(args)

	/*
		Use the address and make an rpc connection to upload the file contents to that node

	*/

	reply := Reply{}
	arguments := Args{content, filename, 0}

	ok := call(string(id), "Node.Store", &arguments, &reply)
	if !ok {
		fmt.Println("cano reach the node")
		return
	}

}

func PrintState(args []string) {
	// dump(args)
	node.mu.Lock()
	fmt.Println(getNodeInfo(node.Address))

	fmt.Println("SUCCESSORS:")
	for _, s := range node.Successors {
		fmt.Println(getNodeInfo(s))
	}

	fmt.Println("FINGER TABLE:")
	for _, f := range node.FingerTable {
		fmt.Println(getNodeInfo(f))
	}

	node.mu.Unlock()
}

func getNodeInfo(address NodeAddress) string {
	id := hashAddress(address)
	return id.String() + " => " + string(address)

}

func loopCP(t time.Duration) {
	for {
		cp([]string{})
		time.Sleep(t * time.Millisecond)
	}

}

func loopFF(t time.Duration) {
	for {
		fix_fingers()
		time.Sleep(t * time.Millisecond)
	}

}

func loopStab(t time.Duration) {

	for {
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

	arguments := Args{"CP", string(node.Address), 0}
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

func fix_fingers() {

	if len(node.FingerTable) == 0 {
		node.FingerTable = []NodeAddress{node.Successors[0]}

		return
	}

	temp := []NodeAddress{}
	node.FingerTable = []NodeAddress{}
	for next := 0; next < 10; next++ {

		offset := int64(math.Pow(2, float64(next)))
		add := node.Address
		flag := false
		for !flag {

			reply := Reply{}
			args := Args{"", string(node.Address), offset}

			ok := call(string(add), "Node.FindSuccessor", &args, &reply)
			if !ok {
				fmt.Println("Failed to fix fingers")
				return
			}
			//fmt.Println(reply.Found)

			switch found := reply.Found; found {
			case true:
				temp = append(temp, NodeAddress(reply.Reply))
				//fmt.Println("SUCC: " + reply.Reply)
				flag = true
				break
			case false:
				//fmt.Println("FORWARD: " + reply.Forward)
				add = NodeAddress(reply.Forward)
				break

			}

		}

	}

	node.FingerTable = temp

}

func stabilize(args []string) {

	arguments := Args{"", string(node.Address), 0}
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

	if between(addH, addressH, succH, false) && reply.Reply != "" {
		node.Successors = []NodeAddress{NodeAddress(reply.Reply)}
	}

	node.mu.Unlock()
	arguments = Args{"", string(node.Address), 0}
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

	arguments = Args{"Stabilize", string(node.Address), 0}
	reply = Reply{}
	notify([]string{})

}

func notify(args []string) {

	arguments := Args{"Notify", string(node.Address), 0}
	reply := Reply{}

	ok := call(string(node.Successors[0]), "Node.Notify", &arguments, &reply)
	if !ok {
		//fmt.Println("Call failed in notify")
	}

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

}

func join(address NodeAddress) {

	reply := Reply{}
	args := Args{"", string(node.Address), 0}

	add := address

	flag := false

	for !flag {

		call(string(add), "Node.FindSuccessor", &args, &reply)
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
