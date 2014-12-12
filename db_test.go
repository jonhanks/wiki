package main

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestMemDB(t *testing.T) {
	db, _ := newMemDB()
	doTestDB(t, db, "memory")
}

func TestFileDB(t *testing.T) {
	tempPath, err := ioutil.TempDir("", "dbTest")
	if err != nil {
		t.Fatal("Unable to generate tempPath")
	}
	defer os.RemoveAll(tempPath)
	db, err := newFileDB(tempPath)
	if err != nil {
		t.Fatal("Unable to create file database")
	}
	doTestDB(t, db, "file")
}

func TestMemDBAttachment(t *testing.T) {
	db, _ := newMemDB()
	doTestAttachments(t, db, "memory")
}

func TestFileDBAttachment(t *testing.T) {
	tempPath, err := ioutil.TempDir("", "dbTest")
	if err != nil {
		t.Fatal("Unable to generate tempPath")
	}
	defer os.RemoveAll(tempPath)
	db, err := newFileDB(tempPath)
	if err != nil {
		t.Fatal("Unable to create file database")
	}
	doTestAttachments(t, db, "file")
}

func doTestAttachments(t *testing.T, db DB, dbType string) {
	attachment1 := "This is a text attachment"
	attachment2 := "This is also a text attachment"
	Convey("Testing a "+dbType+" database, start by creating a page to work with", t, func() {
		page, err := db.GetPage("TestPage")
		So(err, ShouldBeNil)

		Convey("An uninitialized page should have 0 attachments", func() {
			count, err := page.CountAttachments()
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 0)

			list, err := page.ListAttachments()
			So(err, ShouldBeNil)
			So(len(list), ShouldEqual, 0)

			err = page.AddRevision([]byte("AbcDef"))
			So(err, ShouldBeNil)

			Convey("The page can have attachments, initially there will be none", func() {
				count, err = page.CountAttachments()
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)

				Convey("Attachments cannot have an invalid name", func() {
					err = page.AddAttachment(strings.NewReader(attachment1), "$attachment1.txt")
					So(err, ShouldNotBeNil)
					err = page.AddAttachment(strings.NewReader(attachment1), "")
					So(err, ShouldNotBeNil)

					Convey("Adding an attachment should bring the count to 1", func() {
						err = page.AddAttachment(strings.NewReader(attachment1), "attachment1.txt")
						So(err, ShouldBeNil)

						count, err = page.CountAttachments()
						So(err, ShouldBeNil)
						So(count, ShouldEqual, 1)

						Convey("However, duplicate names are overwrite attachments", func() {
							err = page.AddAttachment(strings.NewReader(attachment1), "attachment1.txt")
							So(err, ShouldBeNil)

							count, err = page.CountAttachments()
							So(err, ShouldBeNil)
							So(count, ShouldEqual, 1)

							Convey("Adding a second attachment should increase the count to 2", func() {
								err = page.AddAttachment(strings.NewReader(attachment2), "attachment2.txt")
								So(err, ShouldBeNil)

								count, err = page.CountAttachments()
								So(err, ShouldBeNil)
								So(count, ShouldEqual, 2)

								Convey("We can list attachments as well", func() {
									list, err = page.ListAttachments()
									So(err, ShouldBeNil)
									So(list, ShouldContain, "attachment1.txt")
									So(list, ShouldContain, "attachment2.txt")
									So(len(list), ShouldEqual, 2)

									Convey("We can retreive attachments and check their contents", func() {
										entry, err := page.GetAttachment("attachment1.txt")
										So(err, ShouldBeNil)
										So(entry.Name(), ShouldEqual, "attachment1.txt")

										rc, err := entry.Open()
										So(err, ShouldBeNil)
										defer rc.Close()

										var buf bytes.Buffer
										io.Copy(&buf, rc)
										So(bytes.Compare(buf.Bytes(), []byte(attachment1)), ShouldEqual, 0)

										entry, err = page.GetAttachment("attachment2.txt")
										So(err, ShouldBeNil)
										So(entry.Name(), ShouldEqual, "attachment2.txt")

										Convey("Requestion invalid names and non-existant attachments should fail", func() {
											entry, err = page.GetAttachment("att/a..chment2.txt")
											So(err, ShouldNotBeNil)
											entry, err = page.GetAttachment("attachment3.txt")
											So(err, ShouldNotBeNil)
										})
									})
								})
							})
						})
					})
				})
			})

		})
	})
}

