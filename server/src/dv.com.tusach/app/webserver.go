package main

import (
	"dv.com.tusach/maker"
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

func main() {
	loadData()

	log.Println("Starting GO web server at " + maker.GetConfiguration().ServerBindAddress +
		":" + strconv.Itoa(maker.GetConfiguration().ServerBindPort) +
		", server path: " + maker.GetConfiguration().ServerPath)

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		&rest.Route{"GET", "/systeminfo", GetSystemInfo},
		&rest.Route{"GET", "/books", GetBooks},
		&rest.Route{"GET", "/book/:id", GetBook},
		&rest.Route{"POST", "/book", CreateBook},
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

	bookId, err := maker.SaveBook(newBook)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newBook.ID = bookId

	// schedule

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
	maker.LoadConfig(configFile)

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
