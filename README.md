wiki
====

A simple Go  based wiki.  This is being done because each OS X upgrade breaks my apache (and thus my moin moin install) and to provide a test project for using some go toolkits and libraries (Gorilla web toolkit) and GoConvey).

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
    * Pages have a history
    * Old page revisions can be viewed
	* Attachments and basic image support works    

* Ideas being tested
    * Using middleware as a data/compute pipeline
        * This should make the handlers very minimal and make it easier to do a REST api and a standalone non-js interface as well.

Todo

* Users
* Modular authentication
* LMDB/BoltDB backend
* categories/tags/...
* typing in some scripting/templating for use in pages ?
* make some decent page templates
* make a simple REST api
    * Currently it is a simple html and form based system, simple is good, it doesn't break.  But REST api's are simple and open up possibilties.
