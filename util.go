package main

import (
	"bytes"
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
	return WIKIWORD_ONLY_RE.MatchString(word)
}

func ExpandWikiWords(input []byte) []byte {
	l, ch := NewLexer(input)
	go l.Run()

	buf := &bytes.Buffer{}

	done := false
	for !done {
		item := <-ch
		switch item.Type {
		case TokenErr:
			done = true
		case TokenEOF:
			//fmt.Println("Got an EOF")
			done = true
		case TokenWikiWord:
			buf.Write([]byte("["))
			buf.Write(item.Value)
			buf.Write([]byte("](/"))
			buf.Write(item.Value)
			buf.Write([]byte("/)"))
		default:
			buf.Write(item.Value)
		}
	}
	//fmt.Println("Returning ", buf.String())
	return buf.Bytes()
}

func writeAndClose(wc io.WriteCloser, value []byte) error {
	defer wc.Close()
	_, err := wc.Write(value)
	return err
}

func copyAndClose(wc io.WriteCloser, input io.Reader) error {
	defer wc.Close()
	_, err := io.Copy(wc, input)
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

func getCountFDBAttachments(fInfos []os.FileInfo) int {
	count := 0
	for _, info := range fInfos {
		name := path.Base(info.Name())
		if fdb_Attachment_re.MatchString(name) {
			count++
		}
	}
	return count
}

func generateRevisionSplit(currentRevision, maxRevision int) (cur, min, max int) {
	cur = currentRevision
	if currentRevision < 0 {
		cur = maxRevision
	}
	max = cur + 5
	if max > maxRevision {
		max = maxRevision
	}
	min = cur - 5
	if min < 0 {
		min = 0
	}
	return cur, min, max
}

func generateInt(begin, end int) <-chan int {
	ch := make(chan int)

	go func() {
		defer close(ch)
		step := 1
		if begin >= end {
			step = -1
		}
		for i := begin; i != end; i += step {
			ch <- i
		}
		ch <- end
	}()
	return ch
}

func getListenAddress() string {
	return ""
}
