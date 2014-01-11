package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sync"
)

const (
	fdb_Pages = "pages"

	fdb_attachment_prefix = "a_"

	fdb_Mode = os.ModeDir | 0750

	CURRENT_REVISION = -1
	NO_REVISIONS     = 0
)

var (
	dbErr = errors.New("DB Error")

	NOT_FOUND = errors.New("Page not found")

	attachment_re = regexp.MustCompile("^[0-9A-Za-z\\-\\_]*(\\.[0-9A-Za-z\\-\\_]+)?$")

	fdb_Page_re = regexp.MustCompile("^[0-9]{8}$")

	fdb_Attachment_re = regexp.MustCompile("^a_[0-9A-Za-z\\-\\_]*(\\.[0-9A-Za-z\\-\\_]+)?$")
)

type Attachment interface {
	Open() (io.ReadCloser, error)
	Name() string
}

type Page interface {
	GetData(int) ([]byte, error)
	AddRevision([]byte) error
	Revisions() int
	AddAttachment(io.Reader, string) error
	ListAttachments() ([]string, error)
	CountAttachments() (int, error)
	GetAttachment(string) (Attachment, error)
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
	lock        sync.Mutex
	db          *memDB
	revisions   [][]byte
	attachments map[string][]byte
}

type memAttachment struct {
	page *memPage
	key  string
}

// A simple file system backed wiki database
type fileDB struct {
	lock sync.Mutex
	root string
}

type filePage struct {
	path string
}

type fileAttachment struct {
	path string
	key  string
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

func (fpg *filePage) AddAttachment(data io.Reader, key string) error {
	if !attachment_re.MatchString(key) {
		return dbErr
	}
	f, err := ioutil.TempFile(fpg.path, "ta_")
	if err != nil {
		return err
	}
	tmpName := f.Name()
	defer os.Remove(f.Name())

	if err := copyAndClose(f, data); err != nil {
		return err
	}

	return os.Rename(tmpName, path.Join(fpg.path, "a_"+key))
}

func (fpg *filePage) ListAttachments() ([]string, error) {
	if fInfos, err := ioutil.ReadDir(fpg.path); err == nil {
		count := getCountFDBAttachments(fInfos)
		results := make([]string, 0, count)
		for _, info := range fInfos {
			if fdb_Attachment_re.MatchString(info.Name()) {
				results = append(results, info.Name()[2:])
			}
		}
		return results, nil
	} else {
		return nil, err
	}
}

func (fpg *filePage) CountAttachments() (int, error) {
	if fInfos, err := ioutil.ReadDir(fpg.path); err == nil {
		return getCountFDBAttachments(fInfos), nil
	} else {
		return 0, err
	}
}

func (fpg *filePage) GetAttachment(key string) (Attachment, error) {
	if !attachment_re.MatchString(key) {
		return nil, dbErr
	}
	path := path.Join(fpg.path, "a_"+key)
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &fileAttachment{path: path, key: key}, nil
}

func (fa *fileAttachment) Name() string {
	return fa.key
}

func (fa *fileAttachment) Open() (io.ReadCloser, error) {
	return os.Open(fa.path)
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
		val = &memPage{db: mdb, revisions: make([][]byte, 0), attachments: make(map[string][]byte)}
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
	mp.lock.Lock()
	defer mp.lock.Unlock()

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
	mp.lock.Lock()
	defer mp.lock.Unlock()

	mp.revisions = append(mp.revisions, value)
	return nil
}

func (mp *memPage) Revisions() int {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	return len(mp.revisions)
}

func (mp *memPage) AddAttachment(data io.Reader, key string) error {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	if !attachment_re.MatchString(key) {
		return dbErr
	}
	//if _, ok := mp.attachments[key]; ok {
	//	return dbErr
	//}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, data); err != nil {
		return err
	}
	mp.attachments[key] = buf.Bytes()
	return nil
}

func (mp *memPage) ListAttachments() ([]string, error) {
	count, err := mp.CountAttachments()
	if err != nil {
		return nil, err
	}
	mp.lock.Lock()
	defer mp.lock.Unlock()

	results := make([]string, 0, count)
	for key, _ := range mp.attachments {
		results = append(results, key)
	}
	return results, nil
}

func (mp *memPage) CountAttachments() (int, error) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	return len(mp.attachments), nil
}

func (mp *memPage) GetAttachment(key string) (Attachment, error) {
	mp.lock.Lock()
	defer mp.lock.Unlock()

	if _, ok := mp.attachments[key]; !ok {
		return nil, dbErr
	}
	return &memAttachment{page: mp, key: key}, nil
}

func (ma *memAttachment) Name() string {
	return ma.key
}

func (ma *memAttachment) Open() (io.ReadCloser, error) {
	ma.page.lock.Lock()
	defer ma.page.lock.Unlock()

	return ioutil.NopCloser(bytes.NewReader(ma.page.attachments[ma.key])), nil
}
