package data

import (
	"fmt"
	"math/rand"
	"time"
)

type Node struct { //reprezentuje pojedynczy neouron
	Number        int          //"ID" neuornu
	Layer         int          //numer warstwy,w której się znajduej
	InConnections []Connection //jakie połączenie wchodzi do neuronu
	IsOutput      bool
}

type Connection struct { //reprezentuje połączenie między dwoma nueonami
	In_node  Node    //neuron wejsciowy
	Out_node Node    //neuorn wyjsciowy
	Weight   float64 //waga
	Inno     int     // numer innowacji ?
	Enabled  bool    // czy połączenie jest aktywne
}

type Connectionh struct { //historia wszystkich połączen
	Inputs         int          //liczba wejsc
	Outputs        int          //liczba wyjsc
	AllConnections []Connection //wszystkie znane połączenia
	Global_inno    int          //licznik innowacji
}

func (cH *Connectionh) Exists(n1, n2 *Node) *Connection { //sprawdza czy polaczenie pomiedzy dwoma neuronami juz istnieje
	for _, c := range cH.AllConnections {
		if c.In_node.Number == n1.Number && c.Out_node.Number == n2.Number {
			return &c //zwraca chyba id polaczenia
		}
	}
	return nil //zwraca nil jesli nie ma takiego polaczenia

}

// define Genom class
type Genom struct {
	Ch            Connectionh  //bierzemy obiekt Connectionh - historia polaczen neuronow (lokalna kopia)
	Inputs        int          // ilosc neornow wejscjowych
	Outputs       int          // ilosc neronow wyjsciowych
	Input_Layer   int          //numer warstwy input
	Output_Layer  int          //numer warstwy output
	Total_Nodes   int          //laczna liczba neuronow (poczatkowo 0) - pozwala uniknac dodanie tego samego "ID" różnym neuronom
	Creation_Rate float64      //szansa na dodanie nowego połączenia pomiędzy losowymi neuronami podczas tworzenia sieci
	Nodes         []Node       //lista wezlow
	Connections   []Connection //lista polaczen aktywnych
	Fitness       float64      //wartosc naszego fitness score
	Create        bool         //potrzebujemy do tego czy uruchomic CreateNetwork(), czasem chcemy np. tylko ogolny szkielet genomu, np
	// krzyzyujac genomy , nowy genom ma miec pewne cechy rodzica, wiec automatyczne uruchomienie CreateNetwork() nadpisze te cechy - bez sensu
	// i pozniej na podsatawie danych z connectionh mozemy stworzyc genom
}

// tworzenie sieci
func (cH *Genom) CreateNetwork() {
	fmt.Println("--- Starting CreateNetwork ---")
	for i := 0; i < cH.Inputs; i++ {
		node := Node{
			Number: cH.Total_Nodes,
			Layer:  cH.Input_Layer,
		}
		fmt.Printf("Creating input node: %+v\n", node)
		cH.Nodes = append(cH.Nodes, node)
		cH.Total_Nodes++
	}
	for i := 0; i < cH.Outputs; i++ {
		outputNode := Node{
			Number:   cH.Total_Nodes,
			Layer:    cH.Output_Layer,
			IsOutput: true,
		}
		fmt.Printf("Stworzyłem output node: %+v\n", outputNode)
		cH.Nodes = append(cH.Nodes, outputNode)
		cH.Total_Nodes++
	}
	fmt.Println("--- Finished creating input/output nodes ---")
	fmt.Println("All nodes:")
	for _, node := range cH.Nodes {
		fmt.Printf("Node: %+v\n", node)
	}
	rand.Seed(time.Now().UnixNano())
	totalConnections := cH.Outputs * cH.Inputs
	for i := 0; i < totalConnections; i++ {
		if rand.Float64() < cH.Creation_Rate {
			fmt.Printf("Attempting to add connection %d\n", i)
			cH.AddConnection()
		}
	}
	fmt.Println("--- Finished creating connections ---")
	fmt.Println("All connections:")
	for _, conn := range cH.Ch.AllConnections {
		fmt.Printf("Connection: %+v\n", conn)
	}
	cH.Update_Output_Layer()
	fmt.Println("--- Finished CreateNetwork ---")
}

func (g *Genom) Update_Output_Layer() { //updateowanie numeru warstwy output
	StonestStoner := 0
	for _, node := range g.Nodes {
		if node.Layer > StonestStoner {
			StonestStoner = node.Layer
		}
	}
	g.Output_Layer = StonestStoner + 1
}

