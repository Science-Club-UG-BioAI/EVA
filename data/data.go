package data

import (
	"fmt"
	"math/rand"
	"time"
)

type Node struct {
	Number int
	Layer  int
	InConnections []Connection
}

type Connection struct {
	In_node  Node
	Out_node Node
	Weight   float64
	Inno     int
	Enabled  bool
}

func (c Connection) ShowConn() {
	fmt.Printf("Connection: In_node: %d, Out_node: %d, Weight: %f, Inno: %d, Enabled: %v\n",
		c.In_node, c.Out_node, c.Weight, c.Inno, c.Enabled)
}

type Connectionh struct {
	Inputs         int
	Outputs        int
	AllConnections []Connection
	Global_inno    int
}

func (cH *Connectionh) Exists(n1, n2 *Node) *Connection {
	for _, c := range cH.AllConnections {
		if c.In_node.Number == n1.Number && c.Out_node.Number == n2.Number {
			return &c
		}
	}
	return nil
}

type Genom struct {
	Ch            Connectionh //bierzemy obiekt Connectionh - historia polaczen neurowno
	Inputs        int         // ilosc neornow wejscjowych
	Outputs       int         // ilosc neronow wyjsciowych
	Input_Layer   int         //numer warstwy input
	Output_Layer  int         //numer warstwy output
	Total_Nodes   int         //laczna liczba neuronow (poczatkowo 0)
	Creation_Rate float64
	Nodes         []Node       //lista wezlow
	Connections   []Connection //lista polaczen
	Create        bool         //potrzebujemy do tego czy uruchomic CreateNetwork(), czasem chcemy np. tylko ogolny szkielet genomu, np
	// krzyzyujac genomy , nowy genom ma miec pewne cechy rodzica, wiec automatyczne uruchomienie CreateNetwork() nadpisze te cechy - bez sensu
	// i pozniej na podsatawie danych z connectionh mozemy stworzyc genom
}

func (cH *Genom) CreateNetwork() {
	// dodajemy tutaj wezly wejsciowe
	for i := 0; i < cH.Inputs; i++ { // to jest petla while (wyrazona za pomoca for bo nie ma while w golangu)
		cH.Nodes = append(cH.Nodes, Node{
			Number: cH.Total_Nodes,
			Layer:  cH.Input_Layer,
		})
		cH.Total_Nodes++ // zwiekszamy liczbe nodow o 1
	}
	// dodajemy tutaj wezly wyjsciowe
	for i := 0; i < cH.Outputs; i++ {
		cH.Nodes = append(cH.Nodes, Node{
			Number: cH.Total_Nodes,
			Layer:  cH.Output_Layer,
		})
		cH.Total_Nodes++
		rand.Seed(time.Now().UnixNano()) //to potrzebne do losowisci w nastepnej petli -> ziarno generatora zalezne od czasu (w nanosekudnach)
		// bez tego sekwencja losowych liczb - taka sama za kazdym razem
		for i := 0; i < cH.Outputs*cH.Inputs; i++ {
			if rand.Float64() < cH.Creation_Rate {
				cH.AddConnection()
			}
		}
	}
}

func (cH *Genom) AddConnection() {
	rand.Seed(time.Now().UnixNano())
	n1 := cH.Nodes[rand.Intn(len(cH.Nodes))] //losujemy indeksy dla tablicy Nodes (wyciagamy randomowe neuronoy)
	n2 := cH.Nodes[rand.Intn(len(cH.Nodes))]

	for n1.Layer == cH.Output_Layer { //petla while (for sprawdza za kazdym razem dany warunek, jesli prawidzwy->wykonuje funkcje)
		n1 = cH.Nodes[rand.Intn(len(cH.Nodes))] // sprawdzamy czy peirwszy wyvrany neuron nie jest czasem w ooutput warstwie ->												//
	} //bo nie moze byc!! nie ma nic za to warstwa, wiec z jakim neuronem ma zrobic polaczenie

	for n2.Layer == cH.Input_Layer || n2.Layer <= n1.Layer { //tez petla while; || to or; sprawdzamy czy drugi neuron nie jest w warstwie wejsciowej
		n2 = cH.Nodes[rand.Intn(len(cH.Nodes))] //nie moze byc!! bo z lewej strony nie ma juz warstw wiec w jakiej warstwei mialby
	} // byc pierwszy neuron

	// c to tutaj nasz wskaznik wskazujacy polaczenie miedzy losowymi neuronami, moze byc pusty->polecznie nie istnieje
	c := cH.Ch.Exists(&n1, &n2) // * - wyciaga wartosc ze wskaznika; & - wklada wartosc do wskaznika
	x := Connection{In_node: n1, //tu tworzymy nowe polaczenie miedzy neuronami nawet jesli juz istnieje
		Out_node: n2}

	if c != nil { //jesli wskaznik nie jes tpusty -> polaczenie juz istnialo (w jakimkolwiek genomie) to jego numer innowacji przypisujemy nowemu polaczneiu x
		x.Inno = c.Inno
		if !cH.Exists(x.Inno) { //tu sprawdzamy czy polaczenie istnieje w GENOMIE
			cH.Connections = append(cH.Connections, x) //glowna lista akutalnych polaczen w GENOMIE (a nie w calej populacji jak w przypadku Connectionh)
			n2.InConnections = append(n2.InConnections, x) //lista polaczen ktore wchodza do neuornu n2 #w type node musisz dodac ten atrybut (patrz wyzej)
		} // tu chcemy po prostu zapobiec dodaniu jakiegos drugiego polaczenia miedzy tymi samymi neuronami, i nawet jesli takie poalczenie kiedsy istialo
		// to moze go juz nie byc w genomie i wtedy mozemy na spokojnie je dodac znowu
	} 	else  {
		x.Inno = cH.Ch.Global_inno // jesli polaczenia nigdy nie bylo - nowe innovation number
		cH.Ch.Global_inno += 1 //dodajemy sobie tu 1 zeby przygotowac nasteona liczbe dla nowego polaczenia, ktore nigdy nie istnialo
		cH.Connections = append(cH.Connections, x) //dodajemy polaczenie do genomu
		cH.Ch.AllConnections = append(cH.Ch.AllConnections, x.copy()) //musimy tu dodac kopie polaczenia, ale trzeba zdefiniowac funckje copy
		n2.InConnections = append(n2.InConnections, x) //dodajemy do listy wejsiowych polaczen neuronu n2
	}} //musimy miec dwa rozne polaczenia dla populacji i dla genomu bo np. mutacje w jednym genomie nie moga wplywac na historie polaczen populacji
// InConnections dla neuronu sa wazne dla backwawrd i forward propagation
func (cH *Genom) Exists(nn int) bool {
	for _, c := cH.Connections {
		if c.Inno == nn {
			return True
		}
	return False
	}
}