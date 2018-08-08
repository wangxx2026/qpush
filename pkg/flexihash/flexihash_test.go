package flexihash

import (
	"fmt"
	"testing"
)

func TestFlexihash(t *testing.T) {

	f := NewFlexiHash([]interface{}{"a", "b", "c"}, nil)
	thing := f.Get([]byte("xub"))
	fmt.Println(thing)
}
