package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"sync"
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

var generator int64
var prime int64
var SK *int

var FingerTableSize = 10

var m sync.Mutex
var connections = make(map[string]*rpc.Client)

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
	SK = flag.Int("sk", 0, "Secret key")
	encryptionKey := flag.String("e", "abcdabcdabcdabcdabcdabcdabcdabcd", "Enc key")

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
	if len(*encryptionKey) != 32 {
		fmt.Println("Provide an encryption key of 32 bytes")
		return
	}

	add = NodeAddress(*address + localPort)
	node = &Node{Address: add, Successors: []NodeAddress{}, Predecessor: "", FingerTable: []NodeAddress{}, Bucket: make(map[Key]string), encryptionKey: *encryptionKey}

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
	go loopFF(time.Duration(*timeFixFingers))

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
	m["LookUp"] = LookUp
	m["StoreFile"] = StoreFile
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

	//Generate a random prime number
	//Choose a generator for the group of that prime number

	p := int64(67)
	g := int64(12)

	PK := int64(math.Mod(math.Pow(float64(g), float64(*SK)), float64(p)))

	SendRequest(add, args[1], PK, p, g, 0)

}

func findFile(args []string) string {
	filename := args[1]

	reply := Reply{}
	arguments := Args{Command: "", Address: filename, Offset: 0}

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

	EncryptFile([]byte(node.encryptionKey), filename, filename)

	file, err = os.ReadFile(filename)
	if err != nil {
		fmt.Println("file open error: " + err.Error())
	}

	content := string(file)

	add := findFile(args)
	fmt.Println(add, node.Address)

	//if the address maps to itself then there is no need to make a call. Encryption is done
	if strings.Compare(add, string(node.Address)) == 0 {

		return
	}

	/*
		encrypt file locally
		send it



	*/

	/*
		Use the address and make an rpc connection to upload the file contents to that node

	*/

	reply := Reply{}
	arguments := Args{Command: content, Address: string(node.Address), Filename: filename, Offset: 0}

	ok := call(string(add), "Node.Store", &arguments, &reply)
	if !ok {
		fmt.Println("cano reach the node")
		return
	}

}

func loopCP(t time.Duration) {
	for {
		cp([]string{})
		time.Sleep(t * time.Millisecond)
	}

}

