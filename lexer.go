package main

import (
	"errors"
	"io"
	"unicode"
	"unicode/utf8"
)

type LexToken int

/* Lexer is heavily based on the Lexer presented in Rob Pike's
Lexical Analyzer in Go talk of 2010.

Though it must be admited that the Go team's version had better names and
makes a little more sense.
*/
type Lexer struct {
	input []byte
	start int
	cur   int
	ch    chan LexedItem
}

type LexedItem struct {
	Type  LexToken
	Value []byte
}

type stageFunc func(*Lexer) stageFunc

const (
	TokenErr LexToken = iota
	TokenText
	TokenLink
	TokenImage
	TokenWikiWord
	TokenEOF
)

func (lt LexToken) String() string {
	switch lt {
	case TokenErr:
		return "Lexer error"
	case TokenText:
		return "Lexed text"
	case TokenLink:
		return "Lexed link"
	case TokenImage:
		return "Lexed image"
	case TokenWikiWord:
		return "Lexed WikiWord"
	case TokenEOF:
		return "Lexed EOF"
	}
	return "lexed token"
}

func NewLexer(input []byte) (*Lexer, chan LexedItem) {
	// we buffer 3 deep to allow tests to run
	l := &Lexer{input: input, ch: make(chan LexedItem, 4)}
	return l, l.ch
}

func (l *Lexer) Next() (rune, error) {
	if l.cur < len(l.input) {
		if result, len := utf8.DecodeRune(l.input[l.cur:]); result != utf8.RuneError {
			l.cur += len
			return result, nil
		}
	}
	//fmt.Println("Returning io.EOF!!!!!!!!!!")
	return ' ', io.EOF
}

func (l *Lexer) IsEOF(err error) bool {
	if err == io.EOF {
		return true
	}
	return false
}

func (l *Lexer) Reverse(r rune) error {
	rLen := utf8.RuneLen(r)
	if l.cur >= rLen {
		l.cur -= rLen
		return nil
	}
	if l.start > l.cur {
		l.cur = l.start
	}
	return errors.New("Insufficient buffer to reverse lexer")
}

func (l *Lexer) emit(Type LexToken) {
	segLen := l.cur - l.start
	//fmt.Println(Type)
	//fmt.Printf("segLen %d = %d - %d\n", segLen, l.cur, l.start)
	////fmt.Printf("'%s'\n", string(l.input))
	item := LexedItem{Type: Type, Value: make([]byte, segLen)}
	copy(item.Value, l.input[l.start:l.cur])
	//fmt.Printf("Index values %d %d\nEmitting (%s) '%s'\n", l.start, l.cur, Type, string(item.Value))
	l.start = l.cur
	//fmt.Println("Start advanced to", l.start)
	l.ch <- item
}

func (l *Lexer) saveLocation() int {
	return l.cur
}

func (l *Lexer) jumpToLocation(pos int) error {
	if pos < l.start-1 || pos > len(l.input) {
		return errors.New("location out of range")
	}
	l.cur = pos
	return nil
}

func (l *Lexer) Run() {
	for state := textLexer; state != nil; {
		state = state(l)
		//fmt.Println("Next lexer state is", state)
	}
	close(l.ch)
}

