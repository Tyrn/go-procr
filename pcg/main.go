// pcg project main.go
package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

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

// Returns True, if path has extension ext, case and leading dot insensitive
func HasExtOf(pth, ext string) bool {
	xt := strings.Trim(filepath.Ext(pth), ". ")
	return strings.ToUpper(xt) == strings.ToUpper(strings.Trim(ext, ". "))
}

// Returns a vector of integer numbers
// embedded in a string argument
func StrStripNumbers(s string) []int {
	r, _ := regexp.Compile(`\d+`)
	var p []int
	for _, v := range r.FindAllString(s, -1) {
		n, _ := strconv.Atoi(v)
		p = append(p, n)
	}
	if p == nil {
		return []int{}
	}
	return p
}

// Compares arrays of integers using 'string semantics'
func ArrayCmp(x, y []int) int {
	if len(x) == 0 {
		if len(y) == 0 {
			return 0
		} else {
			return -1
		}
	}
	if len(y) == 0 {
		if len(x) == 0 {
			return 0
		} else {
			return 1
		}
	}
	i := 0
	for x[i] == y[i] {
		if i == len(x)-1 || i == len(y)-1 {
			// Short array is a prefix of the long one; end reached. All is equal so far.
			if len(x) == len(y) {
				// Long array is no longer than the short one.
				return 0
			}
			if len(x) < len(y) {
				return -1
			}
			return 1
		}
		i++
	}
	// Difference encountered.
	if x[i] < y[i] {
		return -1
	}
	return 1
}

// If both strings contain digits, returns numerical comparison based on the numeric
// values embedded in the strings, otherwise returns the standard string comparison.
// The idea of the natural sort as opposed to the standard lexicographic sort is one of coping
// with the possible absence of the leading zeros in 'numbers' of files or directories
func StrCmpNaturally(x, y string) int {
	a := StrStripNumbers(x)
	b := StrStripNumbers(y)
	if len(a) > 0 && len(b) > 0 {
		return ArrayCmp(a, b)
	} else {
		return strings.Compare(x, y)
	}
}

// Reduces a string of names to initials
func MakeInitials(name string) string {
	sep := "."
	trail := "."
	hyph := "-"

	// Remove double quoted substring, if any.
	qr, _ := regexp.Compile(`"`)
	quotes := qr.FindAllString(name, -1)
	qcnt := len(quotes)
	sr, _ := regexp.Compile(`"(.*?)"`) // Replace double quoted substrings.
	enm := ""
	if qcnt == 0 || qcnt%2 != 0 {
		enm = name
	} else {
		enm = sr.ReplaceAllString(name, " ")
	}

	pr, _ := regexp.Compile(`\s+`) // Split by any space.

	// Split (already hyphenless) (sub)string by spaces
	// and reduce every one of them to an uppercase first letter.
	splitBySpace := func(nm string) []string {
		spl := pr.Split(nm, -1)
		var ini []string
		for _, v := range spl {
			if len(v) >= 1 {
				// At least one character to handle.
				ini = append(ini, strings.ToUpper(v[:1]))
			}
		}
		return ini
	}
	fr, _ := regexp.Compile(hyph) // Split by hyph.
	spl := fr.Split(enm, -1)
	var r []string
	for _, v := range spl {
		r = append(r, strings.Join(splitBySpace(strings.TrimSpace(v)), sep))
	}
	return strings.Join(r, hyph) + trail
}

func main() {
	kingpin.Parse()
	sansExt := SansExt("/alfa/bravo/moo/charlie.flac")
	fmt.Printf("%v, %s %s\n", *help, *file_type, sansExt)
}
