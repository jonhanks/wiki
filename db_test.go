package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"testing"
)

func TestMemDBAdd(t *testing.T) {
	db := newMemDB()

	Convey("A memory Database starts out empty", t, func() {
		cnt, err := db.CountPages()
		So(err, ShouldBeNil)
		So(cnt, ShouldEqual, 0)

		dat, err := db.GetPage("PageOne")
		So(err, ShouldNotBeNil)
		So(dat, ShouldBeZeroValue)

		Convey("Adding an entry should increase the count to 1", func() {
			db.SavePage("PageOne", []byte("one"))

			cnt, err = db.CountPages()
			So(err, ShouldBeNil)
			So(cnt, ShouldEqual, 1)

			Convey("Adding a second entry should increase the count to 2", func() {
				db.SavePage("PageTwo", []byte("two"))

				cnt, err = db.CountPages()
				So(err, ShouldBeNil)
				So(cnt, ShouldEqual, 2)
				Convey("The list of pages can be checked", func() {
					lst, err := db.ListPages()

					So(err, ShouldBeNil)
					So(lst, ShouldContain, "PageOne")
					So(lst, ShouldContain, "PageTwo")

					Convey("We can also retreive pages", func() {
						dat, err = db.GetPage("PageOne")
						So(err, ShouldBeNil)
						_, err = db.GetPage("PageTwo")
						So(err, ShouldBeNil)
						_, err = db.GetPage("PageThree")
						So(err, ShouldNotBeNil)

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

								err = db.SavePage("pageOne", []byte("invalid"))
								So(err, ShouldNotBeNil)
							})
						})
					})
				})
			})
		})

	})
}

func TestFileDBAdd(t *testing.T) {
	tempPath, err := ioutil.TempDir("", "dbTest")
	if err != nil {
		t.Fatal("Unable to generate tempPath")
	}
	defer os.RemoveAll(tempPath)
	db, err := newFileDB(tempPath)
	if err != nil {
		t.Fatal("Unable to create file database")
	}

	Convey("A Database starts out empty", t, func() {
		cnt, err := db.CountPages()
		So(err, ShouldBeNil)
		So(cnt, ShouldEqual, 0)

		dat, err := db.GetPage("PageOne")
		So(err, ShouldNotBeNil)
		So(dat, ShouldBeZeroValue)

		Convey("Adding an entry should increase the count to 1", func() {
			db.SavePage("PageOne", []byte("one"))

			cnt, err = db.CountPages()
			So(err, ShouldBeNil)
			So(cnt, ShouldEqual, 1)

			Convey("Adding a second entry should increase the count to 2", func() {
				db.SavePage("PageTwo", []byte("two"))

				cnt, err = db.CountPages()
				So(err, ShouldBeNil)
				So(cnt, ShouldEqual, 2)
				Convey("The list of pages can be checked", func() {
					lst, err := db.ListPages()

					So(err, ShouldBeNil)
					So(lst, ShouldContain, "PageOne")
					So(lst, ShouldContain, "PageTwo")

					Convey("We can also retreive pages", func() {
						dat, err = db.GetPage("PageOne")
						So(err, ShouldBeNil)
						_, err = db.GetPage("PageTwo")
						So(err, ShouldBeNil)
						_, err = db.GetPage("PageThree")
						So(err, ShouldNotBeNil)

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

								err = db.SavePage("pageOne", []byte("invalid"))
								So(err, ShouldNotBeNil)
							})
						})
					})
				})
			})
		})

	})

}
