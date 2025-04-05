package data

import (
	"fmt"
	"math/rand"
	"time"
)

// define the Node class
type Node struct {
	Number int // Node's number
	Layer  int // Node's layer's number
	InConnections []Connection // Connections entering this Node
}

// define the Connection class
type Connection struct {
	In_node  Node // input Node
	Out_node Node // output Node
	Weight   float64 // weight of Connection
	Inno     int // innovation number
	Enabled  bool // if active
}

// define Connection's function
func (c Connection) ShowConn() {
	fmt.Printf("Connection: In_node: %d, Out_node: %d, Weight: %f, Inno: %d, Enabled: %v\n",
		c.In_node, c.Out_node, c.Weight, c.Inno, c.Enabled)
} // printing Connection info

// define Connectionh class (Connection's storage)
type Connectionh struct {
	Inputs         int // input Nodes' amount
	Outputs        int // output Nodes' amount
	AllConnections []Connection 
	Global_inno    int // global innovation number
}

// define Connectionh's function (does Connection exist?)
func (cH *Connectionh) Exists(n1, n2 *Node) *Connection { // takes Nodes as arguments
	for _, c := range cH.AllConnections { // iteration trough indexes and Connections
		if c.In_node.Number == n1.Number && c.Out_node.Number == n2.Number { // checking nodes' numbers
			return &c // Connection pointer
		}
	}
	return nil // empty pointer
}

// define Genom class
type Genom struct {
	Ch            Connectionh // Connectionh object
	Inputs        int         // input Nodes' amount
	Outputs       int         // output Nodes' amount
	Input_Layer   int         // input layer's number
	Output_Layer  int         // output layer's number
	Total_Nodes   int         // Nodes' amount (0 at the beginning)
	Creation_Rate float64 // needed to Connections' draw
	Nodes         []Node       // Nodes' list
	Connections   []Connection // Connections' list
	Create        bool         // do we want to activate CreateNetwork? sometimes we just wanna know general genom skeleton
	// crossing genomes for example so we don't wanna overide parent characteristic
}

// define Genom's function (creating new Genom - 2 layers at the beginning)
func (cH *Genom) CreateNetwork() {
	// adding input Nodes
	for i := 0; i < cH.Inputs; i++ { // adding Nodes for given Inputs number
		cH.Nodes = append(cH.Nodes, Node{
			Number: cH.Total_Nodes,
			Layer:  cH.Input_Layer,
		})
		cH.Total_Nodes++ // increasing Total_Nodes by 1 (in order to add another Node in loop)
	}

	// adding output Nodes
	for i := 0; i < cH.Outputs; i++ { // works like above
		cH.Nodes = append(cH.Nodes, Node{
			Number: cH.Total_Nodes,
			Layer:  cH.Output_Layer,
		})
		cH.Total_Nodes++
		rand.Seed(time.Now().UnixNano()) // ensures randomness
		for i := 0; i < cH.Outputs*cH.Inputs; i++ { // loop for everyone with everyone Connections
			if rand.Float64() < cH.Creation_Rate { // do we create a Connection (randomly)?
				cH.AddConnection()
			}
		}
	}
}


func (c *Connection) copy() Connection {
	return Connection{
		In_node:  c.In_node,
		Out_node: c.Out_node,
		Weight:   c.Weight,
		Inno:     c.Inno,
		Enabled:  c.Enabled,
	}
}