func doTestDB(t *testing.T, db DB, dbType string) {
	var ONE_TEXTA = []byte{'o', 'n', 'e'}
	var ONE_TEXTB = []byte{'o', 'n', 'e', '!'}

	Convey("A "+dbType+" Database starts out empty", t, func() {
		cnt, err := db.CountPages()
		So(err, ShouldBeNil)
		So(cnt, ShouldEqual, 0)

		exists, err := db.PageExists("NonExistant")
		So(err, ShouldBeNil)
		So(exists, ShouldBeFalse)

		Convey("Adding an entry should increase the count to 1", func() {
			page, err := db.GetPage("PageOne")
			So(err, ShouldBeNil)
			So(page.Revisions(), ShouldEqual, NO_REVISIONS)

			err = page.AddRevision(ONE_TEXTA)
			So(err, ShouldBeNil)
			So(page.Revisions(), ShouldEqual, 1)

			cnt, err = db.CountPages()
			So(err, ShouldBeNil)
			So(cnt, ShouldEqual, 1)

			data, err := page.GetData(CURRENT_REVISION)
			So(err, ShouldBeNil)
			So(bytes.Compare(ONE_TEXTA, data), ShouldEqual, 0)

			err = page.AddRevision(ONE_TEXTB)
			So(err, ShouldBeNil)
			So(page.Revisions(), ShouldEqual, 2)

			data, err = page.GetData(CURRENT_REVISION)
			So(err, ShouldBeNil)
			So(bytes.Compare(ONE_TEXTB, data), ShouldEqual, 0)

			data, err = page.GetData(3)
			So(err, ShouldNotBeNil)

			data, err = page.GetData(0)
			So(err, ShouldBeNil)
			So(bytes.Compare(ONE_TEXTA, data), ShouldEqual, 0)

			Convey("Adding a second entry should increase the count to 2", func() {
				page, err = db.GetPage("PageTwo")
				So(err, ShouldBeNil)

				err = page.AddRevision([]byte("two"))
				So(err, ShouldBeNil)
				So(page.Revisions(), ShouldEqual, 1)

				cnt, err = db.CountPages()
				So(err, ShouldBeNil)
				So(cnt, ShouldEqual, 2)
				Convey("The list of pages can be checked", func() {
					lst, err := db.ListPages()

					So(err, ShouldBeNil)
					So(lst, ShouldContain, "PageOne")
					So(lst, ShouldContain, "PageTwo")

					Convey("We can also retreive pages", func() {
						page, err = db.GetPage("PageOne")
						So(err, ShouldBeNil)
						So(page.Revisions(), ShouldBeGreaterThan, 0)
						_, err = db.GetPage("PageTwo")
						So(err, ShouldBeNil)
						page, err = db.GetPage("PageThree")
						So(err, ShouldBeNil)
						So(page.Revisions(), ShouldEqual, NO_REVISIONS)

						Convey("The pages should know what their name is", func() {
							So(page.Name(), ShouldEqual, "PageThree")

							Convey("We can test existance as well", func() {
								dat, err := db.PageExists("PageOne")
								So(err, ShouldBeNil)
								So(dat, ShouldBeTrue)
								dat, err = db.PageExists("PageTwo")
								So(err, ShouldBeNil)
								So(dat, ShouldBeTrue)
								dat, err = db.PageExists("PageThree")
								So(err, ShouldBeNil)
								So(dat, ShouldBeFalse)

								Convey("Page names that are not WikiWords should fail", func() {
									_, err = db.PageExists("pageOne")
									So(err, ShouldNotBeNil)
									_, err = db.GetPage("pageOne")
									So(err, ShouldNotBeNil)

									_, err = db.GetPage("pageOne")
									So(err, ShouldNotBeNil)
								})
							})
						})
					})
				})
			})
		})

	})
}
