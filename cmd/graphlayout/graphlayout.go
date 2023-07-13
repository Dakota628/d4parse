package main

import (
	"bufio"
	"encoding/binary"
	"github.com/awalterschulze/gographviz"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var (
	refsFilePath  = filepath.Join("docs", "refs.bin")
	nodesFilePath = filepath.Join("docs", "nodes.bin")
	dotFilePath   = filepath.Join("docs", "graph.dot")
)

func main() {
	// Read refs file
	refsFile, err := os.Open(refsFilePath)
	if err != nil {
		log.Fatalf("Failed to open refs file: %s", err)
	}

	var refs [][2]int32
	r := bufio.NewReader(refsFile)
	for {
		ref := [2]int32{}
		if err = binary.Read(r, binary.LittleEndian, &ref[0]); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalf("Failed to read ref: %s", err)
			}
		}
		if err = binary.Read(r, binary.LittleEndian, &ref[1]); err != nil {
			log.Fatalf("Failed to read ref: %s", err)
		}
		refs = append(refs, ref)
	}

	graph := "G"
	g := gographviz.NewGraph()

	for _, ref := range refs {
		from := strconv.Itoa(int(ref[0]))
		to := strconv.Itoa(int(ref[1]))

		if !g.IsNode(from) {
			if err = g.AddNode(graph, from, nil); err != nil {
				log.Fatalf("Failed to add node %q: %s", from, err)
			}
		}

		if !g.IsNode(to) {
			if err = g.AddNode(graph, to, nil); err != nil {
				log.Fatalf("Failed to add node %q: %s", to, err)
			}
		}

		if err = g.AddEdge(from, to, true, nil); err != nil {
			log.Fatalf("Failed to add edge %q->%q: %s", from, to, err)
		}
	}

	if err = os.WriteFile(dotFilePath, []byte(g.String()), 0644); err != nil {
		log.Fatalf("Failed to write dot file: %s", err)
	}

	//// Create refs graph
	//g := simple.NewDirectedGraph()
	//for _, ref := range refs {
	//	if ref[0] == ref[1] {
	//		continue
	//	}
	//	from, _ := g.NodeWithID(int64(ref[0]))
	//	to, _ := g.NodeWithID(int64(ref[1]))
	//	g.SetEdge(g.NewEdge(from, to))
	//}
	//
	//// Setup graph layout
	//eades := layout.EadesR2{Repulsion: 0.5, Rate: 0.05, Updates: 30, Theta: 0.2}
	//optimizer := layout.NewOptimizerR2(g, eades.Update)
	//
	//// Perform layout optimization
	//for optimizer.Update() {
	//}
	//
	//// Generate the graph binary
	//nodesFile, err := os.Create(nodesFilePath)
	//if err != nil {
	//	log.Fatalf("Failed to create nodes file: %s", err)
	//}
	//w := bufio.NewWriter(nodesFile)
	//
	//nodes := optimizer.Nodes()
	//for nodes.Next() {
	//	n := nodes.Node()
	//	nid := n.ID()
	//	r2 := optimizer.LayoutNodeR2(nid)
	//
	//	snoId := int32(nid)
	//	x := r2.Coord2.X
	//	y := r2.Coord2.Y
	//	log.Printf("id=%d x=%f y=%f\n", snoId, x, y)
	//
	//	if err := errors.Join(
	//		binary.Write(w, binary.LittleEndian, &snoId),
	//		binary.Write(w, binary.LittleEndian, &x),
	//		binary.Write(w, binary.LittleEndian, &y),
	//	); err != nil {
	//		log.Fatalf("Failed to write node to nodes file: %s", err)
	//	}
	//}
	//
	//if err = w.Flush(); err != nil {
	//	log.Fatalf("Failed to flush nodes file: %s", err)
	//}
}
