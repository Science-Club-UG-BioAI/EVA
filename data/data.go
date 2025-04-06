package data

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2/inpututil"
)


type Node struct { //reprezentuje pojedynczy neouron
	Number        int          //"ID" neuornu
	Layer         int          //numer warstwy,w której się znajduej
	InConnections []Connection //jakie połączenie wchodzi do neuronu
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
	Create        bool         //potrzebujemy do tego czy uruchomic CreateNetwork(), czasem chcemy np. tylko ogolny szkielet genomu, np
	// krzyzyujac genomy , nowy genom ma miec pewne cechy rodzica, wiec automatyczne uruchomienie CreateNetwork() nadpisze te cechy - bez sensu
	// i pozniej na podsatawie danych z connectionh mozemy stworzyc genom
}

func (cH *Genom) CreateNetwork() { //tworzy genom dla kazdego osobnika z naszej populacji
	// dodajemy tutaj wezly wejsciowe
	for i := 0; i < cH.Inputs; i++ { // to jest petla while (wyrazona za pomoca for bo nie ma while w golangu)
		cH.Nodes = append(cH.Nodes, Node{ //przypisuje kazdemu neuronowi "id" i warstwe

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
		rand.Seed(time.Now().UnixNano()) //to potrzebne do losowisci w nastepnej petli -> ziarno generatora zalezne od czasu (w nanosekudnach)
		// bez tego sekwencja losowych liczb - taka sama za kazdym razem
		for i := 0; i < cH.Outputs*cH.Inputs; i++ { //tworzenie losowych połączen miedzy neuronami zaleznie od creation_rate
			if rand.Float64() < cH.Creation_Rate {

				cH.AddConnection()
			}
		}
	}
	cH.Update_Output_Layer() //uaktulnia numer warstwy
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


func (c *Connection) copy() Connection { //tworzy kopie polaczenia - potrzebne aby, nie nadpisac historii polaczen

	return Connection{
		In_node:  c.In_node,
		Out_node: c.Out_node,
		Weight:   c.Weight,
		Inno:     c.Inno,
		Enabled:  c.Enabled,
	}
}

//funckja forward
func (g *Genom) Forward(inputs []float64) []float64 {
	nodeValues := make(map[int]float64)

	for i := 0; i < g.Inputs; i++ {
		nodeValues[g.Nodes[i].Number] = inputs[i]
	}
	sortedNodes := make([]Node, len(g.Nodes))
	copy(sortedNodes, g.Nodes)
	sort.Slice(sortedNodes, func(i, j int) bool {
		return sortedNodes[i].Layer < sortedNodes[j].Layer
	})
	for _, node := range sortedNodes {
		if node.Layer == g.Input_Layer {
			continue
		}
		sum := 0.0
		for _,conn := range node.InConnections {
			if conn.Enabled {
				sum += nodeValues[conn.In_node.Number] * conn.Weight
			}
		}
		nodeValues[node.Number] = relu(sum)
	}
	outputs := []float64{}
	for _, node := range g.Nodes {
		if node.Layer == g.Output_Layer {
			outputs = append(outputs, nodeValues[node.Number])
		}
	}
	return outputs
}
//funkcja aktywacji relu
func relu(x float64) float64 {
	if x > 0 {
		return x
	}
	return 0 
}


//MUTACJE
//mutacja z tworzeniem nowych połączeń

func (cH *Genom) AddConnection() { //odpowiada za tworzenie NOWEGO połączenia między neuronami
	rand.Seed(time.Now().UnixNano())         //losujemy wybor dwoch neuronow
	n1 := cH.Nodes[rand.Intn(len(cH.Nodes))] //losujemy indeksy dla tablicy Nodes (wyciagamy randomowe neurony)
	n2 := cH.Nodes[rand.Intn(len(cH.Nodes))]

	for n1.Layer == cH.Output_Layer { //petla while (for sprawdza za kazdym razem dany warunek, jesli prawidzwy->wykonuje funkcje)
		n1 = cH.Nodes[rand.Intn(len(cH.Nodes))] // sprawdzamy czy peirwszy wyvrany neuron nie jest czasem w ooutput warstwie ->
	} //bo nie moze byc!! nie ma nic za to warstwa, wiec z jakim neuronem ma zrobic polaczenie


	for n2.Layer == cH.Input_Layer || n2.Layer <= n1.Layer { // checking if Node2 belong to input layer and if Nodes are one after another
		n2 = cH.Nodes[rand.Intn(len(cH.Nodes))] // taking new random nNode if so
	}


	// c to tutaj nasz wskaznik wskazujacy polaczenie miedzy losowymi neuronami, moze byc pusty->polecznie nie istnieje
	c := cH.Ch.Exists(&n1, &n2)  // * - wyciaga wartosc ze wskaznika; & - wklada wartosc do wskaznika
	x := Connection{In_node: n1, //tu tworzymy nowe polaczenie miedzy neuronami nawet jesli juz istnieje
		Out_node: n2}


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
