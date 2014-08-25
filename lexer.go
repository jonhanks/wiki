package main

import (
	"errors"
	"fmt"
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
	fmt.Println(Type)
	fmt.Printf("segLen %d = %d - %d\n", segLen, l.cur, l.start)
	//fmt.Printf("'%s'\n", string(l.input))
	item := LexedItem{Type: Type, Value: make([]byte, segLen)}
	copy(item.Value, l.input[l.start:l.cur])
	fmt.Printf("Index values %d %d\nEmitting '%s'\n", l.start, l.cur, string(item.Value))
	l.start = l.cur
	l.ch <- item
}

func (l *Lexer) saveLocation() int {
	return l.cur
}

func (l *Lexer) jumpToLocation(pos int) error {
	if pos < l.start || pos > len(l.input) {
		return errors.New("location out of range")
	}
	l.cur = pos
	return nil
}

func (l *Lexer) Run() {
	for state := textLexer; state != nil; {
		state = state(l)
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
		fmt.Println("In emitCurrent")
		if InWikiWord {
			fmt.Println("InWikiWord")
			// flush any pending text prior to the wiki word
			if BeforeWikiWordStart >= 0 {
				fmt.Println("There is something earlier")
				curPos := l.saveLocation()
				if err := l.jumpToLocation(BeforeWikiWordStart); err != nil {
					fmt.Println("Unable to jump to location")
				}
				l.emit(TokenText)
				fmt.Println("Returning to current")
				if err := l.jumpToLocation(curPos); err != nil {
					fmt.Println("Unable to jump to location")
				}
			}
			l.emit(TokenWikiWord)
		} else {
			l.emit(TokenText)
		}
	}

	var r rune = ' '
	var err error

	for {
		r, err = l.Next()
		if err != nil {
			if l.IsEOF(err) {
				fmt.Println("EOF")
				emitCurrent()
				l.emit(TokenEOF)
				return nil
			}
			l.start = l.cur
			l.emit(TokenErr)
			return nil
		}
		fmt.Printf("%s %v %d\n", string(r), InWikiWord, BeforeWikiWordStart)

		if InWikiWord {
			if unicode.IsUpper(r) {
				UpperCount++
			}
			if !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
				fmt.Printf("%s %v %v\n", string(r), unicode.IsLetter(r), unicode.IsDigit(r))
				// just left wiki word
				if UpperCount >= 2 {
					// This is a WikiWord
					if BeforeWikiWordStart >= 0 {
						l.Reverse(r)
						emitCurrent()
						l.Next()
					}
				}
				fmt.Println("Not a WikiWord")
				// reset wiki word variables
				resetWikiWord()
			}
		} else {
			// not in wiki word
			if unicode.IsSpace(r) || unicode.IsPunct(r) {
				CanStartWikiWord = true
				BeforeWikiWordStart = l.saveLocation()
			} else if CanStartWikiWord && unicode.IsUpper(r) {
				fmt.Println("Starting wiki word ")
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
			fmt.Println("Found a !")
			r, err = l.Next()
			if err != nil {
				if l.IsEOF(err) {
					l.emit(TokenText)
					l.emit(TokenEOF)
					return nil
				}
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
	l.cur = len(l.input)
	l.emit(TokenText)
	return nil
}

func linkMatcher(l *Lexer, Type LexToken, startRunes, endRunes [2]rune) stageFunc {
	var r rune
	var err error
	// links are of the form [...][...] or []() so run through
	// this loop twice, each one matches [...] or ()
	// the delimiters are specified in startRunes and endRunes
	for i := 0; i < 2; i++ {
		if r, err = l.Next(); err != nil || r != startRunes[i] {
			l.start = l.cur
			l.emit(TokenErr)
			return nil
		}
		for {
			r, err = l.Next()
			if err != nil {
				l.start = l.cur
				l.emit(TokenErr)
				return nil
			}
			if r == endRunes[i] {
				break
			}
		}
	}
	l.emit(Type)
	return textLexer
}

func linkLexer(l *Lexer) stageFunc {
	return linkMatcher(l, TokenLink, [2]rune{'[', '['}, [2]rune{']', ']'})
}

func imageLexer(l *Lexer) stageFunc {
	if r, err := l.Next(); err != nil || r != '!' {
		l.start = l.cur
		l.emit(TokenErr)
		return nil
	}
	return linkMatcher(l, TokenImage, [2]rune{'[', '('}, [2]rune{']', ')'})
}