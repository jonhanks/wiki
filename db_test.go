package main

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
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
}
