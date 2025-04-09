package data

import (
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// – – – – – – – – – – – – – – DATA TYPES AND STRUCTURES – – – – – – – – – – – – – – – – –

type NodeType int

const (
	Input  NodeType = iota // Input = 0
	Hidden                 // Hidden = 1
	Output                 // Output = 2
)

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

type Genom struct {
	NumInputs        int                // number of Input nodes
	NumOutputs       int                // number of Output nodes
	TotalNodes       int                // total number of nodes
	Nodes            []*Node            // list of all nodes in the genome
	Connections      []Connection       // list of all connections in the genome
	ConnCreationRate float64            // chance of adding connection while creating new network
	IH               *InnovationHistory // global innovation history
	Fitness          float64            // fitness score
}

type Species struct {
	Genoms         []*Genom // list of genoms within the species
	AverageFitness float64  // average fitness within the species
	BreedingRate   int      // how many offsprings this species is allowed to produce
}

type Population struct {
	AllSpecies        []*Species // list of all species within the population
	PopSize           int        // total numer of genomes within the population
	CurrentGeneration int        // number of current generation
	C1                float64    // constant which multiplies deltaGenes
	C2                float64    // constant which multiplies deltaWeights
	Threshold         float64    // threshold for speciating
}

type AIDecision struct { //przechowuje informacje o pojedyńczej decyzji podjętej przez AI (np. decyzja żeby iść w prawo)
	Inputs      []float64        //wartości wejściowe (np. jak daleko jest enemy)
	Outputs     []float64        //wartości wyjściowe (decyzja idź w prawo)
	Connections []ConnectionInfo //lista wszystkich aktywnych połączeń i ich wpływu - jak do tego doszło
}

type ConnectionInfo struct { //szczegół jednego połączenia z podjętej decyzji
	From   int     //ID neuronu źródłowego - informacja
	To     int     // ID neuronu docelowego - decyzja
	Weight float64 //waga
	Effect float64 //input * weight
}

// – – – – – – – – – – – – – – MAIN FUNCTIONALITY – – – – – – – – – – – – – – – –– – – – – – –

func (genom *Genom) CreateNetwork() {
	// creates new network with only input and output nodes
	// adds connections randomly

	// adding inputs nodes
	for i := 0; i < genom.NumInputs; i++ {
		genom.Nodes = append(genom.Nodes, &Node{ID: genom.TotalNodes, Type: Input})
		genom.TotalNodes++ //sprawdzic bo podwaja nodey
	}
	// adding output nodes
	for i := 0; i < genom.NumOutputs; i++ {
		genom.Nodes = append(genom.Nodes, &Node{ID: genom.TotalNodes, Type: Output})
		genom.TotalNodes++
	}
	// adding random connections between nodes
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < genom.NumInputs*genom.NumOutputs; i++ {
		if rand.Float64() < genom.ConnCreationRate {
			node1, node2 := genom.randomNodes()
			if !genom.connectionExist(node1, node2) {
				weight := rand.Float64()*2 - 1
				genom.addConnetion(node1, node2, weight, true)
			}
		}
	}
	// making sure genom has at least one connection
	genom.forceConnection()
}

func (genom *Genom) EvaluateFitness(score int, foodEaten, enemiesKilled, timeSurvived int, hp float64) float64 {
	fitness := float64(foodEaten)*10 + float64(enemiesKilled)*20 + (math.Min(hp/15, 1.0))*10 + (math.Min(float64(timeSurvived)/900.0, 1.0))*20.0
	if foodEaten == 0 && enemiesKilled == 0 && score == 0 && timeSurvived > 840 {
		fitness -= 80
	}
	if fitness < 0 {
		fitness = 0
	}

	genom.Fitness = fitness
	return fitness
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
	parent2Map := make(map[int]Connection)
	for _, conn := range parent2.Connections {
		parent2Map[conn.Innovation] = conn
	}

	// creating offspring genome
	offspring := &Genom{
		IH:               parent1.IH,
		NumInputs:        parent1.NumInputs,
		NumOutputs:       parent1.NumOutputs,
		TotalNodes:       parent1.TotalNodes,
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
		if conn2, exist := parent2Map[conn1.Innovation]; exist {
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

func (pop *Population) AddToSpecies(genom *Genom) {
	// adds genome to a compatible species
	// if no match is found, creates new species and adds the genome to it

	added := false
	for _, species := range pop.AllSpecies {
		if pop.sameSpecies(genom, species.Genoms[0]) {
			species.Genoms = append(species.Genoms, genom)
			added = true
			return
		}
	}

	if !added {
		newSpecies := Species{}
		newSpecies.Genoms = append(newSpecies.Genoms, genom)
		pop.AllSpecies = append(pop.AllSpecies, &newSpecies)
	}
}

func (genom *Genom) Forward(inputs []float64) ([]float64, AIDecision) {
	fmt.Println("=== START FORWARD ===")
	fmt.Printf("Wejścia: %v\n", inputs)
	nodeValues := make(map[int]float64)
	inputIndex := 0

	for _, node := range genom.Nodes {
		if node.Type == 0 {
			if inputIndex < len(inputs) {
				nodeValues[node.ID] = inputs[inputIndex]
				inputIndex++
			} else {
				nodeValues[node.ID] = 0
			}
		}
	}
	// accumulating weighted sums per node
	incomingSums := make(map[int]float64)
	for _, conn := range genom.Connections {
		if !conn.Enabled {
			continue
		}
		inVal := nodeValues[conn.InNode.ID]
		incomingSums[conn.OutNode.ID] += inVal * conn.Weight
	}
	// applying activation functions
	for _, node := range genom.Nodes {
		if node.Type == Hidden {
			nodeValues[node.ID] = relu(incomingSums[node.ID])
		}
		if node.Type == Output {
			nodeValues[node.ID] = sigmoid(incomingSums[node.ID])
		}
	}

	outputs := []float64{}
	fmt.Println("Zawartość nodeValues:")
	for k, v := range nodeValues {
		fmt.Printf("Node %d = %.4f\n", k, v)
	}
	for _, node := range genom.Nodes {
		if node.Type == 2 {
			val, exists := nodeValues[node.ID]
			if exists {
				outputs = append(outputs, val)
			} else {
				outputs = append(outputs, 0)
				fmt.Printf("Output node %d nie ma wartości — ustawiamy 0\n", node.ID)
			}
		}
	}
	fmt.Println("=== KONIEC FORWARD ===")
	fmt.Printf("Outputs: %v\n", outputs)
	// zbiera informacje o wszystkich połączeniach
	connectionInfo := make([]ConnectionInfo, 0) //pusta lista obiektow typu ConnectionInfo
	for _, conn := range genom.Connections {    //przetwarzanie tylko aktywnych połączeń
		if conn.Enabled {
			effect := nodeValues[conn.InNode.ID] * conn.Weight      //rzeczywisty wpływ na decyzje
			connectionInfo = append(connectionInfo, ConnectionInfo{ //zapisywanie danych połączenia
				From:   conn.InNode.ID,
				To:     conn.OutNode.ID,
				Weight: conn.Weight,
				Effect: effect,
			})
		}
	}
	return outputs, AIDecision{
		Inputs:      inputs,
		Outputs:     outputs,
		Connections: connectionInfo,
	}
}

// – – – – – – – – – – – – – – – – – MUTATIONS – – – – – – – – – – – – – – – – – – – – – – –

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
	genom.TotalNodes++
	newNode := Node{ID: genom.TotalNodes - 1, Type: Hidden}
	// creates two new connections
	// weighs are chosen in such way, two new connections behave in the same way as old one
	// this way mutation is not too drastic
	genom.addConnetion(n1, &newNode, 1.0, true)
	genom.addConnetion(&newNode, n2, conn.Weight, true)
	genom.Nodes = append(genom.Nodes, &newNode)
}

func (genom *Genom) mutateToggleConnection() {
	// randomly toggles the "Enabled" state for connections
	rand.Seed(time.Now().UnixNano())

	// We randomly select one connection from the Connections list
	if len(genom.Connections) == 0 {
		return
	}

	conn := &genom.Connections[rand.Intn(len(genom.Connections))]

	// We change the "Enabled" state of the connection (if it was enabled, we disable it, and vice versa)
	conn.Enabled = !conn.Enabled
}

// – – – – – – – – – – – – – – SELECTION PROCESS – – – – – – – – – – – – – – – – – – – – – – –

func ranked(species *Species, k int) *Genom { // wybor rodzicow - typ turniejowy
	best := species.Genoms[rand.Intn(len(species.Genoms))]
	for i := 1; i < k; i++ {
		syzyf := species.Genoms[rand.Intn(len(species.Genoms))]
		if syzyf.Fitness > best.Fitness {
			best = syzyf
		}
	}
	return best
}

func GenerateNewPopulation(pop *Population) []*Genom {
	fmt.Printf("[INFO] Generating new population - number of species: %d\n", len(pop.AllSpecies))
	newGenomes := []*Genom{}
	rand.Seed(time.Now().UnixNano())

	if len(pop.AllSpecies) == 0 {
		//fmt.Println("Brak gatunków — nie można wygenerować nowej populacji.")
		return []*Genom{}
	}

	//here we calculate the average fitness for species n how many offsprings a species can have
	totalFitness := 0.0
	for _, species := range pop.AllSpecies {
		speciesTotal := 0.0
		for _, g := range species.Genoms {
			speciesTotal += g.Fitness

		}
		species.AverageFitness = speciesTotal / float64(len(species.Genoms))
		totalFitness += species.AverageFitness
	}

	for _, species := range pop.AllSpecies {
		offspringCount := int((species.AverageFitness / totalFitness) * float64(pop.PopSize))
		for i := 0; i < offspringCount && len(newGenomes) < pop.PopSize; i++ {
			if len(species.Genoms) == 0 {
				continue
			}
			parent1 := ranked(species, 3) // 3 means we choosin 3 candidates
			parent2 := ranked(species, 3)
			child := crossover(parent1, parent2)

			// Mutations in offsprings
			child.mutateWeight()
			if rand.Float64() < 0.7 {
				child.mutateAddConnection()
			}
			if rand.Float64() < 0.2 {
				child.mutateAddNode()
			}
			if rand.Float64() < 0.1 {
				child.mutateToggleConnection()
			}

			newGenomes = append(newGenomes, child)
		}
	}

	// if somehow well end up with less offsprings we will add randoms from the best pokemons
	for len(newGenomes) < pop.PopSize {
		if len(pop.AllSpecies) == 0 {
			fmt.Println("Brak dostępnych gatunków przy tworzeniu nowej generacji.")
			break
		}
		bestSpecies := pop.AllSpecies[rand.Intn(len(pop.AllSpecies))]
		if len(bestSpecies.Genoms) == 0 {
			continue
		}
		parent := bestSpecies.Genoms[rand.Intn(len(bestSpecies.Genoms))]

		//better version of copy - should work
		newGen := &Genom{
			NumInputs:        parent.NumInputs,
			NumOutputs:       parent.NumOutputs,
			ConnCreationRate: parent.ConnCreationRate,
			IH:               parent.IH,
			TotalNodes:       parent.TotalNodes,
		}

		nodeMap := make(map[int]*Node)
		for _, node := range parent.Nodes {
			newNode := &Node{
				ID:   node.ID,
				Type: node.Type,
			}
			newGen.Nodes = append(newGen.Nodes, newNode)
			nodeMap[node.ID] = newNode
		}

		for _, conn := range parent.Connections {
			newConn := Connection{
				InNode:     nodeMap[conn.InNode.ID],
				OutNode:    nodeMap[conn.OutNode.ID],
				Weight:     conn.Weight,
				Innovation: conn.Innovation,
				Enabled:    conn.Enabled,
			}
			newGen.Connections = append(newGen.Connections, newConn)
			nodeMap[conn.OutNode.ID].IncomingConns = append(nodeMap[conn.OutNode.ID].IncomingConns, newConn)
		}

		newGenomes = append(newGenomes, newGen)
	}

	fmt.Printf("[INFO] New population – number of genoms: %d\n", len(newGenomes))
	return newGenomes
}

// – – – – – – – – – – – – – – UTILITY FUNCTIONS – – – – – – – – – – – – – – – – – – – – – – –

func (ih *InnovationHistory) getInnovation(inNode, outNode *Node) int {
	// given nodes, returns innovation nubmer of the connection between them
	if ih.History == nil {
		ih.History = make(map[InnovationKey]int)
	}
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

func (genom *Genom) forceConnection() {
	// forces genome to have at least one connection
	if len(genom.Connections) == 0 {
		rand.Seed(time.Now().UnixNano())
		n1, n2 := genom.randomNodes()
		weight := rand.Float64()
		genom.addConnetion(n1, n2, weight, true)
	}
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

func deltaGenes(genom1, genom2 *Genom) float64 {
	// returns normalized number of disjoint genes (connections which don't match)
	disjointGenes := 0.0

	// number of genes in longer genome for normalization
	var longerGenome int
	if len(genom1.Connections) > len(genom2.Connections) {
		longerGenome = len(genom1.Connections)
	} else {
		longerGenome = len(genom2.Connections)
	}

	// creating maps of connections for easier comparison
	genom1Map := make(map[int]bool)
	for _, conn := range genom1.Connections {
		genom1Map[conn.Innovation] = true
	}
	genom2Map := make(map[int]bool)
	for _, conn := range genom2.Connections {
		genom2Map[conn.Innovation] = true
	}

	// comparing genomes
	for _, conn := range genom1.Connections {
		if _, exists := genom2Map[conn.Innovation]; !exists {
			disjointGenes++
		}
	}
	for _, conn := range genom2.Connections {
		if _, exists := genom1Map[conn.Innovation]; !exists {
			disjointGenes++
		}
	}
	return disjointGenes / float64(longerGenome)
}

func deltaWeights(genom1, genom2 *Genom) float64 {
	// calcutates average weight difference between connections of genomes
	differences := 0.0
	numMatches := 0.0

	genom2Map := make(map[int]Connection)
	for _, conn := range genom2.Connections {
		genom2Map[conn.Innovation] = conn
	}
	for _, conn1 := range genom1.Connections {
		if conn2, exist := genom2Map[conn1.Innovation]; exist {
			differences += (conn1.Weight - conn2.Weight)
			numMatches++
		}
	}
	return math.Abs(differences / numMatches)
}

func (pop *Population) sameSpecies(genom1, genom2 *Genom) bool {
	// checks if two genomes are within same species
	// C1, C2 and Threshold are constants, which need to be tuned experimentally
	dg := deltaGenes(genom1, genom2)
	dw := deltaWeights(genom1, genom2)
	delta := pop.C1*dg + pop.C2*dw

	return delta < pop.Threshold
}

func relu(x float64) float64 { //funkcja aktywacji relu - wywolywana w funkcji forward
	if x > 0 {
		return x
	}
	return 0
}

func sigmoid(x float64) float64 {
	// scaled sigmoid activating function, range [-1,1]
	return 2.0/(1.0+math.Exp(-4.9*x)) - 1.0
}

func SavePopulationToFile(pop *Population, generation int) error { //funkcja testowa sprawdzajaca dzialanie NEAT
	os.MkdirAll("generations", os.ModePerm)
	filename := fmt.Sprintf("generations/generation_%d.txt", generation)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for speciesIdx, species := range pop.AllSpecies {
		fmt.Fprintf(file, "=== SPECIES %d ===\n", speciesIdx)
		fmt.Fprintf(file, "Average Fitness: %.2f\n", species.AverageFitness)
		for genomIdx, genom := range species.Genoms {
			fmt.Fprintf(file, "\n--- Genom %d (Fitness: %.2f) ---\n", genomIdx, genom.Fitness)

			fmt.Fprintln(file, "Nodes:")
			for _, node := range genom.Nodes {
				fmt.Fprintf(file, "  Node ID: %d, Type: %s\n", node.ID, node.Type.String())
			}

			fmt.Fprintln(file, "Connections:")
			for _, conn := range genom.Connections {
				mutationNote := ""
				if conn.InNode.Type == Hidden || conn.OutNode.Type == Hidden {
					mutationNote = " [mutation]"
				}
				fmt.Fprintf(file,
					"  %d -> %d | Weight: %.4f | Enabled: %v%s\n",
					conn.InNode.ID, conn.OutNode.ID, conn.Weight, conn.Enabled, mutationNote)
			}
		}
		fmt.Fprintln(file)
	}

	fmt.Fprintf(file, "\n--- TOTAL GENOMES: %d ---\n", len(AllGenomesFromPopulation(pop)))
	return nil
}

func AllGenomesFromPopulation(pop *Population) []*Genom {
	var all []*Genom
	for _, species := range pop.AllSpecies {
		all = append(all, species.Genoms...)
	}
	return all
}

func AppendBestFitnessLog(generation int, population []*Genom) error {
	if len(population) == 0 {
		return nil
	}

	// Znalezienie najlepszego genomu
	bestGenom := population[0]
	for _, g := range population[1:] {
		if g.Fitness > bestGenom.Fitness {
			bestGenom = g
		}
	}

	// Obliczenie średniego fitnessu
	totalFitness := 0.0
	for _, g := range population {
		totalFitness += g.Fitness
	}
	avgFitness := totalFitness / float64(len(population))

	// Zapis do pliku CSV
	file, err := os.OpenFile("best_fitness_log.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{
		strconv.Itoa(generation),
		strconv.FormatFloat(bestGenom.Fitness, 'f', 4, 64),
		strconv.FormatFloat(avgFitness, 'f', 4, 64),
	}

	return writer.Write(record)
}

// – – – – – – – – – – – – – – – – – TESTING – – – – – – – – – – – – – – – – – – – – – – –

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

// func main() {
// 	// testing

// 	globalInno := InnovationHistory{
// 		History: map[InnovationKey]int{},
// 	}
// 	pop := Population{
// 		PopSize:           2,
// 		CurrentGeneration: 0,
// 		C1:                1.0,
// 		C2:                1.0,
// 		Threshold:         2.0,
// 	}

// 	g1 := Genom{NumInputs: 1, NumOutputs: 1, IH: &globalInno, ConnCreationRate: 1.0}
// 	g2 := Genom{NumInputs: 1, NumOutputs: 1, IH: &globalInno, ConnCreationRate: 0.0}
// 	g1.CreateNetwork()
// 	g2.CreateNetwork()
// 	pop.addToSpecies(&g1)
// 	pop.addToSpecies(&g2)

// 	fmt.Println("num species:", len(pop.AllSpecies))
// }