func (cH *Genom) AddConnection() {
	rand.Seed(time.Now().UnixNano())
	n1 := cH.Nodes[rand.Intn(len(cH.Nodes))] // drawing index (taking random Node)
	n2 := cH.Nodes[rand.Intn(len(cH.Nodes))] // as above

	for n1.Layer == cH.Output_Layer { // checking if Node1 belong to output layer
		n1 = cH.Nodes[rand.Intn(len(cH.Nodes))] // taking new random Node if so					
	}

	for n2.Layer == cH.Input_Layer || n2.Layer <= n1.Layer { // checking if Node2 belong to input layer and if Nodes are one after another
		n2 = cH.Nodes[rand.Intn(len(cH.Nodes))] // taking new random nNode if so
	}

	c := cH.Ch.Exists(&n1, &n2) // taking Connection pointer
	x := Connection{In_node: n1,
		Out_node: n2} // creating new Connection (not adding to Connectionh yet)

	if c != nil { // if Connection's pointer exists -> Connection was made before (in any Genom and it may be gone now) -> Connection x gets its innovation number
		x.Inno = c.Inno
		if !cH.Exists(x.Inno) { // does Connection exists in this Genom?
			cH.Connections = append(cH.Connections, x) // if no -> adding Connection to Connections' list in Genom
			n2.InConnections = append(n2.InConnections, x) // if no -> adding Connection to output Node's Connections
		}
		
	} 	else  { // Connection's pointer doesn't exist -> Connection was never made
		x.Inno = cH.Ch.Global_inno // new innovation number for Connection
		cH.Ch.Global_inno += 1 // increasing global innovation number by 1 (prepare for next Connection)
		cH.Connections = append(cH.Connections, x) // adding Connection to Genom
		cH.Ch.AllConnections = append(cH.Ch.AllConnections, x.copy()) // adding Connection's copy to Connectionh (we have to distinguish Genom's Connections and Population's Connections)
		n2.InConnections = append(n2.InConnections, x) // adding Connection to output Node's Connection
	}}

// define Genom's function
func (cH *Genom) Exists(nn int) bool { // taking Connection's innovation number 
	for _, c := cH.Connections { // return true if Connection exists, False otherwise
		if c.Inno == nn {
			return true
		}
	return false
	}
}

//mutacje
// mutacja z roznica wag
func (cH *Genom) Mutate_weight() {
	rand.Seed(time.Now().UnixNano())

	for i, conn := range cH.Connections { //idziemy po wszystkich połączeniach
		if rand.Float64() < 0.8 { //80% na małą zmianę, 20% na mega duza zmiane
			delta := (rand.Float64() * 0.4) - 0.2 //losuje zmiane ktora dodamy do wagi z (-0.2,0.2)
			conn.Weight += delta                  //dodajemy tą zmianę
		} else {
			conn.Weight = (rand.Float64() * 2.0) - 1.0 //przypisuje nowa wage z zakresu (-1.0,1.0)
		}
		// aktualizujemy nasza zmiane
		cH.Connections[i] = conn // wprowadza aktualizację
	}
}
// mutacje z tworzeniem nowych połączeń
func (g *Genom) AddConnectionMutation() {
	rand.Seed(time.Now().UnixNano())

	var n1, n2 Node
	valid := false

	for !valid {
		n1 = g.Nodes[rand.Intn(len(g.Nodes))]
		n2 = g.Nodes[rand.Intn(len(g.Nodes))]

		if n1.Layer == g.Output_Layer || n2.Layer == g.Input_Layer || n1.Layer >= n2.Layer {
			continue
		}

		exists := false
		for _, c := range g.Connections {
			if c.In_node.Number == n1.Number && c.Out_node.Number == n2.Number {
				exists = true
				break
			}
		}

		if !exists {
			valid = true
		}
	}

	histConn := g.Ch.Exists(&n1, &n2)

	newConn := Connection{
		In_node:  n1,
		Out_node: n2,
		Weight:   rand.Float64()*2.0 - 1.0, // waga [-1, 1]
		Enabled:  true,
	}

	if histConn != nil {
		newConn.Inno = histConn.Inno
	} else {
		newConn.Inno = g.Ch.Global_inno
		g.Ch.Global_inno++
		g.Ch.AllConnections = append(g.Ch.AllConnections, newConn.copy())
	}

	g.Connections = append(g.Connections, newConn)
	n2.InConnections = append(n2.InConnections, newConn)
}
