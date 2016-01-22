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
