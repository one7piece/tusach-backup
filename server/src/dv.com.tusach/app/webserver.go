package main

import (
	"dv.com.tusach/maker"
	"dv.com.tusach/util"
	"errors"
	"flag"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var systemInfo maker.SystemInfo
var users []maker.User
var books []maker.Book
var scripts []maker.ParserScript
var channels map[int]chan string

func main() {
	loadData()

	log.Println("Starting GO web server at " + maker.GetConfiguration().ServerBindAddress +
		":" + strconv.Itoa(maker.GetConfiguration().ServerBindPort) +
		", server path: " + maker.GetConfiguration().ServerPath)

	// create channels map
	channels = make(map[int]chan string)

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/systeminfo", GetSystemInfo},
		&rest.Route{"GET", "/books", GetBooks},
		&rest.Route{"GET", "/book/:id", GetBook},
		&rest.Route{"POST", "/book", CreateBook},
		&rest.Route{"PUT", "/book:cmd", UpdateBook},
	)
	if err != nil {
		log.Fatal("GOWebServer - ERROR! ", err)
		os.Exit(1)
	}

	api.SetApp(router)
	// api handler
	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))
	// static file handler
	//http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(configuration.ServerPath))))
	http.Handle("/", http.FileServer(http.Dir(maker.GetConfiguration().ServerPath)))

	log.Println("GOWebServer started successfully")

	if err := http.ListenAndServe(maker.GetConfiguration().ServerBindAddress+":"+
		strconv.Itoa(maker.GetConfiguration().ServerBindPort), nil); err != nil {
		log.Fatal("GOWebServer - ERROR! ", err)
		os.Exit(1)
	}
}

func GetSystemInfo(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(systemInfo)
}

func GetBooks(w rest.ResponseWriter, r *rest.Request) {
	w.WriteJson(books)
}

func GetBook(w rest.ResponseWriter, r *rest.Request) {
	idstr := r.PathParam("id")
	log.Printf("GetBook: %s", idstr)
	id, err := strconv.Atoi(idstr)
	if err != nil {
		rest.Error(w, "Invalid book ID, value must be an integer.", 400)
		return
	}

	var result maker.Book
	for _, book := range books {
		if book.ID == id {
			result = book
			break
		}
	}
	w.WriteJson(result)
}

func UpdateBook(w rest.ResponseWriter, r *rest.Request) {
	op := r.PathParam("op")
	log.Printf("UpdateBook: %s", op)
	if op != "abort" && op != "resume" && op != "update" && op != "delete" {
		rest.Error(w, "Invalid op value: "+op, 400)
		return
	}
	updateBook := maker.Book{}
	err := r.DecodeJsonPayload(&updateBook)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	currentBook, err := maker.LoadBook(updateBook.ID)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c, present := channels[updateBook.ID]
	if present {
		// can only abort
		if op == "abort" {
			c <- "abort"
		} else {
			err = errors.New("Book is currently in progress.")
		}
	} else {
		switch op {
		case "update":
			_, err = maker.SaveBook(updateBook)

		case "resume":
			currentBook.Status = maker.STATUS_WORKING
			bookId, err := maker.SaveBook(currentBook)
			if err == nil {
				// find parser
				parser := maker.GetParserName(currentBook.CurrentPageUrl)
				if parser == "" {
					err = errors.New("No parser found for url: " + currentBook.CurrentPageUrl)
				} else {
					// schedule goroutine to create book
					c := make(chan string)
					channels[currentBook.ID] = c
					go maker.CreateBook(c, currentBook, parser)
				}
			}

		case "delete":
			maker.DeleteBook(updateBook.ID)
		}
	}

	message := "OK"
	if err != nil {
		message = "ERROR: " + err.Error()
	}

	w.WriteJson()
}

func CreateBook(w rest.ResponseWriter, r *rest.Request) {
	newBook := maker.Book{}
	err := r.DecodeJsonPayload(&newBook)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// validate
	if newBook.Title == "" {
		rest.Error(w, "Missing book title", 400)
		return
	}
	if newBook.StartPageUrl == "" {
		rest.Error(w, "Missing start page URL", 400)
		return
	}

	// prevent too many concurrent books creation
	numActive := 0
	for _, book := range books {
		if book.Status == maker.STATUS_WORKING {
			numActive++
		}
	}
	if numActive >= maker.GetConfiguration().MaxActionBooks {
		rest.Error(w, "Too many concurrent books in progress", 400)
		return
	}

	parser := GetParserName(book.StartPageUrl)
	if parser == "" {
		rest.Error(w, "No parser found for url: "+book.StartPageUrl, 400)
		return
	}

	newBook.Status = maker.STATUS_WORKING
	bookId, err := maker.SaveBook(newBook)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newBook.ID = bookId

	// schedule goroutine to create book
	c := make(chan string)
	channels[bookId] = c
	go maker.CreateBook(c, newBook, parser)

	w.WriteJson(newBook)
}

func loadData() {
	var configFile string
	flag.StringVar(&configFile, "configFile", "", "Configuration file name")
	flag.Parse()

	if configFile == "" {
		log.Fatal("GOWebserver - ERROR! missing parameter: configFile")
		os.Exit(1)
	}
	now := time.Now()

	// load configuration
	util.LoadConfig(configFile)

	// init system info
	systemInfo = maker.SystemInfo{SystemUpTime: now, BookLastUpdateTime: now, ParserEditing: false}
	err := maker.SaveSystemInfo(systemInfo)
	if err != nil {
		panic("Error saving system info! " + err.Error())
	}

	// init users
	users, err = maker.LoadUsers()
	if err != nil {
		panic("Error loading users! " + err.Error())
	}
	if len(users) == 0 {
		adminUser := maker.User{Name: "admin", Password: "spidey", Role: "administrator"}
		err = maker.SaveUser(adminUser)
		if err != nil {
			panic("Error saving user! " + err.Error())
		}
		users = []maker.User{adminUser}
	}

	// load books
	books, err = maker.LoadBooks()
	if err != nil {
		panic("Error loading books! " + err.Error())
	}

	// init parser scripts
}
