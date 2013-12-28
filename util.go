package main

import (
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
)

const _WIKIWORD_RE = "([A-Z]+[A-Za-z0-9_]*){2,}"
const _WIKIWORD_ONLY_RE = "^" + _WIKIWORD_RE + "$"

var WIKIWORD_RE = regexp.MustCompile(_WIKIWORD_RE)
var WIKIWORD_ONLY_RE = regexp.MustCompile(_WIKIWORD_ONLY_RE)

// Is the given string a WikiWord
func IsWikiWord(word string) bool {
	return WIKIWORD_RE.MatchString(word)
}

// Given a block of text as a []byte, return the list of WikiWords
func ExtractWikiWords(input []byte) [][]byte {
	// FIXME,  The uniqueness set is funny, []byte -> string -> []byte bad
	// potentially a lot of garbage
	unique := make(map[string]bool)
	for _, val := range WIKIWORD_RE.FindAll(input, -1) {
		unique[string(val)] = true
	}
	results := make([][]byte, 0, len(unique))
	for key, _ := range unique {
		results = append(results, []byte(key))
	}
	return results
}

func writeAndClose(wc io.WriteCloser, value []byte) error {
	defer wc.Close()
	_, err := wc.Write(value)
	return err
}

func getMaxFDBRevision(fInfos []os.FileInfo) int {
	maxRevision := -1
	for _, info := range fInfos {
		name := path.Base(info.Name())
		if fdb_Page_re.MatchString(name) {
			// Atoi cannot fail after passing the re test above
			if cur, err := strconv.Atoi(name); err == nil && cur > maxRevision {
				maxRevision = cur
			}
		}
	}
	return maxRevision
}
