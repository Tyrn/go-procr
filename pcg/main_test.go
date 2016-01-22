package main

import (
	"fmt"
	"testing"
)

func TestSanExt(t *testing.T) {
	fmt.Println("Moo!")
	fmt.Println(SansExt("/alfa/bravo/charlie.flac"))
}
