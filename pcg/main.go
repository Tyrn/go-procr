// pcg project main.go
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pebbe/zmq4"
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

	context, _   = zmq4.NewContext()
	requester, _ = context.NewSocket(zmq4.REQ)
)

// Parses the command line and checks some conditions
func ParseArgs() {
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	if !IsDir(*src_dir) {
		fmt.Printf("Source directory \"%s\" is not there.\n", *src_dir)
		os.Exit(2)
	}
	if !IsDir(*dst_dir) {
		fmt.Printf("Destination path \"%s\" is not there.\n", *dst_dir)
		os.Exit(2)
	}
}

// Return true, if directory
func IsDir(pth string) bool {
	finfo, err := os.Stat(pth)
	if err != nil {
		return false
	}
	if finfo.IsDir() {
		return true
	}
	return false
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

// Recursively traverses the source directory and copies all the audio files to destination;
// the destination directory and file names get decorated according to options.
// All files go to a single destination directory
func TraverseFlatDst(srcDir, dstRoot string, fcount *int, tot int) {
	dirs, files := ListDirGroom(srcDir)
	for _, v := range dirs {
		TraverseFlatDst(v, dstRoot, fcount, tot)
	}
	for _, v := range files {
		dst := filepath.Join(dstRoot, DecorateFileName(len(strconv.Itoa(tot)), *fcount, BaseName(v)))
		CopyFile(*fcount, tot, v, dst)
		*fcount++
	}
}

// Recursively traverses the source directory and copies all the audio files to destination;
// the destination directory and file names get decorated according to options.
// The copy order is reverse;
// All files go to a single destination directory
func TraverseFlatDstReverse(srcDir, dstRoot string, fcount *int, tot int) {
	dirs, files := ListDirGroom(srcDir)
	for _, v := range files {
		dst := filepath.Join(dstRoot, DecorateFileName(len(strconv.Itoa(tot)), *fcount, BaseName(v)))
		CopyFile(*fcount, tot, v, dst)
		*fcount--
	}
	for _, v := range dirs {
		TraverseFlatDstReverse(v, dstRoot, fcount, tot)
	}
}

// Recursively traverses the source directory and copies all the audio files to destination;
// the destination directory and file names get decorated according to options
func TraverseTreeDst(srcDir, dstRoot, dstStep string, fcount *int, tot int) {
	dirs, files := ListDirGroom(srcDir)
	for i, v := range dirs {
		step := filepath.Join(dstStep, DecorateDirName(i, BaseName(v)))
		os.Mkdir(filepath.Join(dstRoot, step), 0777)
		TraverseTreeDst(v, dstRoot, step, fcount, tot)
	}
	for i, v := range files {
		dstFile := filepath.Join(dstStep, DecorateFileName(len(strconv.Itoa(tot)), i, BaseName(v)))
		CopyFile(*fcount, tot, v, filepath.Join(dstRoot, dstFile))
		*fcount++
	}
}

// Copy one audio file to destination and set its tags
func CopyFile(i, tot int, src, dst string) {

	CopySync(src, dst)

	buildTitle := func(s string) string {
		title := fmt.Sprintf("%d %s", i, s)
		if *file_title {
			title = SansExt(BaseName(dst))
		}
		return title
	}
	buildTag := func(tag, value string) string {
		return fmt.Sprintf(",\"%s\":\"%s\"", tag, value)
	}

	rqs := fmt.Sprintf("{\"request\":\"settags\",\"file\":\"%s\",\"tags\":{\"tracknumber\":\"%d/%d\"", dst, i, tot)

	if len(*artist_tag) > 0 && len(*album_tag) > 0 {
		rqs += buildTag("title", buildTitle(MakeInitials(*artist_tag))+" - "+*album_tag)
		rqs += buildTag("artist", *artist_tag)
		rqs += buildTag("album", *album_tag)
	} else if len(*artist_tag) > 0 {
		rqs += buildTag("title", buildTitle(*artist_tag))
		rqs += buildTag("artist", *artist_tag)
	} else if len(*album_tag) > 0 {
		rqs += buildTag("title", buildTitle(*album_tag))
		rqs += buildTag("album", *album_tag)
	}

	rqs += "}}"

	_, serr := requester.Send(rqs, 0)
	if serr != nil {
		fmt.Printf("Request error\n")
		os.Exit(2)
	}
	_, rerr := requester.Recv(0)
	if rerr != nil {
		fmt.Printf("Reply error\n")
		os.Exit(2)
	}

	if *verbose {
		fmt.Printf("%5d/%d %s\n", i, tot, dst)
	} else {
		fmt.Printf(".")
	}
}

// Copies src file to dst
func CopySync(src, dst string) (int64, error) {
	src_file, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer src_file.Close()

	src_file_stat, err := src_file.Stat()
	if err != nil {
		return 0, err
	}

	if !src_file_stat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	dst_file, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dst_file.Close()
	return io.Copy(dst_file, src_file)
}

// Traverses the source tree according to options
func Groom(src, dst string, tot int) {

	err := requester.Connect("tcp://localhost:64107")
	defer requester.Close()
	if err != nil {
		fmt.Printf("Connection failed.\n")
		os.Exit(2)
	}

	if !*verbose {
		fmt.Printf("Starting ")
	}

	if *tree_dst {
		if *reverse {
			fmt.Printf("Remove either -t or -r\n")
			os.Exit(2)
		} else {
			c := 1
			TraverseTreeDst(src, dst, "", &c, tot)
		}
	} else {
		if *reverse {
			c := tot
			TraverseFlatDstReverse(src, dst, &c, tot)
		} else {
			c := 1
			TraverseFlatDst(src, dst, &c, tot)
		}
	}

	if !*verbose {
		fmt.Printf(" Done (%d)\n", tot)
	}
}

// Sets up boilerplate required by the options
// and actually runs the copying
func BuildAlbum(src, dst string) {
	srcName := BaseName(src)
	prefix := ""
	if len(*album_num) > 0 {
		n, _ := strconv.Atoi(*album_num)
		prefix = ZeroPad(2, n) + "-"
	}
	baseDst := srcName
	if len(*unified_name) > 0 {
		baseDst = prefix + *unified_name
	}
	ex := baseDst
	if *drop_dst {
		ex = ""
	}

	executiveDst := filepath.Join(dst, ex)
	tot := FileCount(src, IsAudioFile)
	if tot < 1 {
		fmt.Printf("There are no supported audio files in the source directory \"%s\".\n", src)
		os.Exit(2)
	}

	if !*drop_dst {
		if IsDir(executiveDst) {
			fmt.Printf("Destination directory \"%s\" already exists.\n", executiveDst)
			os.Exit(2)
		} else {
			os.Mkdir(executiveDst, 0777)
		}
	}
	Groom(src, executiveDst, tot)
}

func main() {
	ParseArgs()
	BuildAlbum(strings.TrimRight(*src_dir, "/\\"), strings.TrimRight(*dst_dir, "/\\"))
}