// funkcja forward
func (g *Genom) Forward(inputs []float64) []float64 {
	fmt.Println("=== START FORWARD ===")
	fmt.Printf("Wejścia: %v\n", inputs)
	nodeValues := make(map[int]float64)
	for i := 0; i < g.Inputs; i++ {
		nodeValues[g.Nodes[i].Number] = inputs[i]
	}
	for _, conn := range g.Ch.AllConnections {
		if !conn.Enabled {
			continue
		}
		inVal := nodeValues[conn.In_node.Number]
		nodeValues[conn.Out_node.Number] += inVal * conn.Weight
		fmt.Printf("Połączenie: %+v => Przekazuje: %.3f * %.3f = %.3f\n",
			conn, inVal, conn.Weight, inVal*conn.Weight)
	}
	fmt.Println("Zawartość nodeValues:")
	for k, v := range nodeValues {
		fmt.Printf("Node %d = %.4f\n", k, v)
	}
	outputs := []float64{}
	for _, node := range g.Nodes {
		if node.IsOutput {
			val, exists := nodeValues[node.Number]
			if exists {
				outputs = append(outputs, val)
			} else {
				outputs = append(outputs, 0)
				fmt.Printf("⚠️ Output node %d nie ma wartości — ustawiamy 0\n", node.Number)
			}
		}
	}
	fmt.Println("=== KONIEC FORWARD ===")
	fmt.Printf("Outputs: %v\n", outputs)
	return outputs
}

// funkcja aktywacji relu
func relu(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0
}

// funkcja mierzenia fitness
func (g *Genom) EvaluateFitness(score int, foodEaten int, enemiesKilled int, timeSurvived int, hp float64) float64 {
	fitness := float64(score) + float64(foodEaten)*10 + float64(enemiesKilled)*20 + hp*5 - float64(timeSurvived)*0.1
	g.Fitness = fitness
	return fitness
}

//MUTACJE
//mutacja z tworzeniem nowych połączeń

func (cH *Genom) AddConnection() {
	if len(cH.Nodes) == 0 {
		return
	}

	// Wybieramy dwa różne nody: n1 (źródło), n2 (cel)
	n1 := cH.Nodes[rand.Intn(len(cH.Nodes))]
	n2 := cH.Nodes[rand.Intn(len(cH.Nodes))]

	// Unikamy self-loop i połączeń wstecznych (ważne dla feed-forward sieci)
	if n1.Number == n2.Number || n1.Layer >= n2.Layer {
		return
	}

	// Sprawdź, czy takie połączenie już istnieje
	if cH.Ch.Exists(&n1, &n2) != nil {
		return // Połączenie już istnieje, nie dodajemy duplikatu
	}

	// Tworzymy połączenie
	newConn := Connection{
		In_node:  n1,
		Out_node: n2,
		Weight:   rand.Float64()*2 - 1, // losowa waga [-1, 1]
		Inno:     cH.Ch.Global_inno,    // unikalny numer innowacji
		Enabled:  true,
	}

	// Zwiększamy licznik innowacji i dodajemy połączenie do historii
	cH.Ch.Global_inno++
	cH.Ch.AllConnections = append(cH.Ch.AllConnections, newConn)
	cH.Connections = append(cH.Connections, newConn)
	fmt.Printf("[AddConnection] Nowe połączenie: z %d do %d, waga: %.2f, enabled: %v\n",
		newConn.In_node.Number, newConn.Out_node.Number, newConn.Weight, newConn.Enabled)
	// Dodajemy to połączenie do nodu docelowego
	for i := range cH.Nodes {
		if cH.Nodes[i].Number == n2.Number {
			cH.Nodes[i].InConnections = append(cH.Nodes[i].InConnections, newConn)
			break
		}
	}
}

func (g *Genom) GetNodeByNumber(num int) *Node {
	for i := range g.Nodes {
		if g.Nodes[i].Number == num {
			return &g.Nodes[i]
		}
	}
	return nil
}

// define Genom's function
func (cH *Genom) Exists(nn int) bool { // taking Connection's innovation number
	for _, c := range cH.Connections { // return true if Connection exists, False otherwise

		if c.Inno == nn {
			return true
		}
	}
	return false
}

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

