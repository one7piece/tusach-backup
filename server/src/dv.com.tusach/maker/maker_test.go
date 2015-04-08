package maker

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestConfiguration(t *testing.T) {
	LoadConfig("/home/dvan/vshared/dev/tusach/tusach-config.json")
}

func TestPersistence(t *testing.T) {
	initDB()
	defer func() {
		if err := recover(); err != nil {
			closeDB()
			log.Printf("Recover from panic: %s\n", err)
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			log.Printf("Stack of %d bytes: %s\n", count, trace)
		}
	}()

	LoadSystemInfo()
	info := SystemInfo{SystemUpTime: time.Now(), BookLastUpdateTime: time.Now(), ParserEditing: false}
	SaveSystemInfo(info)
	LoadSystemInfo()

	book := Book{Title: "Gia Thien", Author: "Than Dong"}
	bookId, err := SaveBook(book)
	if err != nil {
		log.Println("Error saving book.", err)
	} else {
		book.ID = bookId
		log.Println("Saved book ID: ", book.ID)
	}
	LoadBooks()

	chapter1 := Chapter{BookId: book.ID, ChapterNo: 1, Title: "Chuong 1", Html: []byte("<html></html>")}
	err = SaveChapter(chapter1)
	if err != nil {
		log.Println("Error saving chapter.", err)
	} else {
		LoadChapters(0)
	}

	DeleteBook(book.ID)

	LoadBooks()
	LoadChapters(0)

	LoadUsers()
	user := User{Name: "Dung Van", Role: "admin", Password: "pwd", LastLoginTime: time.Now()}
	SaveUser(user)
	DeleteUser("Dung Van")
	LoadUsers()
}

func xTestSqlite3(t *testing.T) {
	os.Remove(configuration.DBFilename)

	log.Println("opening database...")
	db, err := sql.Open("sqlite3", configuration.DBFilename)
	if err != nil {
		t.Errorf("failed to open databse foo.db.", err)
	}
	//defer db.Close()
	defer func() {
		if err := recover(); err != nil {
			db.Close()
			log.Printf("Recover from panic: %s\n", err)
			trace := make([]byte, 1024)
			count := runtime.Stack(trace, true)
			log.Printf("Stack of %d bytes: %s\n", count, trace)
		}
	}()

	log.Println("creating table foo...")
	stmt := `
		create table foo (id integer not null primary key, name text); 
		delete from foo;
	`
	_, err = db.Exec(stmt)
	if err != nil {
		t.Errorf("failed to create table foo.", err)
	}

	log.Println("start transaction...")
	tx, err := db.Begin()
	if err != nil {
		t.Errorf("failed to start transaction.", err)
	}
	pstmt, err := tx.Prepare("insert into foo(id, name) values(?,?)")
	if err != nil {
		t.Errorf("failed to prepare statement.", err)
	}
	defer pstmt.Close()
	log.Println("executing inserts...")
	for i := 0; i < 100; i++ {
		_, err = pstmt.Exec(i, fmt.Sprintf("Hello%03d", i))
		if err != nil {
			t.Errorf("failed to execute statement.", err)
		}
	}
	log.Println("end transaction...")
	tx.Commit()

	rows, err := db.Query("select id, name from foo")
	if err != nil {
		t.Errorf("failed to query select.", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		log.Printf("found row: %d, %s\n", id, name)
	}
	rows.Close()

	pstmt, err = db.Prepare("select name from foo where id = ?")
	if err != nil {
		t.Errorf("failed to prepare statement.", err)
	}
	defer pstmt.Close()

}
