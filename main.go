// TESTOWA DO NODE I CONNECTION
package main

import (
	"NEAT/data"
	"math/rand"
)

func main() {
	n1 := data.Node{Number: 1, Layer: 2}
	n2 := data.Node{Number: 2, Layer: 3}
	con := data.Connection{
		In_node:  n1,
		Out_node: n2,
		Weight:   rand.Float64()*2 - 1,
		Inno:     -1,
		Enabled:  true,
	}
	con.ShowConn()

}
