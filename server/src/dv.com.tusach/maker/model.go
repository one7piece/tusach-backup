package maker

import (
	"time"
)

type SystemInfo struct {
	SystemUpTime       time.Time
	BookLastUpdateTime time.Time
	ParserEditing      bool
}

type User struct {
	Name          string
	Role          string
	Password      string
	LastLoginTime time.Time
}

type ParserScript struct {
	Name           string
	Code           string
	LastUpdateTime time.Time
}

const STATUS_NONE = ""
const STATUS_WORKING = "WORKING"
const STATUS_COMPLETED = "COMPLETED"
const STATUS_ERROR = "ERROR"
const STATUS_ABORTED = "ABORTED"

type Book struct {
	ID             int
	Title          string
	Author         string
	CreatedTime    time.Time
	CreatedBy      string
	Status         string
	BuildTimeSec   int
	StartPageUrl   string
	CurrentPageUrl string
	CurrentPageNo  int
	MaxNumPages    int
	LastUpdateTime time.Time
	ErrorMsg       string
	EpubCreated    bool
}

type Chapter struct {
	ChapterNo int
	BookId    int
	Title     string
	Html      []byte         `persist:"false"`
	images    []ChapterImage `persist:"false"`
}

// ByChapterNo implements sort.Interface for []Chapter on the ChapterNo field
type ByChapterNo []Chapter

func (c ByChapterNo) Len() int           { return len(c) }
func (c ByChapterNo) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByChapterNo) Less(i, j int) bool { return c[i].ChapterNo < c[j].ChapterNo }

type ChapterImage struct {
	fileName  string
	imageData []byte
}
