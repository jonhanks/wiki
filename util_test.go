package main

import (
	"errors"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"strings"
	"testing"
	"time"
)

type testWriteCloser struct {
	failOnWrite  bool
	bytesWritten int
	closed       bool
}

func (twc *testWriteCloser) Write(value []byte) (int, error) {
	if twc.failOnWrite {
		return 0, errors.New("Some error")
	}
	twc.bytesWritten = twc.bytesWritten + len(value)
	return len(value), nil
}

func (twc *testWriteCloser) Close() error {
	if twc.closed {
		return errors.New("double close")
	}
	twc.closed = true
	return nil
}

type testFDBFileInfo struct {
	rev int
}

func (tInfo *testFDBFileInfo) Name() string {
	return fmt.Sprintf("%08d", tInfo.rev)
}
func (tInfo *testFDBFileInfo) Size() int64 { return 0 }
func (tInfo *testFDBFileInfo) Mode() os.FileMode {
	return os.FileMode(0750)
}
func (tInfo *testFDBFileInfo) ModTime() time.Time {
	return time.Now()
}
func (tInfo *testFDBFileInfo) IsDir() bool      { return false }
func (tInfo *testFDBFileInfo) Sys() interface{} { return nil }

func TestWikiWordRe(t *testing.T) {
	Convey("All WikiWords match a regexp", t, func() {
		So(IsWikiWord("abcd"), ShouldBeFalse)
		So(IsWikiWord("123"), ShouldBeFalse)
		So(IsWikiWord("123Abc"), ShouldBeFalse)
		So(IsWikiWord("abcDef"), ShouldBeFalse)
		So(IsWikiWord("Abc+d3"), ShouldBeFalse)
		So(IsWikiWord("Abc_Def"), ShouldBeTrue)
		So(IsWikiWord("AbcD_ef"), ShouldBeTrue)
		So(IsWikiWord("AbcD_123"), ShouldBeTrue)
		So(IsWikiWord("Abc"), ShouldBeFalse)
		So(IsWikiWord("Jonathan@Hanks"), ShouldBeFalse)
		So(IsWikiWord("FirstLast"), ShouldBeTrue)
		So(IsWikiWord("FirstLast123"), ShouldBeTrue)
		So(IsWikiWord("FirstLastOther"), ShouldBeTrue)
	})
}

func TestExtractWikiWords(t *testing.T) {
	test0 := []byte("There are no wiki words in\nthis piece of text.")
	test1 := []byte("There is only\nOneWikiWord in this text.")
	test2 := []byte("Emily pointing to the CD player - This is where the CDEFG goes!")
	test4 := []byte("OneWikiWord not a wiki word\nTwoWikiWord ThreeWikiWords\n FourWords")
	testMulti := []byte("WordWordOne\nWordWord\n WordOne\nWordWordOne WordWordOne")

	Convey("A string/[]byte may have multiple wikiwords in it", t, func() {
		words0 := ExtractWikiWords(test0)
		So(len(words0), ShouldEqual, 0)

		words1 := ExtractWikiWords(test1)
		So(len(words1), ShouldEqual, 1)

		words2 := ExtractWikiWords(test2)
		So(len(words2), ShouldEqual, 2)

		words4 := ExtractWikiWords(test4)
		So(len(words4), ShouldEqual, 4)

		wordsMulti := ExtractWikiWords(testMulti)
		So(len(wordsMulti), ShouldEqual, 3)
	})
}

func TestExtractAndExpandWikiWords(t *testing.T) {
	test0 := []byte("There are no wiki words in\nthis piece of text.")
	test1 := []byte("There is only\nOneWikiWord in this text.")
	test2 := []byte("Emily pointing to the CD player - This is where the CDEFG goes!")
	test4 := []byte("OneWikiWord not a wiki word\nTwoWikiWord ThreeWikiWords\n FourWords")
	testMulti := []byte("WordWordOne\nWordWord\n WordOne\nWordWordOne WordWordOne")

	expected0 := []byte("There are no wiki words in\nthis piece of text.")
	expected1 := []byte("There is only\n[OneWikiWord](/OneWikiWord/) in this text.")
	expected2 := []byte("Emily pointing to the [CD](/CD/) player - This is where the [CDEFG](/CDEFG/) goes!")
	expected4 := []byte("[OneWikiWord](/OneWikiWord/) not a wiki word\n[TwoWikiWord](/TwoWikiWord/) [ThreeWikiWords](/ThreeWikiWords/)\n [FourWords](/FourWords/)")
	expectedMulti := []byte("[WordWordOne](/WordWordOne/)\n[WordWord](/WordWord/)\n [WordOne](/WordOne/)\n[WordWordOne](/WordWordOne/) [WordWordOne](/WordWordOne/)")

	//word := string(wikiWord)
	//rawPage = bytes.Replace(rawPage, wikiWord, []byte("["+word+"](/"+word+"/)"), -1)
	Convey("We must expand WikiWords to markdown links", t, func() {
		fmt.Println("Test4")
		out4 := ExtrandAndExpandWikiWords(test4)
		So(string(out4), ShouldEqual, string(expected4))

		fmt.Println("TestMulti")
		outMulti := ExtrandAndExpandWikiWords(testMulti)
		So(string(outMulti), ShouldEqual, string(expectedMulti))

		fmt.Println("Test0")
		out0 := ExtrandAndExpandWikiWords(test0)
		So(string(out0), ShouldEqual, string(expected0))

		fmt.Println("Test1")
		out1 := ExtrandAndExpandWikiWords(test1)
		So(string(out1), ShouldEqual, string(expected1))

		fmt.Println("Test2")
		out2 := ExtrandAndExpandWikiWords(test2)
		So(string(out2), ShouldEqual, string(expected2))

	})
}

