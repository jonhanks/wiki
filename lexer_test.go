package main

import (
	"bytes"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestLexer(t *testing.T) {
	test1 := []byte("A AbcDef123 [](AbcDef123) ![AbcDef123][AbcDef123]")

	Convey("Create a Lexer", t, func() {
		l, _ := NewLexer(test1)
		So(l, ShouldNotBeNil)
	})
}

func TestLexerNext(t *testing.T) {
	Convey("Create a lexer to test Next", t, func() {
		l, _ := NewLexer([]byte("abc"))
		dat, err := l.Next()
		So(dat, ShouldEqual, 'a')
		So(err, ShouldBeNil)
		dat, err = l.Next()
		So(dat, ShouldEqual, 'b')
		So(err, ShouldBeNil)
		dat, err = l.Next()
		So(dat, ShouldEqual, 'c')
		So(err, ShouldBeNil)
		dat, err = l.Next()
		So(err, ShouldNotBeNil)
		So(l.IsEOF(err), ShouldBeTrue)
	})
}

func TestLexerReverse(t *testing.T) {
	Convey("Create a lexer to test Reverse", t, func() {
		l, _ := NewLexer([]byte("abc"))
		l.Next()
		err := l.Reverse('a')
		So(err, ShouldBeNil)
		err = l.Reverse('a')
		So(err, ShouldNotBeNil)
	})
}

func TestLexerTextState(t *testing.T) {
	input1 := []byte("abc def")
	input2 := []byte("[][abc]")
	input2a := []byte("[](abc)")
	input2b := []byte("[]:")
	input3 := []byte("![](abc)")
	input3a := []byte("![][abc]")
	input4 := []byte("efg!abc")
	input5 := []byte("efg!abc!")

	input6 := []byte("WikiWord")
	input7 := []byte("abc WikiWord")
	input8 := []byte("abc WikiWord def")

	Convey("Create a lexer to test the text state", t, func() {
		l, ch := NewLexer(input1)
		nextState := textLexer(l)
		item := <-ch
		So(nextState, ShouldEqual, nil)
		So(item.Type, ShouldEqual, TokenText)
		So(bytes.Compare(item.Value, input1), ShouldEqual, 0)

		Convey("leading into a link", func() {
			l, ch = NewLexer(input2)
			nextState := textLexer(l)
			item := <-ch
			So(nextState, ShouldEqual, linkLexer)
			So(item.Type, ShouldEqual, TokenText)
			So(len(item.Value), ShouldEqual, 0)

			nextState = nextState(l)
			item = <-ch
			So(item.Type, ShouldEqual, TokenLink)
		})
		Convey("Using a regular (non-wiki link) should work as well", func() {
			l, ch = NewLexer(input2a)
			nextState := textLexer(l)
			item := <-ch
			So(nextState, ShouldEqual, linkLexer)
			So(item.Type, ShouldEqual, TokenText)
			So(len(item.Value), ShouldEqual, 0)

			nextState = nextState(l)
			item = <-ch
			So(item.Type, ShouldEqual, TokenLink)
		})
		Convey("Manually registering a link/reference should work too (as text)", func() {
			l, ch = NewLexer(input2b)
			nextState := textLexer(l)
			item := <-ch
			So(nextState, ShouldEqual, linkLexer)
			So(item.Type, ShouldEqual, TokenText)
			So(len(item.Value), ShouldEqual, 0)

			nextState = nextState(l)
			item = <-ch
			So(item.Type, ShouldEqual, TokenText)
		})
		Convey("leading into a image", func() {
			l, ch = NewLexer(input3)
			nextState := textLexer(l)
			item := <-ch
			So(nextState, ShouldEqual, imageLexer)
			So(item.Type, ShouldEqual, TokenText)
			So(len(item.Value), ShouldEqual, 0)

			nextState = nextState(l)
			item = <-ch
			So(item.Type, ShouldEqual, TokenImage)
		})
		Convey("leading into a image (referenced)", func() {
			l, ch = NewLexer(input3a)
			nextState := textLexer(l)
			item := <-ch
			So(nextState, ShouldEqual, imageLexer)
			So(item.Type, ShouldEqual, TokenText)
			So(len(item.Value), ShouldEqual, 0)

			nextState = nextState(l)
			item = <-ch
			So(item.Type, ShouldEqual, TokenImage)
		})
		Convey("testing text ending with an !", func() {
			l, ch = NewLexer(input4)
			nextState = textLexer(l)
			item = <-ch
			So(nextState, ShouldEqual, nil)
			So(item.Type, ShouldEqual, TokenText)
			So(bytes.Compare(item.Value, input4), ShouldEqual, 0)
		})
		Convey("testing text ending with an !", func() {
			l, ch = NewLexer(input5)
			nextState = textLexer(l)
			item = <-ch
			So(nextState, ShouldEqual, nil)
			So(item.Type, ShouldEqual, TokenText)
			So(bytes.Compare(item.Value, input5), ShouldEqual, 0)
		})
		Convey("testing text that is only a wiki word", func() {
			fmt.Println("666666666666666")
			l, ch = NewLexer(input6)
			nextState = textLexer(l)
			item = <-ch
			So(nextState, ShouldEqual, nil)
			So(item.Type, ShouldEqual, TokenWikiWord)
			So(bytes.Compare(item.Value, input6), ShouldEqual, 0)
		})
		Convey("testing text with a wiki word at the end", func() {
			fmt.Println("777777777777777")
			l, ch = NewLexer(input7)
			nextState = textLexer(l)
			So(nextState, ShouldEqual, nil)

			item = <-ch
			So(item.Type, ShouldEqual, TokenText)
			So(bytes.Compare(item.Value, []byte("abc ")), ShouldEqual, 0)
			fmt.Printf("Expecting 'abc' %s - %s\n", string(input7), string(item.Value))

			item = <-ch
			So(item.Type, ShouldEqual, TokenWikiWord)
			So(bytes.Compare(item.Value, []byte("WikiWord")), ShouldEqual, 0)
			fmt.Printf("Expecting 'WikiWord' %s - %s\n", string(input7), string(item.Value))
			fmt.Println("----------------")
		})
		Convey("testing text with a wiki word in the middle", func() {
			fmt.Println("888888888888888")
			l, ch = NewLexer(input8)
			nextState = textLexer(l)
			So(nextState, ShouldEqual, nil)

			item = <-ch
			So(item.Type, ShouldEqual, TokenText)
			So(bytes.Compare(item.Value, []byte("abc ")), ShouldEqual, 0)
			item = <-ch
			So(item.Type, ShouldEqual, TokenWikiWord)
			So(bytes.Compare(item.Value, []byte("WikiWord")), ShouldEqual, 0)
			item = <-ch
			So(item.Type, ShouldEqual, TokenText)
			So(bytes.Compare(item.Value, []byte(" def")), ShouldEqual, 0)
			fmt.Println("----------------")
		})
	})
}

func TestLexerLinkState(t *testing.T) {
	input1 := []byte("[][abcd]")

	Convey("Create a lexer to test the text state", t, func() {
		l, ch := NewLexer(input1)
		nextState := linkLexer(l)
		item := <-ch
		So(nextState, ShouldEqual, textLexer)
		So(item.Type, ShouldEqual, TokenLink)
		So(bytes.Compare(item.Value, input1), ShouldEqual, 0)
	})
}

func TestLexerImageState(t *testing.T) {
	input1 := []byte("![](abcd)")
	input2 := []byte("[](abcd)")
	input3 := []byte("[][abcd]")
	input4 := []byte("![(abcd)")
	input5 := []byte("![][abcd")

	Convey("Create a lexer to test the image state", t, func() {
		l, ch := NewLexer(input1)
		nextState := imageLexer(l)
		item := <-ch
		So(nextState, ShouldEqual, textLexer)
		So(item.Type, ShouldEqual, TokenImage)
		So(bytes.Compare(item.Value, input1), ShouldEqual, 0)

		Convey("Test without the leading !", func() {
			l, ch = NewLexer(input2)
			nextState := imageLexer(l)
			item := <-ch
			So(nextState, ShouldEqual, nil)
			So(item.Type, ShouldEqual, TokenErr)

			Convey("Test with a regular link", func() {
				l, ch = NewLexer(input3)
				nextState := imageLexer(l)
				item := <-ch
				So(nextState, ShouldEqual, nil)
				So(item.Type, ShouldEqual, TokenErr)

				Convey("Test with missing braces", func() {
					l, ch = NewLexer(input4)
					nextState := imageLexer(l)
					item := <-ch
					So(nextState, ShouldEqual, nil)
					So(item.Type, ShouldEqual, TokenErr)

					Convey("Test with missing closing brace", func() {
						l, ch = NewLexer(input5)
						nextState := imageLexer(l)
						item := <-ch
						So(nextState, ShouldEqual, nil)
						So(item.Type, ShouldEqual, TokenErr)
					})
				})
			})

		})
	})
}

func TestLexerRunLoop(t *testing.T) {
	Convey("Create a lexer to test the run loop", t, func() {
		l, ch := NewLexer([]byte("This is text with a WikiWord in it."))
		Convey("We should be able to run the lexer", func() {
			go l.Run()
			item := <-ch
			So(item.Type, ShouldEqual, TokenText)
			item = <-ch
			So(item.Type, ShouldEqual, TokenWikiWord)
			item = <-ch
			So(item.Type, ShouldEqual, TokenText)
			time.Sleep(2)
			item = <-ch
			So(item.Type, ShouldEqual, TokenEOF)
		})
	})
}
