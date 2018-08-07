package flexihash

import (
	"hash/crc32"
)

// flexihash is an easy to use consistent hashing implementation

// Node is a member of the hash ring
type Node struct {
	Weight int
	Thing  interface{}
}

// GetNode gets a node from the ring
func GetNode(key []byte, nodes []Node) *Node {
	if len(nodes) == 0 {
		panic("empty nodes")
	}

	totalWeight := 0
	for i := range nodes {
		totalWeight += nodes[i].Weight
	}
	targetWeight := crc32.ChecksumIEEE(key) % totalWeight
	totalWeight = 0
	for i := range nodes {
		if totalWeight+nodes[i].Weight > targetWeight {
			return &nodes[i]
		}
		totalWeight += n.Weight
	}

	// should never happen
	return nil
}

// GetThing get thing from equal weighted things
func GetThing(key []byte, things []interface{}) interface{} {
	nodes := make([]Node, len(things))
	for i: range things {
		nodes[i] = Node{Weight: 1, Thing: things[i]}
	}
	node := getNode(key, nodes)
	return node.Thing
}
