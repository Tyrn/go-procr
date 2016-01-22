// pcg project main.go
package main

import (
	"fmt"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	help         = kingpin.Flag("help", "Prints help").Short('h').Bool()
	verbose      = kingpin.Flag("verbose", "Verbose output").Short('v').Bool()
	file_title   = kingpin.Flag("file-title", "Use file name for title tag").Short('f').Bool()
	sort_lex     = kingpin.Flag("sort-lex", "Sort files lexicographically").Short('x').Bool()
	tree_dst     = kingpin.Flag("tree-dst", "Retain the tree structure of the source album at destination").Short('t').Bool()
	drop_dst     = kingpin.Flag("drop_dst", "Do not create destination directory").Short('p').Bool()
	reverse      = kingpin.Flag("reverse", "Copy files in reverse order (last file first)").Short('r').Bool()
	file_type    = kingpin.Flag("file-type", "Accept only audio files of the specified type").Short('e').String()
	unified_name = kingpin.Flag("unified-name", "Base name for everything but the \"Artist\" tag").Short('u').String()
	album_num    = kingpin.Flag("album-num", "Album number").Short('b').String()
	artist_tag   = kingpin.Flag("artist-tag", "\"Artist\" tag").Short('a').String()
	album_tag    = kingpin.Flag("album-tag", "\"Album\" tag").Short('g').String()
)

// Discards file extension
func SansExt(pth string) string {
	dir, file := filepath.Split(pth)
	return filepath.Join(dir, file[:len(file)-len(filepath.Ext(pth))])
}

func main() {
	kingpin.Parse()
	sansExt := SansExt("/alfa/bravo/moo/charlie.flac")
	fmt.Printf("%v, %s %s\n", *help, *file_type, sansExt)
}
