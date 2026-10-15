// ringsim shows key distribution and remapping when a node joins.
package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/avohra25/hashring"
)

func main() {
	nodes := flag.Int("nodes", 5, "initial node count")
	keys := flag.Int("keys", 100000, "number of keys")
	vnodes := flag.Int("vnodes", 160, "virtual nodes per node")
	flag.Parse()

	r := hashring.New(*vnodes)
	for i := 0; i < *nodes; i++ {
		r.Add("node-" + strconv.Itoa(i))
	}
	before := make([]string, *keys)
	count := map[string]int{}
	for i := range before {
		before[i], _ = r.Get("key-" + strconv.Itoa(i))
		count[before[i]]++
	}
	fmt.Printf("Distribution of %d keys over %d nodes (%d vnodes each):\n", *keys, *nodes, *vnodes)
	for _, n := range r.Nodes() {
		fmt.Printf("  %-8s %6d (%.1f%%)\n", n, count[n], 100*float64(count[n])/float64(*keys))
	}
	r.Add("node-new")
	moved := 0
	for i, old := range before {
		if n, _ := r.Get("key-" + strconv.Itoa(i)); n != old {
			moved++
		}
	}
	fmt.Printf("Adding 1 node moved %d keys (%.1f%%; ideal ~%.1f%%)\n",
		moved, 100*float64(moved)/float64(*keys), 100/float64(*nodes+1))
}
