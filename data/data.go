package main

import (
	"fmt"
	"math/rand"
	"time"
)

type NodeType int

const (
	Input  NodeType = iota // Input = 0
	Hidden                 // Hidden = 1
	Output                 // Output = 2
)

func (t NodeType) String() string {
	// helper funtion
	// for more readable testing
	switch t {
	case Input:
		return "Input"
	case Output:
		return "Output"
	default:
		return "Hidden"
	}
}

type Node struct {
	ID            int // unique node ID
	Type          NodeType
	IncomingConns []Connection // list of connections entering the node
}

type Connection struct {
	InNode     *Node
	OutNode    *Node
	Weight     float64
	Innovation int
	Enabled    bool
}

type InnovationKey struct {
	// key for the InnovationHistory map
	// recognizes connection by IDs of nodes that make it up
	inNodeID  int
	outNodeID int
}

type InnovationHistory struct {
	// map, that gives each connection unique innovation number
	// tracks innovation globally
	History map[InnovationKey]int
	Counter int
}

func (ih *InnovationHistory) getInnovation(inNode, outNode *Node) int {
	// given nodes, returns innovation nubmer of the connection between them
	key := InnovationKey{inNodeID: inNode.ID, outNodeID: outNode.ID}
	if inno, exist := ih.History[key]; exist {
		return inno
		// if the connection hadn't existed, function gives it new innovation number
		// and updates global innovation counter
	}
	ih.Counter++
	ih.History[key] = ih.Counter
	return ih.History[key]
}

func (genom *Genom) addConnetion(node1, node2 *Node, weight float64, enabled bool) {
	// helper function
	// given nodes, weight and enabled, adds specific connection to the genome
	inno := genom.IH.getInnovation(node1, node2)
	newConn := Connection{
		InNode:     node1,
		OutNode:    node2,
		Weight:     weight,
		Innovation: inno,
		Enabled:    enabled,
	}
	genom.Connections = append(genom.Connections, newConn)
	// adds connection to the node right away
	node2.IncomingConns = append(node2.IncomingConns, newConn)
}

func (genom *Genom) connectionExist(inNode, outNode *Node) bool {
	// helper function
	// checks, if the connection already exist in the genome
	for _, conn := range genom.Connections {
		if conn.InNode.ID == inNode.ID && conn.OutNode.ID == outNode.ID {
			return true
		}
	}
	return false
}

func (genom *Genom) randomNodes() (*Node, *Node) {
	// helper function
	// returns two nodes from the genome, which can make connection n1 –> n2
	rand.Seed(time.Now().UnixNano())
	n1 := genom.Nodes[rand.Intn(len(genom.Nodes))]
	n2 := genom.Nodes[rand.Intn(len(genom.Nodes))]

	// makes sure the connection will be made in the valid direction
	for n1.Type == Hidden && n2.Type == Hidden {
		n2 = genom.Nodes[rand.Intn(len(genom.Nodes))]
	}

	for n2.Type == Input {
		n2 = genom.Nodes[rand.Intn(len(genom.Nodes))]
	}

	for n1.Type == Output {
		n1 = genom.Nodes[rand.Intn(len(genom.Nodes))]
	}

	return n1, n2
}

type Genom struct {
	numInputs        int                // number of Input nodes
	numOutputs       int                // number of Output nodes
	totalNodes       int                // total number of nodes
	Nodes            []*Node            // list of all nodes in the genome
	Connections      []Connection       // list of all connections in the genome
	ConnCreationRate float64            // chance of adding connection while creating new network
	IH               *InnovationHistory // global innovation history
	Fitness          float32            // fitness score
}

func (genom *Genom) createNetwork() {
	// creates new network with only input and output nodes
	// adds connections randomly

	// adding inputs nodes
	for i := 0; i < genom.numInputs; i++ {
		genom.Nodes = append(genom.Nodes, &Node{ID: genom.totalNodes, Type: Input})
		genom.totalNodes++
	}
	// adding output nodes
	for i := 0; i < genom.numOutputs; i++ {
		genom.Nodes = append(genom.Nodes, &Node{ID: genom.totalNodes, Type: Output})
		genom.totalNodes++
	}
	// adding random connections between nodes
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < genom.numInputs*genom.numOutputs; i++ {
		if rand.Float64() < genom.ConnCreationRate {
			node1, node2 := genom.randomNodes()
			if !genom.connectionExist(node1, node2) {
				weight := rand.Float64()
				genom.addConnetion(node1, node2, weight, true)
			}
		}
	}
}

func (genom *Genom) mutateWeight() {
	// mutates genome by changing weights of genome's connections
	rand.Seed(time.Now().UnixNano())
	for i, conn := range genom.Connections {
		if rand.Float64() < 0.8 {
			delta := rand.Float64()*0.4 - 0.2
			conn.Weight += delta
		} else {
			conn.Weight = rand.Float64()*2.0 - 1.0
		}
		genom.Connections[i] = conn
	}
}

func (genom *Genom) mutateAddConnection() {
	// mutates genome by adding new connection with random weight
	rand.Seed(time.Now().UnixNano())
	n1, n2 := genom.randomNodes()
	if !genom.connectionExist(n1, n2) {
		weight := rand.Float64()
		genom.addConnetion(n1, n2, weight, true)
	}
}

