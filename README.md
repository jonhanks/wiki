wiki
====

A simple Go  based wiki.  This is being done because each OS X upgrade breaks my moin moin install and to provide a test project for using the Martini project.

Basic goals:

* Simple markup (initially markdown)
* Simple backend(s)
    * Memory based
    * Flat file possibly as this is readable when the system dies
    * Sqlite - not yet implemented

Current Status:

* Basic Wiki features work
    * Create/edit pages
    * Automatically setup links between pages

Todo

* Revision/History
* Users
* Modular authentication
* Attachments & images
* sql backend
* categories/tags/...
* typing in some scripting/templating for use in pages ?
* make some decent page templates
* make a simple REST api
    * Currently it is a simple html and form based system, simple is good, it does break.  But REST api's are simple and open up possibilties.
