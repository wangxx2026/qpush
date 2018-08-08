package flexihash

import (
	"hash/crc32"
)

// flexihash is an easy to use consistent hashing implementation

// Node is a member of the hash ring
type Node struct {
	Weight uint32
	Thing  interface{}
}

// FlexiHash for flexihash
type FlexiHash struct {
	nodes       []Node
	totalWeight uint32
}

// NewFlexiHash returns a flexihash instance
func NewFlexiHash(things []interface{}, weights []uint32) *FlexiHash {
	if len(things) == 0 {
		panic("empty")
	}

	if weights == nil {
		weights = make([]uint32, len(things))
		for idx := range weights {
			weights[idx] = 1
		}
	}
	nodes := make([]Node, len(things))
	totalWeight := uint32(0)
	for idx := range things {
		nodes[idx] = Node{Weight: weights[idx], Thing: things[idx]}
		totalWeight += weights[idx]
	}

	return &FlexiHash{nodes: nodes, totalWeight: totalWeight}
}

// getNode gets a node from the ring
func (f *FlexiHash) getNode(key []byte) *Node {

	targetWeight := crc32.ChecksumIEEE(key) % f.totalWeight
	sumWeight := uint32(0)
	for i := range f.nodes {
		if sumWeight+f.nodes[i].Weight > targetWeight {
			return &f.nodes[i]
		}
		sumWeight += f.nodes[i].Weight
	}

	// should never happen
	return nil
}

// Get a thing by flexihash
func (f *FlexiHash) Get(key []byte) interface{} {
	node := f.getNode(key)
	return node.Thing
}
