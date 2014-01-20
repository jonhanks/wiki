package main

import (
	"github.com/codegangsta/martini"
)

var wiki DB

func main() {
	var err error
	wiki, err = newFileDB("wiki_db")
	//wiki, err = newMemDB()
	if err != nil {
		panic(err.Error())
	}

	r := martini.Classic()
	r.Map(func() DB {
		return wiki
	})

	r.Get("/", ListPagesHandler)
	r.Get("/About/", AboutPageHandler)
	r.Get("/edit/:name/", ShowEditPageHandler)
	r.Post("/edit/:name/", EditPageHandler)
	r.Post("/edit/:name/attachment/", AddAttachmentHandler)
	r.Get("/:name/", PageHandler)
	r.Get("/:name/:attachment", AttachmentHandler)
	r.Run() //http.ListenAndServe(getListenAddress(), r)
}
