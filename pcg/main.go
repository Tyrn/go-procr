// pcg project main.go
package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
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
	src_dir      = kingpin.Arg("src", "Source directory").Required().String()
	dst_dir      = kingpin.Arg("dst", "Destination directory").Required().String()
)

func ParseArgs() {
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	fmt.Printf("%s %s \"%s\"\n", strings.TrimRight(*src_dir, "/\\"), strings.TrimRight(*dst_dir, "/\\"), *album_num)
}

// Discards file extension
func SansExt(pth string) string {
	dir, file := filepath.Split(pth)
	return filepath.Join(dir, file[:len(file)-len(filepath.Ext(pth))])
}

// Returns base name (file name complete with extension)
func BaseName(pth string) string {
	_, file := filepath.Split(pth)
	return file
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

// Returns true, if pth is a recognized audio file
func IsAudioFile(pth string) bool {
	rx := []string{".MP3", ".M4A", ".M4B", ".OGG", ".WMA", ".FLAC"}
	for _, v := range rx {
		if HasExtOf(pth, v) {
			return true
		}
	}
	return false
}

// Returns a list of directories in absPath directory, and a list of files filtered by fileCondition
func CollectDirsAndFiles(absPath string, fileCondition func(string) bool) ([]string, []string) {
	haul, _ := ioutil.ReadDir(absPath)
	var dirs, files []string
	for _, v := range haul {
		if v.IsDir() {
			dirs = append(dirs, filepath.Join(absPath, v.Name()))
		} else {
			if fileCondition(v.Name()) {
				files = append(files, filepath.Join(absPath, v.Name()))
			}
		}
	}
	return dirs, files
}

// Returns a total number of files in the dirPath directory filtered by fileCondition
func FileCount(dirPath string, fileCondition func(string) bool) int {
	cnt := 0
	dirs, files := CollectDirsAndFiles(dirPath, fileCondition)
	for _, v := range dirs {
		cnt += FileCount(v, fileCondition)
	}
	for _, v := range files {
		if fileCondition(v) {
			cnt++
		}
	}
	return cnt
}

// Compares two paths, ignoring extensions
func ComparePath(xp, yp string) int {
	x := SansExt(xp)
	y := SansExt(yp)
	if *sort_lex {
		return strings.Compare(x, y)
	}
	return StrCmpNaturally(x, y)
}

// Compares two paths, filenames only, ignoring extensions
func CompareFile(xf, yf string) int {
	x := SansExt(BaseName(xf))
	y := SansExt(BaseName(yf))
	if *sort_lex {
		return strings.Compare(x, y)
	}
	return StrCmpNaturally(x, y)
}

// Sorting interface implementation: natural, lexicographical, and
// reverse order, according to command line options.

type CustomDir []string

func (s CustomDir) Len() int {
	return len(s)
}
func (s CustomDir) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s CustomDir) Less(i, j int) bool {
	if *reverse {
		return ComparePath(s[i], s[j]) > 0
	}
	return ComparePath(s[i], s[j]) < 0
}

type CustomFile []string

func (s CustomFile) Len() int {
	return len(s)
}
func (s CustomFile) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s CustomFile) Less(i, j int) bool {
	if *reverse {
		return CompareFile(s[i], s[j]) > 0
	}
	return CompareFile(s[i], s[j]) < 0
}

// Returns (0) a naturally sorted list of
// offspring directory paths (1) a naturally sorted list
// of offspring file paths.
func ListDirGroom(absPath string) ([]string, []string) {
	dirs, files := CollectDirsAndFiles(absPath, IsAudioFile)
	sort.Sort(CustomDir(dirs))
	sort.Sort(CustomFile(files))
	return dirs, files
}

func ZeroPad(w, i int) string {
	fs := fmt.Sprintf("%%0%dd", w)
	return fmt.Sprintf(fs, i)
}

func DecorateDirName(i int, name string) string {
	return ZeroPad(3, i) + "-" + name
}

func DecorateFileName(cntw, i int, name string) string {
	if len(*unified_name) > 0 {
		return ZeroPad(cntw, i) + "-" + *unified_name + filepath.Ext(name)
	}
	return ZeroPad(cntw, i) + "-" + name
}

func TraverseFlatDst(srcDir, dstRoot string, fcount *int, cntw int) {
	dirs, files := ListDirGroom(srcDir)
	for _, v := range dirs {
		TraverseFlatDst(v, dstRoot, fcount, cntw)
	}
	for _, v := range files {
		dst := filepath.Join(dstRoot, DecorateFileName(cntw, *fcount, BaseName(v)))
		fmt.Printf("%d><%s**%s\n", *fcount, v, dst)
		*fcount++
	}
}

func main() {
	ParseArgs()
	dirs, files := ListDirGroom(*src_dir)
	fmt.Printf("%v\n%v\n", dirs, files)
	fmt.Printf("ZeroPad(4, 16): \"%s\"\n", ZeroPad(4, 16))
	fmt.Printf("FileCount(): %d\n", FileCount(*src_dir, IsAudioFile))
	cnt := 1
	TraverseFlatDst(*src_dir, *dst_dir, &cnt, 3)
}
