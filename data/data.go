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

type Node struct {
	ID            int          // unikalne ID naszego noda
	Type          NodeType     // jeden z trzech typów
	IncomingConns []Connection // lista połączeń wchodzących DO noda
}

type Connection struct {
	InNode     *Node
	OutNode    *Node
	Weight     float64
	Innovation int
	Enabled    bool
}

type InnovationKey struct {
	// potrzebny do InnovationHistory, który jest mapą (coś jak słownik w pythonie)
	// rozpoznaje połączenia poprzez ID tworzących go nodów
	inNodeID  int
	outNodeID int
}

type InnovationHistory struct {
	// mapa, która daje każdemu połączeniu unikalny numer innowacji
	History map[InnovationKey]int
	Counter int
}

func (ih *InnovationHistory) getInnovation(inNode, outNode *Node) int {
	// wypluwa innowację połączenia, a jeśli połączenie nie istnieje,
	// nadaje nową innowację i zwiększa globalny licznik innowacji
	key := InnovationKey{inNodeID: inNode.ID, outNodeID: outNode.ID}
	if inno, exist := ih.History[key]; exist {
		return inno
	}
	ih.Counter++
	ih.History[key] = ih.Counter
	return ih.History[key]
}

func (genom *Genom) addConnetion(node1, node2 *Node, weight float64, enabled bool) {
	// funkcja pomocnicza,
	// dodaje połączenie do genomu, od razu dodaje połączenie do noda
	inno := genom.IH.getInnovation(node1, node2)
	newConn := Connection{
		InNode:     node1,
		OutNode:    node2,
		Weight:     weight,
		Innovation: inno,
		Enabled:    enabled,
	}
	genom.Connections = append(genom.Connections, newConn)
	node2.IncomingConns = append(node2.IncomingConns, newConn)
}

func (genom *Genom) connectionExist(inNode, outNode *Node) bool {
	// funkcja pomocnicza
	// sprawdza, czy połączenie istnieje
	for _, conn := range genom.Connections {
		if conn.InNode.ID == inNode.ID && conn.OutNode.ID == outNode.ID {
			return true
		}
	}
	return false
}

func (genom *Genom) randomNodes() (*Node, *Node) {
	// funkcja pomocnicza
	// wypluwa dwa nody z genomu, które nadają się do stworzenia połączenia n1 ––> n2
	rand.Seed(time.Now().UnixNano())
	n1 := genom.Nodes[rand.Intn(len(genom.Nodes))]
	n2 := genom.Nodes[rand.Intn(len(genom.Nodes))]

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
	numInputs        int                // liczba nodów Input
	numOutputs       int                // liczba nodów Output
	totalNodes       int                // łączna liczba nodów
	Nodes            []*Node            // lista nodów w genomie
	Connections      []Connection       // lista połączeń w genomie
	ConnCreationRate float64            // szansa na stworzenie połączenia przy generowaniu nowego genomu
	IH               *InnovationHistory // globalna historia innowacji
}

func (genom *Genom) createNetwork() {
	// tworzy nową sieć z losowymi połączeniami

	//dodanie warswty input
	for i := 0; i < genom.numInputs; i++ {
		genom.Nodes = append(genom.Nodes, &Node{ID: genom.totalNodes, Type: Input})
		genom.totalNodes++
	}
	// dodanie warstwy output
	for i := 0; i < genom.numOutputs; i++ {
		genom.Nodes = append(genom.Nodes, &Node{ID: genom.totalNodes, Type: Output})
		genom.totalNodes++
	}
	// dodanie losowych połączeń
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
	// funkcja mutacji wag w genomie
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
	// funkcja mutacji dodająca nowe połączenie w genomie
	rand.Seed(time.Now().UnixNano())
	n1, n2 := genom.randomNodes()
	if !genom.connectionExist(n1, n2) {
		weight := rand.Float64()
		genom.addConnetion(n1, n2, weight, true)
	}
}

