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
)

var (
	dbErr = errors.New("DB Error")

	NOT_FOUND = errors.New("Page not found")

	fdb_Page_re = regexp.MustCompile("^[0-9]{8}$")
)

type DB interface {
	PageExists(string) (bool, error)
	GetPage(string) ([]byte, error)
	SavePage(string, []byte) error
	ListPages() ([]string, error)
	CountPages() (int, error)
}

type memDB struct {
	lock  sync.Mutex
	pages map[string][]byte
}

type fileDB struct {
	lock sync.Mutex
	root string
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

func (fdb *fileDB) GetPage(key string) ([]byte, error) {
	fdb.lock.Lock()
	defer fdb.lock.Unlock()

	if !IsWikiWord(key) {
		return nil, dbErr
	}

	pagePath := fdb.pageDirName(key)

	fInfos, err := ioutil.ReadDir(pagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NOT_FOUND
		}
		return nil, err
	}

	maxRevision := getMaxFDBRevision(fInfos)
	name := fmt.Sprintf("%08d", maxRevision)
	return ioutil.ReadFile(path.Join(pagePath, name))
}

func (fdb *fileDB) SavePage(key string, value []byte) error {
	fdb.lock.Lock()
	defer fdb.lock.Unlock()

	if !IsWikiWord(key) {
		return dbErr
	}

	pagePath := fdb.pageDirName(key)

	err := os.MkdirAll(pagePath, fdb_Mode)
	if err != nil {
		return err
	}
	fInfos, err := ioutil.ReadDir(pagePath)
	if err != nil {
		return err
	}
	f, err := ioutil.TempFile(pagePath, "tmp")
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
	err = os.Rename(tmpName, path.Join(pagePath, newFName))
	return err
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

func newMemDB() DB {
	return &memDB{pages: make(map[string][]byte)}
}

func (mdb *memDB) PageExists(key string) (bool, error) {
	mdb.lock.Lock()
	defer mdb.lock.Unlock()

	if !IsWikiWord(key) {
		return false, dbErr
	}
	_, ok := mdb.pages[key]
	return ok, nil
}

func (mdb *memDB) GetPage(key string) ([]byte, error) {
	mdb.lock.Lock()
	defer mdb.lock.Unlock()

	if !IsWikiWord(key) {
		return nil, dbErr
	}
	val, ok := mdb.pages[key]
	if !ok {
		return nil, NOT_FOUND
	}
	return val, nil
}

func (mdb *memDB) SavePage(key string, value []byte) error {
	mdb.lock.Lock()
	defer mdb.lock.Unlock()

	if !IsWikiWord(key) {
		return dbErr
	}
	mdb.pages[key] = value
	return dbErr
}

func (mdb *memDB) ListPages() ([]string, error) {
	mdb.lock.Lock()
	defer mdb.lock.Unlock()

	cnt := len(mdb.pages)
	results := make([]string, 0, cnt)
	for key, _ := range mdb.pages {
		results = append(results, key)
	}
	return results, nil

}

func (mdb *memDB) CountPages() (int, error) {
	mdb.lock.Lock()
	defer mdb.lock.Unlock()

	return len(mdb.pages), nil
}
