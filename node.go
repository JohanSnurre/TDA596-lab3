package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"strconv"
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

	encryptionKey string

	PK_A, PK_B int64
}

func (n *Node) HandlePing(arguments *Args, reply *Reply) error {
	n.mu.Lock()
	//fmt.Println("In HandlePing")
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

		//n.mu.Lock()
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

		//n.mu.Unlock()
		return nil
	}
*/
func (n *Node) closest_preceding_node(id *big.Int) NodeAddress {

	for i := len(n.FingerTable) - 1; i >= 0; i-- {

		addH := hashAddress(n.Address)
		//fmt.Println("FINGER TABLE: ", n.FingerTable)
		fingerH := hashAddress(n.FingerTable[i])

		if between(addH, fingerH, id, true) {
			return n.FingerTable[i]
		}

	}
	return n.Successors[0]

}

func (n *Node) test(id *big.Int) NodeAddress {

	fmt.Println(n.Successors)

	for i := len(n.Successors) - 1; i >= 0; i-- {

		addH := hashAddress(n.Address)
		fingerH := hashAddress(n.Successors[i])

		if between(addH, fingerH, id, true) {
			return n.Successors[i]
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
	ID.Add(ID, big.NewInt(args.Offset))
	ID.Mod(ID, big.NewInt(int64(math.Pow(2, float64(FingerTableSize)))))

	succH := hashAddress(NodeAddress(n.Successors[0]))

	//If the ID is between self and immediate successor
	if between(addH, ID, succH, false) {
		reply.Found = true
		reply.Reply = string(n.Successors[0])
		//fmt.Println(addH, reply.Reply, succH)
		//reply.Successors = n.Successors
	} else {
		//if the file is outside. Should return the closest preceding node before ID. Have to implement fix_fingers for this to work.
		//Right now it will return the next successor, jumping only 1 step on the ring. Search time is O(N), we want O(log(N))
		reply.Found = false

		/*
			forward := string(n.closest_preceding_node(ID))
			if strings.Compare(string(forward), string(n.Address)) == 0 {
				reply.Found = true
				reply.Reply = forward
			} else {
				reply.Found = false
				reply.Forward = forward
			}
		*/

		//fmt.Println(addH, ID, succH)
		reply.Forward = string(n.closest_preceding_node(ID))
		//reply.Forward = string(n.Successors[0])
		//reply.Forward = string(n.test(ID))
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

func (n *Node) Store(args *Args, reply *Reply) error {
	filename := args.Filename
	content := []byte(args.Command)
	address := args.Address

	//fmt.Println(string(content))
	/*

		Get the file

		get the decryption key

		decrypt the file

		encrypt the file


	*/

	//if the file is to be stored locally then there is no need to make a call
	if hashAddress(NodeAddress(add)) == hashAddress(node.Address) {
		return nil
	}

	PK := int64(math.Mod(math.Pow(float64(generator), float64(*SK)), float64(prime)))
	//fmt.Println(generator, *SK, prime, PK)
	arguments := Args{Prime: prime, Generator: generator, PublicKey: PK}

	ok := call(address, "Node.HandleRequest", &arguments, &reply)
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

	arguments = Args{Filename: filename, PublicKey: PK, Prime: prime, Generator: generator}
	rep := Reply{}

	ok = call(address, "Node.GetKey", &arguments, &rep)
	if !ok {
		fmt.Println("Error requesting")
		return nil
	}

	dKey, err := DecryptMessage([]byte(secretExt), rep.EncKey)
	if err != nil {
		fmt.Println("Error decrypting ", err)
		return nil
	}

	text, err := DecryptMessage([]byte(dKey), string(content))
	if err != nil {
		fmt.Println("Error decrypting ", err)
		return nil

	}

	encTxt, err := EncryptMessage([]byte(n.encryptionKey), string(text))

	err = os.WriteFile(filename, []byte(encTxt), 0777)
	if err != nil {
		fmt.Println("problem writing file")
	}

	return nil
}

func (n *Node) HandleRequest(args *Args, reply *Reply) error {

	//filename := args.Filename
	generator = args.Generator
	prime = args.Prime

	//fmt.Println(args.PublicKey)

	n.PK_B = args.PublicKey

	ret := math.Mod(math.Pow(float64(generator), float64(*SK)), float64(prime))
	reply.PublicKey = int64(ret)

	return nil
}

func (n *Node) GetFile(args *Args, reply *Reply) error {

	f, err := os.Open(args.Filename)
	if err != nil {
		return nil
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return nil
	}

	reply.Content = string(content)
	return nil
}

func (n *Node) GetKey(args *Args, reply *Reply) error {

	secret := int64(math.Mod(math.Pow(float64(n.PK_B), float64(*SK)), float64(args.Prime)))

	//fmt.Println("PK: ", n.PK_B)
	//fmt.Println("Secret: ", secret)

	secretExt := strconv.FormatInt(secret, 10)
	for len(secretExt) < 32 {
		secretExt = secretExt + secretExt
	}
	secretExt = secretExt[:32]

	//fmt.Println(secretExt)

	eKey, err := EncryptMessage([]byte(secretExt), n.encryptionKey)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	_, err = DecryptMessage([]byte(secretExt), eKey)
	if err != nil {
		fmt.Println("Error decrypting ", err)
		return nil
	}

	//fmt.Println(eKey, dKey)

	reply.EncKey = eKey

	return nil
}

func (n *Node) checkPredecessor() {

}

func run() {

	//start an RPC server over http
	// define request and reply structs
	// handle requests appropriatley

}