func textLexer(l *Lexer) stageFunc {
	InWikiWord := false
	CanStartWikiWord := true
	UpperCount := 0
	BeforeWikiWordStart := -1

	resetWikiWord := func() {
		InWikiWord = false
		CanStartWikiWord = true
		UpperCount = 0
	}
	startWikiWord := func() {
		InWikiWord = true
		UpperCount = 1
	}
	emitCurrent := func() {
		//fmt.Println("In emitCurrent")
		if InWikiWord {
			//fmt.Println("InWikiWord")
			// flush any pending text prior to the wiki word
			if BeforeWikiWordStart >= 0 {
				//fmt.Println("There is something earlier")
				curPos := l.saveLocation()
				if err := l.jumpToLocation(BeforeWikiWordStart); err != nil {
					//fmt.Println("Unable to jump to location -", BeforeWikiWordStart, " start is ", l.start)
				}
				l.emit(TokenText)
				//fmt.Println("Returning to current")
				if err := l.jumpToLocation(curPos); err != nil {
					//fmt.Println("Unable to jump to location")
				}
			}
			l.emit(TokenWikiWord)
		} else {
			l.emit(TokenText)
		}
		BeforeWikiWordStart = l.saveLocation()
		resetWikiWord()
		//fmt.Println("Before advanced to ", BeforeWikiWordStart)
		//fmt.Println("leaving emitCurrent")
	}

	var r rune = ' '
	var err error

	for {
		r, err = l.Next()
		if err != nil {
			if l.IsEOF(err) {
				//fmt.Println("EOF")
				emitCurrent()
				l.emit(TokenEOF)
				return nil
			}
			l.start = l.cur
			l.emit(TokenErr)
			return nil
		}
		//fmt.Printf("'%s' %v before:%d up:%d s:%d c:%d\n", string(r), InWikiWord, BeforeWikiWordStart, UpperCount, l.start, l.cur)

		if InWikiWord {
			if unicode.IsUpper(r) {
				UpperCount++
			}
			if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
				//fmt.Printf("'%s' let:%v dig:%v\n", string(r), unicode.IsLetter(r), unicode.IsDigit(r))
				// just left wiki word
				if UpperCount >= 2 {
					// This is a WikiWord
					//if BeforeWikiWordStart >= 0 {
					l.Reverse(r)
					emitCurrent()
					// do not advance, we need to re-evaluate the rune, outside of the context of a wiki word
					//l.Next()
					//}
				} else {
					//fmt.Println("Not a WikiWord")
				}
				// reset wiki word variables
				resetWikiWord()
			}
		} else {
			// not in wiki word
			if unicode.IsSpace(r) || unicode.IsPunct(r) {
				CanStartWikiWord = true
				BeforeWikiWordStart = l.saveLocation()
			} else if CanStartWikiWord && unicode.IsUpper(r) {
				//fmt.Println("Starting wiki word ")
				startWikiWord()
			} else {
				CanStartWikiWord = false
			}
		}

		if r == '[' {
			l.Reverse(r)
			l.emit(TokenText)
			return linkLexer
		} else if r == '!' {
			// this is overly complex, fix it
			//fmt.Println("Found a !")
			r, err = l.Next()
			if err != nil {
				if l.IsEOF(err) {
					l.emit(TokenText)
					l.emit(TokenEOF)
					return nil
				}
				//fmt.Println("Error: ", err)
				l.start = l.cur
				l.emit(TokenErr)
				return nil
			}
			// we always put this one back
			l.Reverse('[')
			if r == '[' {
				l.Reverse('!')
				l.emit(TokenText)
				return imageLexer
			}
		}
	}
}

// Match link type objects (images, links, and references)
// []() or [][] or []:
func linkMatcher(l *Lexer, Type LexToken) stageFunc {
	var r rune
	var err error

	matcher := func(startRune, endRune rune) bool {
		if r, err = l.Next(); err != nil || r != startRune {
			l.start = l.cur
			//fmt.Printf("Did not find matching '%s' for link type got '%s'.\n", string(startRune), string(r))
			l.emit(TokenErr)
			return false
		}
		for {
			r, err = l.Next()
			if err != nil {
				l.start = l.cur
				l.emit(TokenErr)
				return false
			}
			if r == endRune {
				break
			}
		}
		return true
	}

	// links are of the form [...][...] or []() so run through
	// this loop twice, each one matches [...] or ()

	//always look [] first
	if !matcher(rune('['), rune(']')) {
		return nil
	}
	// peek to see if we have '[' or '('
	r, err = l.Next()
	if err != nil {
		if err == io.EOF {
			l.emit(TokenText)
			l.emit(TokenEOF)
		} else {
			l.emit(TokenErr)
		}
		return nil
	}
	var start, end rune
	switch r {
	case rune('('):
		start = rune('(')
		end = rune(')')
	case rune('['):
		start = rune('[')
		end = rune(']')
	case rune(':'):
		l.emit(TokenText)
		return textLexer(l)
	default:
		//fmt.Println("Did not find link/image body")
		l.emit(TokenErr)
		return nil
	}
	l.Reverse(r)
	if !matcher(start, end) {
		return nil
	}
	l.emit(Type)
	return textLexer
}

func linkLexer(l *Lexer) stageFunc {
	return linkMatcher(l, TokenLink)
}

func imageLexer(l *Lexer) stageFunc {
	if r, err := l.Next(); err != nil || r != '!' {
		l.start = l.cur
		//fmt.Println("No ! found for image")
		l.emit(TokenErr)
		return nil
	}
	return linkMatcher(l, TokenImage)
}