// wersja paleozoik (muszę dopracować przesuwanie warstw)
func (cH *Genom) AddNodeMutation() {
	if len(cH.Connections) == 0 {
		fmt.Println("Sieć nie ma żadnych połączeń.")
		return
	}
	rand.Seed(time.Now().UnixNano())
	// bierzemy istniejące połączenie wraz z nodami
	conn := &cH.Connections[rand.Intn(len(cH.Connections))]
	n1 := &conn.In_node
	n2 := &conn.Out_node
	conn.Enabled = false // wyłączamy stare połączenie

	// nowe połączenie powstanie pomiędzy n1 i n2, więc trzeba odpowiednio zwiększyć numery
	// warstw wszystkich nodów należących do warstw za n1
	for i := range cH.Nodes {
		if cH.Nodes[i].Layer > n1.Layer {
			cH.Nodes[i].Layer++
		}
	}

	cH.Total_Nodes++ // będziemy dodawać nowy node, więc trzeba to zwiększyć

	// Tworzymy nowy node
	newNode := Node{ // nasz nowy node
		Number: cH.Total_Nodes - 1, // bo indeksujemy od zera
		Layer:  n1.Layer + 1,       // bo znajduje się za n1
	}
	cH.Nodes = append(cH.Nodes, newNode) // dodajemy nowy nod do genomu

	// tworzymy połączenie między n1 a nowym nodem
	newConn1 := Connection{
		In_node:  *n1,
		Out_node: newNode,
		Weight:   rand.Float64()*2.0 - 1.0,
		Enabled:  true,
	}
	// sprawdzamy, czy takie połączenie już istniało
	histConn1 := cH.Ch.Exists(n1, &newNode)
	if histConn1 != nil {
		newConn1.Inno = histConn1.Inno
	} else {
		newConn1.Inno = cH.Ch.Global_inno
		cH.Ch.Global_inno++
		cH.Ch.AllConnections = append(cH.Ch.AllConnections, newConn1)
	}

	// tworzymy połączenie nowym nodem a n2
	newConn2 := Connection{
		In_node:  newNode,
		Out_node: *n2,
		Weight:   rand.Float64()*2.0 - 1.0,
		Enabled:  true,
	}
	// sprawdzamy, czy takie połączenie już istniało
	histConn2 := cH.Ch.Exists(&newNode, n2)
	if histConn2 != nil {
		newConn2.Inno = histConn2.Inno
	} else {
		newConn2.Inno = cH.Ch.Global_inno
		cH.Ch.Global_inno++
		cH.Ch.AllConnections = append(cH.Ch.AllConnections, newConn2)
	}
	// dodajemy nowe połączenia do nodów
	n2.InConnections = append(n2.InConnections, newConn2)
	newNode.InConnections = append(newNode.InConnections, newConn1)

	// dodajemy nowe połączenia do genomu
	cH.Connections = append(cH.Connections, newConn1, newConn2)
	// warstwy nam się przesunęły, więc ostatnia warstwa zwiększa nam się o 1
	cH.Update_Output_Layer()

}

// prawie działa, jedynie muszę dopracować dziedziczenie warstw w nodach
func crossover(parent1 *Genom, parent1FitScore int,
	parent2 *Genom, parent2FitScore int) *Genom {
	// upewnienie się, że parent1 będzie miał większy fitness
	if parent2FitScore > parent1FitScore {
		tmp := parent1
		parent1 = parent2
		parent2 = tmp
	}
	// tworzenie dzieciaka
	offspring := Genom{}
	offspring.Creation_Rate = parent1.Creation_Rate

	// dzieciak dziedziczy nody po rodzicu z większym fitnessem
	for _, node := range parent1.Nodes {
		offspringNode := Node{
			Number: node.Number,
			Layer:  node.Layer,
		}
		offspring.Nodes = append(offspring.Nodes, offspringNode)
	}
	offspring.Inputs = parent1.Inputs
	offspring.Outputs = parent1.Outputs
	offspring.Input_Layer = parent1.Input_Layer
	offspring.Output_Layer = parent1.Output_Layer

	// dopasowujemy połączenia między nodami za pomocą Inno
	for _, conn1 := range parent1.Connections {
		for _, conn2 := range parent2.Connections {
			if conn1.Inno == conn2.Inno {
				rand.Seed(time.Now().UnixNano())
				parentNum := rand.Intn(2) // dzieciak losowo odziedziczy połączenie po
				if parentNum == 0 {       // którymś z rodziców
					offspring.Connections = append(offspring.Connections, conn1)
				} else { // ale warstwy dalej dziedziczy po bardziej fit rodzicu
					conn2.In_node.Layer = conn1.In_node.Layer
					conn2.Out_node.Layer = conn1.Out_node.Layer
					offspring.Connections = append(offspring.Connections, conn2)
				}
				// jeśli u mniej fit rodzica nie będzie połączenia z tym samym Inno,
				// połączenie jest dziedziczone po bardziej fit rodzicu
			} else {
				offspring.Connections = append(offspring.Connections, conn1)
			}
		}
	}
	// dodawanie połączeń do nodów dzieciaka
	for _, conn := range offspring.Connections {
		conn.Out_node.InConnections = append(conn.Out_node.InConnections, conn)
	}
	return &offspring
}