func TestWriteAndClose(t *testing.T) {
	Convey("writeAndClose should always call close no matter what happens", t, func() {
		twc1 := testWriteCloser{}
		err := writeAndClose(&twc1, []byte("1234"))
		So(err, ShouldBeNil)
		So(twc1.closed, ShouldBeTrue)
		twc2 := testWriteCloser{failOnWrite: true}
		err = writeAndClose(&twc2, []byte("1234"))
		So(err, ShouldNotBeNil)
		So(twc2.closed, ShouldBeTrue)
	})
}

func TestCopyAndClose(t *testing.T) {
	Convey("copyAndClose should always call close no matter what happens", t, func() {
		twc1 := testWriteCloser{}
		err := copyAndClose(&twc1, strings.NewReader("1234"))
		So(err, ShouldBeNil)
		So(twc1.closed, ShouldBeTrue)
		twc2 := testWriteCloser{failOnWrite: true}
		err = copyAndClose(&twc2, strings.NewReader("1234"))
		So(err, ShouldNotBeNil)
		So(twc2.closed, ShouldBeTrue)
	})
}

func TestGetMaxFDBRevision(t *testing.T) {
	Convey("getMaxFDBRevision should return the highest file revision number for a list of os.FileInfos, where the format of the filename is the regexp [0-9]{8}", t, func() {
		fInfos := make([]os.FileInfo, 0, 3)
		fInfos = append(fInfos, &testFDBFileInfo{rev: 0})
		fInfos = append(fInfos, &testFDBFileInfo{rev: 5})
		fInfos = append(fInfos, &testFDBFileInfo{rev: 3})
		So(getMaxFDBRevision(fInfos), ShouldEqual, 5)
		So(getMaxFDBRevision(make([]os.FileInfo, 0, 0)), ShouldEqual, -1)
	})
}

func TestGetCountFDBAttachments(t *testing.T) {
	Convey("getCountFDBAttachments should return the number of attachments found by matching names to the attachment regexp", t, func() {
		fInfos := make([]os.FileInfo, 0, 3)
		fInfos = append(fInfos, &testFDBFileInfo{rev: 0})
		fInfos = append(fInfos, &testFDBFileInfo{rev: 5})
		fInfos = append(fInfos, &testFDBFileInfo{rev: 3})
		So(getMaxFDBRevision(fInfos), ShouldEqual, 5)
		So(getMaxFDBRevision(make([]os.FileInfo, 0, 0)), ShouldEqual, -1)
	})
}

func TestGenerateRevisionSplit(t *testing.T) {
	tests := []struct {
		InCur, InMax, OutCur, OutMin, OutMax int
	}{
		{-1, 5, 5, 0, 5},
		{5, 5, 5, 0, 5},
		{-1, 10, 10, 5, 10},
		{10, 10, 10, 5, 10},
		{5, 20, 5, 0, 10},
		{-1, 0, 0, 0, 0},
	}

	for _, testVal := range tests {
		cur, min, max := generateRevisionSplit(testVal.InCur, testVal.InMax)
		if testVal.OutCur != cur || testVal.OutMin != min || testVal.OutMax != max {
			t.Errorf("inputs (%d, %d) outputs (%d, %d, %d) expecting (%d, %d, %d)", testVal.InCur, testVal.InMax, cur, min, max, testVal.OutCur, testVal.OutMin, testVal.OutMax)
		}
	}
}

func TestGenerateInt(t *testing.T) {
	var v int

	for v = range generateInt(0, 5) {
	}
	if v != 5 {
		t.Error("Expecting 5, got ", v)
	}
	for v = range generateInt(5, 0) {
	}
	if v != 0 {
		t.Error("Expecting 0, got ", v)
	}
}
