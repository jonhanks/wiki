package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sync"
)

const (
	fdb_Pages = "pages"

	fdb_Mode = os.ModeDir | 0750

	CURRENT_REVISION = -1
	NO_REVISIONS     = 0
)

var (
	dbErr = errors.New("DB Error")

	NOT_FOUND = errors.New("Page not found")

	fdb_Page_re = regexp.MustCompile("^[0-9]{8}$")
)

type Page interface {
	GetData(int) ([]byte, error)
	AddRevision([]byte) error
	Revisions() int
	//AddAttachment(io.Reader, string) error
	//ListAttachments() ([]string, error)
	//CountAttachments() (int, error)
	//GetAttachment(string)
}

// Generic interface into the database
type DB interface {
	PageExists(string) (bool, error) // given a page name query to see if it exists
	GetPage(string) (Page, error)    // retreive a page given the name, it will return the error NOT_FOUND if the page does not exist
	ListPages() ([]string, error)    // list the pages in the wiki
	CountPages() (int, error)        // return the number of pages in the wiki
}

// A simple memory based wiki database
type memDB struct {
	lock  sync.Mutex
	pages map[string]*memPage
}

// a page in the memory based wiki
type memPage struct {
	db        *memDB
	revisions [][]byte
}

// A simple file system backed wiki database
type fileDB struct {
	lock sync.Mutex
	root string
}

type filePage struct {
	path string
}

func newFileDB(root string) (DB, error) {
	err := os.MkdirAll(path.Join(root, fdb_Pages), fdb_Mode)
	if err != nil {
		return nil, err
	}
	return &fileDB{root: root}, nil
}

func (fdb *fileDB) pageDirName(key string) string {
	return path.Join(fdb.root, fdb_Pages, key)
}

func (fdb *fileDB) PageExists(key string) (bool, error) {
	fdb.lock.Lock()
	defer fdb.lock.Unlock()

	if !IsWikiWord(key) {
		return false, dbErr
	}

	fInfo, err := os.Stat(fdb.pageDirName(key))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if fInfo.IsDir() == false {
		return false, dbErr
	}
	return true, nil
}

func (fdb *fileDB) GetPage(key string) (Page, error) {
	fdb.lock.Lock()
	defer fdb.lock.Unlock()

	if !IsWikiWord(key) {
		return nil, dbErr
	}

	return Page(&filePage{path: fdb.pageDirName(key)}), nil
}

func (fdb *fileDB) ListPages() ([]string, error) {
	fdb.lock.Lock()
	defer fdb.lock.Unlock()

	fInfos, err := ioutil.ReadDir(path.Join(fdb.root, fdb_Pages))
	if err != nil {
		return nil, dbErr
	}
	results := make([]string, 0, len(fInfos))
	for _, info := range fInfos {
		name := path.Base(info.Name())
		if IsWikiWord(name) {
			results = append(results, name)
		}
	}
	return results, nil
}

func (fdb *fileDB) CountPages() (int, error) {
	fdb.lock.Lock()
	defer fdb.lock.Unlock()

	fInfos, err := ioutil.ReadDir(path.Join(fdb.root, fdb_Pages))
	if err != nil {
		return 0, err
	}
	return len(fInfos), nil
}

func (fpg *filePage) GetData(index int) ([]byte, error) {
	fInfos, err := ioutil.ReadDir(fpg.path)
	if err != nil {

		return nil, err
	}
	max := getMaxFDBRevision(fInfos)
	if max < 0 || index > max || (index < 0 && index != CURRENT_REVISION) {
		return nil, dbErr
	}
	if index == CURRENT_REVISION {
		index = max
	}
	fname := fmt.Sprintf("%08d", index)
	return ioutil.ReadFile(path.Join(fpg.path, fname))
}

func (fpg *filePage) AddRevision(value []byte) error {
	// this is a noop if it already exists
	err := os.MkdirAll(fpg.path, fdb_Mode)
	if err != nil {
		return err
	}
	fInfos, err := ioutil.ReadDir(fpg.path)
	if err != nil {
		return err
	}
	f, err := ioutil.TempFile(fpg.path, "tp_")
	if err != nil {
		return err
	}
	tmpName := f.Name()
	defer os.Remove(tmpName)

	if err := writeAndClose(f, value); err != nil {
		return err
	}

	maxRevision := getMaxFDBRevision(fInfos)

	newFName := fmt.Sprintf("%08d", maxRevision+1)
	err = os.Rename(tmpName, path.Join(fpg.path, newFName))
	return err
}

func (fpg *filePage) Revisions() int {
	if fInfos, err := ioutil.ReadDir(fpg.path); err == nil {
		return getMaxFDBRevision(fInfos) + 1
	}
	return NO_REVISIONS
}

func newMemDB() (DB, error) {
	return &memDB{pages: make(map[string]*memPage)}, nil
}

func (mdb *memDB) PageExists(key string) (bool, error) {
	mdb.lock.Lock()
	defer mdb.lock.Unlock()

	if !IsWikiWord(key) {
		return false, dbErr
	}
	page, ok := mdb.pages[key]
	if ok {
		return page.Revisions() != NO_REVISIONS, nil
	}
	return false, nil
}

func (mdb *memDB) GetPage(key string) (Page, error) {
	mdb.lock.Lock()
	defer mdb.lock.Unlock()

	if !IsWikiWord(key) {
		return nil, dbErr
	}
	val, ok := mdb.pages[key]
	if !ok {
		val = &memPage{db: mdb, revisions: make([][]byte, 0)}
		mdb.pages[key] = val
	}
	return Page(val), nil
}

func (mdb *memDB) ListPages() ([]string, error) {
	mdb.lock.Lock()
	defer mdb.lock.Unlock()

	cnt := len(mdb.pages)
	results := make([]string, 0, cnt)
	for key, page := range mdb.pages {
		if page.Revisions() > 0 {
			results = append(results, key)
		}
	}
	return results, nil

}

func (mdb *memDB) CountPages() (int, error) {
	mdb.lock.Lock()
	defer mdb.lock.Unlock()

	count := 0
	for _, page := range mdb.pages {
		if page.Revisions() > 0 {
			count++
		}
	}
	return count, nil
}

func (mp *memPage) GetData(index int) ([]byte, error) {
	max := len(mp.revisions)
	if max == 0 || index >= max || (index < 0 && index != CURRENT_REVISION) {
		return nil, dbErr
	}
	if index == CURRENT_REVISION {
		return mp.revisions[max-1], nil
	}
	return mp.revisions[index], nil
}

func (mp *memPage) AddRevision(value []byte) error {
	mp.revisions = append(mp.revisions, value)
	return nil
}

func (mp *memPage) Revisions() int {
	return len(mp.revisions)
}