func (genom *Genom) mutateAddNode() {
	// funkcja mutacji dodająca nowy nod
	if len(genom.Connections) == 0 {
		return
	}
	// bierze losowe połączenie w genomie i je wyłącza
	rand.Seed(time.Now().UnixNano())
	conn := &genom.Connections[rand.Intn(len(genom.Connections))]
	n1 := conn.InNode
	n2 := conn.OutNode
	conn.Enabled = false

	// tworzy nowy nowy i wsadza je na miejsce starego połączenia
	genom.totalNodes++
	newNode := Node{ID: genom.totalNodes - 1, Type: Hidden}
	// tworzą się dwa nowe połączenia. ich wagi odpowiadają wadze starego połączenia,
	// aby mutacja nie była zbyt drastyczna
	genom.addConnetion(n1, &newNode, 1.0, true)
	genom.addConnetion(&newNode, n2, conn.Weight, true)
	genom.Nodes = append(genom.Nodes, &newNode)
}

func crossover(parent1, parent2 *Genom, parent1Fit, parent2Fit int) *Genom {
	// tworzenie dziecka z genomów dwóch rodziców
	// topologia sieci jest dziedziczona po rodzicu z większym fitnessScore

	// parent1 ma mieć większy fitnessScore
	if parent2Fit > parent1Fit {
		tmp := parent1
		parent1 = parent2
		parent2 = tmp
	}
	// mapowanie połączeń poprzez numer innowacji
	parent2Conns := make(map[int]Connection)
	for _, conn := range parent2.Connections {
		parent2Conns[conn.Innovation] = conn
	}

	// tworzenie dzieciaka
	offspring := &Genom{
		IH:               parent1.IH,
		numInputs:        parent1.numInputs,
		numOutputs:       parent1.numOutputs,
		totalNodes:       parent1.totalNodes,
		ConnCreationRate: parent1.ConnCreationRate,
	}

	// mapowanie nodów po ID
	// dzieciak dziedziczy nody po rodzicu z większym fitnessScore
	nodeMap := make(map[int]*Node)
	for _, node := range parent1.Nodes {
		newNode := &Node{ID: node.ID, Type: node.Type}
		offspring.Nodes = append(offspring.Nodes, newNode)
		nodeMap[node.ID] = newNode
	}

	// dzieciak losowo dziedziczy połączenia po którymś z rodziców
	for _, conn1 := range parent1.Connections {
		var chosenConn Connection
		if conn2, exist := parent2Conns[conn1.Innovation]; exist {
			if rand.Intn(2) == 0 {
				chosenConn = conn1
			} else {
				chosenConn = conn2
			}
		} else {
			chosenConn = conn1 // jeśli w parent2 nie ma odpowiednika parent1, to
		} // dzieciak dziedziczy połączenie po parent1

		inNode := nodeMap[chosenConn.InNode.ID]
		outNode := nodeMap[chosenConn.OutNode.ID]
		offspring.addConnetion(inNode, outNode, chosenConn.Weight, chosenConn.Enabled)
	}

	return offspring
}

func (genom *Genom) showConnections() {
	// funkcja pomocnicza
	// pokazuje połączenia w genomie
	for _, conn := range genom.Connections {
		fmt.Println("\nConnection inno:", conn.Innovation)
		fmt.Println("InNode ID:", conn.InNode.ID, "InNode type:", conn.InNode.Type)
		fmt.Println("OutNode ID:", conn.OutNode.ID, "OutNode type:", conn.OutNode.Type)
		fmt.Println("Connection weight:", conn.Weight)
		fmt.Println("Connection enabled:", conn.Enabled)
	}
}

func (genom *Genom) showNodes() {
	// funkcja pomocnicza
	// pokazuje nody w genomie
	for _, node := range genom.Nodes {
		fmt.Println("\n Node id:", node.ID, "Node type:", node.Type)
		fmt.Println("Node incoming connections:")
		for _, conn := range node.IncomingConns {
			fmt.Println("Inno:", conn.Innovation, "From node:",
				conn.InNode.ID, "type:", conn.InNode.Type)
		}
	}
}

func main() {
	// testowanie

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

	network1 := Genom{numInputs: 1, numOutputs: 1, IH: &globalInno, ConnCreationRate: 1.0}
	network2 := Genom{numInputs: 1, numOutputs: 1, IH: &globalInno, ConnCreationRate: 1.0}
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
	offspring := crossover(&network1, &network2, 1, 2)
	offspring.showConnections()
	network2.showNodes()
}

