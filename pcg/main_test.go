package main

import (
	//	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tsbe = "they should be equal"

func TestSansExt(t *testing.T) {
	assert.Equal(t, "/alfa/bravo/charlie", SansExt("/alfa/bravo/charlie.flac"), tsbe)
	assert.Equal(t, "/alfa/bravo/charlie", SansExt("/alfa/bravo/charlie"), tsbe)
	assert.Equal(t, "/alfa/bravo/charlie", SansExt("/alfa/bravo/charlie/"), tsbe)
	assert.Equal(t, "/alfa/bra.vo/charlie", SansExt("/alfa/bra.vo/charlie.dat"), tsbe)
	assert.Equal(t, "", SansExt(""), tsbe)
}

func TestHasExtOf(t *testing.T) {
	assert.Equal(t, true, HasExtOf("/alfa/bra.vo/charlie.ogg", "OGG"), tsbe)
	assert.Equal(t, true, HasExtOf("/alfa/bra.vo/charlie.ogg", ".ogg"), tsbe)
	assert.Equal(t, false, HasExtOf("/alfa/bra.vo/charlie.ogg", "mp3"), tsbe)
}

func TestStrStripNumbers(t *testing.T) {
	assert.Equal(t, []int{11, 2, 144}, StrStripNumbers("ab11cdd2k.144"), tsbe)
	assert.Equal(t, []int{}, StrStripNumbers("Ignacio Vazquez-Abrams"), tsbe)
}

func TestArrayCmp(t *testing.T) {
	assert.Equal(t, 0, ArrayCmp([]int{}, []int{}), tsbe)
	assert.Equal(t, 1, ArrayCmp([]int{1}, []int{}), tsbe)
	assert.Equal(t, 1, ArrayCmp([]int{3}, []int{}), tsbe)
	assert.Equal(t, -1, ArrayCmp([]int{1, 2, 3}, []int{1, 2, 3, 4, 5}), tsbe)
	assert.Equal(t, -1, ArrayCmp([]int{1, 4}, []int{1, 4, 16}), tsbe)
	assert.Equal(t, 1, ArrayCmp([]int{2, 8}, []int{2, 2, 3}), tsbe)
	assert.Equal(t, -1, ArrayCmp([]int{0, 0, 2, 4}, []int{0, 0, 15}), tsbe)
	assert.Equal(t, 1, ArrayCmp([]int{0, 13}, []int{0, 2, 2}), tsbe)
	assert.Equal(t, 0, ArrayCmp([]int{11, 2}, []int{11, 2}), tsbe)
}

func TestCmpStrNaturally(t *testing.T) {
	assert.Equal(t, 0, StrCmpNaturally("", ""), tsbe)
	assert.Equal(t, -1, StrCmpNaturally("2a", "10a"), tsbe)
	assert.Equal(t, -1, StrCmpNaturally("alfa", "bravo"), tsbe)
}

func TestMakeInitials(t *testing.T) {
	assert.Equal(t, "J.R.R.T.", MakeInitials("John ronald reuel Tolkien"), tsbe)
	assert.Equal(t, "A.C-G.", MakeInitials("Apsley Cherry-Garrard"), tsbe)
}