func (genom *Genom) mutateAddNode() {
	// mutates genome by adding new hidden node
	// splits old connection by adding new hidden node and two new connections
	if len(genom.Connections) == 0 {
		return
	}
	// takes random connection from the genome and disables it
	// (the connection is still kept in the genome)
	rand.Seed(time.Now().UnixNano())
	conn := &genom.Connections[rand.Intn(len(genom.Connections))]
	n1 := conn.InNode
	n2 := conn.OutNode
	conn.Enabled = false

	// creates new hidden node
	genom.totalNodes++
	newNode := Node{ID: genom.totalNodes - 1, Type: Hidden}
	// creates two new connections
	// weighs are chosen in such way, two new connections behave in the same way as old one
	// this way mutation is not too drastic
	genom.addConnetion(n1, &newNode, 1.0, true)
	genom.addConnetion(&newNode, n2, conn.Weight, true)
	genom.Nodes = append(genom.Nodes, &newNode)
}

func crossover(parent1, parent2 *Genom) *Genom {
	// creating offspring genome
	// networks's structure is inherited from the parent with higher fitness score

	// making sure parent1 has higher fitness score
	if parent2.Fitness > parent1.Fitness {
		tmp := parent1
		parent1 = parent2
		parent2 = tmp
	}
	// mapping parent2 connections by their innovation score
	parent2Conns := make(map[int]Connection)
	for _, conn := range parent2.Connections {
		parent2Conns[conn.Innovation] = conn
	}

	// creating offspring genome
	offspring := &Genom{
		IH:               parent1.IH,
		numInputs:        parent1.numInputs,
		numOutputs:       parent1.numOutputs,
		totalNodes:       parent1.totalNodes,
		ConnCreationRate: parent1.ConnCreationRate,
	}

	// mapping nodes by their IDs to add new connections easier
	// offspring inherits nodes from fitter parent
	nodeMap := make(map[int]*Node)
	for _, node := range parent1.Nodes {
		newNode := &Node{ID: node.ID, Type: node.Type}
		offspring.Nodes = append(offspring.Nodes, newNode)
		nodeMap[node.ID] = newNode
	}

	// offspring randomly inherits connections from either parent
	// aligning connections by their innovation numbers
	for _, conn1 := range parent1.Connections {
		var chosenConn Connection
		if conn2, exist := parent2Conns[conn1.Innovation]; exist {
			if rand.Intn(2) == 0 {
				chosenConn = conn1
			} else {
				chosenConn = conn2
			}
			// excess (unmatched) connections are inherited from fitter parent
		} else {
			chosenConn = conn1
		}
		// adding inherited connections to the offspring's genome
		inNode := nodeMap[chosenConn.InNode.ID]
		outNode := nodeMap[chosenConn.OutNode.ID]
		offspring.addConnetion(inNode, outNode, chosenConn.Weight, chosenConn.Enabled)
	}

	return offspring
}

func (genom *Genom) showConnections() {
	// helper function
	// shows all genome's connections
	for _, conn := range genom.Connections {
		fmt.Println("\nConnection inno:", conn.Innovation)
		fmt.Println("InNode ID:", conn.InNode.ID, "InNode type:", conn.InNode.Type)
		fmt.Println("OutNode ID:", conn.OutNode.ID, "OutNode type:", conn.OutNode.Type)
		fmt.Println("Connection weight:", conn.Weight)
		fmt.Println("Connection enabled:", conn.Enabled)
	}
}

func (genom *Genom) showNodes() {
	// helper function
	// shows all genome's nodes
	for _, node := range genom.Nodes {
		fmt.Println("\n Node id:", node.ID, "Node type:", node.Type)
		fmt.Println("Node incoming connections:")
		for _, conn := range node.IncomingConns {
			fmt.Println("Innovation:", conn.Innovation, "\nNode ID:",
				conn.InNode.ID, "type:", conn.InNode.Type, "–>",
				"Node ID:", node.ID, "type:", node.Type)
		}
	}
}

func main() {
	// testing

	globalInno := InnovationHistory{
		History: map[InnovationKey]int{},
	}

	//n0 := Node{ID: 0, Type: Input}
	//n1 := Node{ID: 1, Type: Output}
	//n2 := Node{ID: 2, Type: Output}

	//conn := Connection{
	//	InNode:  n0,
	//	OutNode: n1,
	//	Weight:  1.0,
	//	Enabled: true,
	//}

	network1 := Genom{numInputs: 1, numOutputs: 1,
		IH: &globalInno, ConnCreationRate: 1.0, Fitness: 1.0}
	network2 := Genom{numInputs: 1, numOutputs: 1,
		IH: &globalInno, ConnCreationRate: 1.0, Fitness: 2.0}
	network1.createNetwork()
	network2.createNetwork()
	network2.mutateAddNode()

	fmt.Println("\n Network 1")
	network1.showConnections()

	network1.showNodes()
	fmt.Println("\n Network 2")
	network2.showConnections()
	network2.showNodes()

	fmt.Println("\n Offspring:")
	offspring := crossover(&network1, &network2)
	offspring.showConnections()
	network2.showNodes()
}