func next() func() int {

	next := 0
	return func() int {
		next++
		if next > 10 {
			next = 1
		}
		return next
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
	m.Lock()
	defer m.Unlock()
	fmt.Println(len(connections))
	for add, conn := range connections {
		err := conn.Close()
		if err != nil {
			fmt.Println("error closing :", add, err)
		}
	}

	fmt.Print("Quitting!\n")
}

func cp(args []string) {
	node.Lock()
	defer node.Unlock()
	arguments := Args{Command: "CP", Address: string(node.Address), Offset: 0}
	reply := Reply{}

	if string(node.Predecessor) == "" {
		return
	}

	ok := call(string(node.Predecessor), "Node.HandlePing", &arguments, &reply)
	if !ok {
		//node.mu.Lock()
		fmt.Println("Can not connect to predecessor")
		node.Predecessor = NodeAddress("")
		//node.mu.Unlock()
		return
	}

}

func fix_fingers() {
	node.Lock()
	defer node.Unlock()
	if len(node.FingerTable) == 0 {
		//node.mu.Lock()
		node.FingerTable = []NodeAddress{node.Successors[0]}
		//node.mu.Unlock()

		return
	}

	temp := []NodeAddress{}
	for next := 1; next <= FingerTableSize; next++ {
		offset := int64(math.Pow(2, float64(next)-1))
		add := node.Address
		flag := false
		for !flag {

			reply := Reply{}
			args := Args{Command: "", Address: string(node.Address), Offset: offset}

			ok := call(string(add), "Node.FindSuccessor", &args, &reply)
			if !ok {
				fmt.Println("Failed to fix fingers : ")
				return
			}
			//fmt.Println(reply.Found)

			switch found := reply.Found; found {
			case true:
				//node.mu.Lock()
				temp = append(temp, NodeAddress(reply.Reply))
				fmt.Println("SUCC: "+reply.Reply, "Offset: ", offset)
				flag = true
				//node.mu.Unlock()
				break
			case false:

				add = NodeAddress(reply.Forward)
				fmt.Println("FORWARD: " + add)
				break

			}

		}

	}
	node.FingerTable = temp

}

func stabilize(args []string) {

	node.Lock()
	defer node.Unlock()

	arguments := Args{Command: "", Address: string(node.Address), Offset: 0}
	reply := Reply{}

	ok := call(string(node.Successors[0]), "Node.Get_predecessor", &arguments, &reply)
	if !ok {
		fmt.Println("Could not connect to successor")
		dump([]string{})
		//node.mu.Lock()
		node.Successors = node.Successors[1:]
		//node.mu.Unlock()
		return
	}
	//node.mu.Lock()
	addH := hashAddress(node.Address)
	addressH := hashAddress(NodeAddress(reply.Reply))
	succH := hashAddress(node.Successors[0])

	if between(addH, addressH, succH, true) && reply.Reply != "" {
		node.Successors = []NodeAddress{NodeAddress(reply.Reply)}
	}

	//node.mu.Unlock()
	arguments = Args{Command: "", Address: string(node.Address), Offset: 0}
	reply = Reply{}
	ok = call(string(node.Successors[0]), "Node.Get_successors", &arguments, &reply)
	if !ok {
		//fmt.Println("Call failed to successor in stabilize <2>")
	}
	//node.mu.Lock()
	//fmt.Println(reply.Successors)
	node.Successors = []NodeAddress{node.Successors[0]}
	node.Successors = append(node.Successors, reply.Successors...)
	if len(node.Successors) > *successorAmount {
		node.Successors = node.Successors[:*successorAmount]
	}
	//node.mu.Unlock()

	arguments = Args{Command: "Stabilize", Address: string(node.Address), Offset: 0}
	reply = Reply{}
	notify([]string{})

	return

}

func notify(args []string) {

	arguments := Args{Command: "Notify", Address: string(node.Address), Offset: 0}
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
	args := Args{Command: "", Address: string(node.Address), Offset: 0}

	add := address

	flag := false

	for !flag {

		call(string(add), "Node.FindSuccessor", &args, &reply)
		//fmt.Println(reply.Reply)

		fmt.Println(reply.Successors)

		switch found := reply.Found; found {
		case true:
			node.join(NodeAddress(reply.Reply))
			//fmt.Println("True")
			flag = true
			break
		case false:
			add = NodeAddress(reply.Forward)
			//fmt.Println("False", add)
			break

		}

	}

}

func call(address string, method string, args interface{}, reply interface{}) bool {

	m.Lock()
	defer m.Unlock()

	c, ok := connections[address]
	if ok { // if already connection to address
		err := c.Call(method, args, reply)
		if err == nil {
			return true
		}

		fmt.Println("CALL ERROR: ", err)
		delete(connections, address)
		return false
	}

	//then := time.Now().Nanosecond()
	c, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		//fmt.Println("Could not connect, return false")
		fmt.Println("error :", err)
		return false
	}
	connections[address] = c

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

// func delete(args []string) {

// }

func dump(args []string) {

	fmt.Println("Address: " + node.Address)
	fmt.Println("ID: " + hashAddress(node.Address).String())
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

func EncryptMessage(key []byte, message string) (string, error) {
	byteMsg := []byte(message)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("could not create new cipher: %v", err)
	}

	cipherText := make([]byte, aes.BlockSize+len(byteMsg))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("could not encrypt: %v", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], byteMsg)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func EncryptFile(key []byte, filename string, out string) {

	f, err := os.Open(filename)
	if err != nil {

	}

	content, err := io.ReadAll(f)
	if err != nil {

	}

	f.Close()

	enc, err := EncryptMessage(key, string(content))
	if err != nil {

	}

	outFile, err := os.Create(out)
	if err != nil {

	}

	outFile.Write([]byte(enc))

	outFile.Close()

}

func SendRequest(address string, filename string, PK int64, prime int64, generator int64, s int64) error {

	args := Args{Filename: filename, PublicKey: PK, Prime: prime, Generator: generator}
	reply := Reply{}

	ok := call(address, "Node.HandleRequest", &args, &reply)
	if !ok {
		fmt.Println("Error requesting")
		return nil
	}

	secret := int64(math.Mod(math.Pow(float64(reply.PublicKey), float64(*SK)), float64(prime)))

	secretExt := strconv.FormatInt(secret, 10)
	for len(secretExt) < 32 {
		secretExt = secretExt + secretExt
	}
	secretExt = secretExt[:32]

	args = Args{Filename: filename, PublicKey: PK, Prime: prime, Generator: generator}
	reply = Reply{}

	ok = call(address, "Node.GetKey", &args, &reply)
	if !ok {
		fmt.Println("Error requesting")
		return nil
	}
	//fmt.Println(reply.EncKey)

	dKey, err := DecryptMessage([]byte(secretExt), reply.EncKey)
	if err != nil {
		fmt.Println("Error decrypting ", err)
		return nil
	}

	args = Args{Filename: filename, PublicKey: PK, Prime: prime, Generator: generator}
	reply = Reply{}

	ok = call(address, "Node.GetFile", &args, &reply)
	if !ok {
		fmt.Println("Error requesting")
		return nil
	}

	//fmt.Println(reply.Content)

	text, err := DecryptMessage([]byte(dKey), reply.Content)
	if err != nil {
		fmt.Println("Error decrypting ", err)
		return nil

	}
	fmt.Println(text)
	return nil

}

func DecryptMessage(key []byte, message string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return "", fmt.Errorf("could not base64 decode: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("could not create new cipher: %v", err)
	}

	if len(cipherText) < aes.BlockSize {
		return "", fmt.Errorf("invalid ciphertext block size")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}
