package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	WebServer()
}

func PopulateStaticPages() *template.Template {
	result := template.New("templates")
	templatePaths := new([]string)

	basePath := "content"
	templateFolder, _ := os.Open(basePath)
	defer templateFolder.Close()
	templatePathsRaw, _ := templateFolder.Readdir(-1)

	for _, pathInfo := range templatePathsRaw {
		log.Println(pathInfo.Name())
		*templatePaths = append(*templatePaths, basePath+"/"+pathInfo.Name())
	}

	basePath = "includes"
	templateFolder, _ = os.Open(basePath)
	defer templateFolder.Close()
	templatePathsRaw, _ = templateFolder.Readdir(-1)

	for _, pathInfo := range templatePathsRaw {
		log.Println(pathInfo.Name())
		*templatePaths = append(*templatePaths, basePath+"/"+pathInfo.Name())
	}

	result.ParseFiles(*templatePaths...)
	return result
}

func ServeResource(w http.ResponseWriter, req *http.Request) {
	//path := "assets" + themeName + req.URL.Path
	path := "assets" + req.URL.Path
	var contentType string

	if strings.HasSuffix(path, ".css") {
		contentType = "text/css; charset=utf-8"
	} else if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript; charset=utf-8"
	} else if strings.HasSuffix(path, ".png") {
		contentType = "image/png; charset=utf-8"
	} else if strings.HasSuffix(path, ".jpg") {
		contentType = "image/jpg; charset=utf-8"
	} else if strings.HasSuffix(path, ".svg") {
		contentType = "image/svg+xml; charset=utf-8"
	} else {
		contentType = "text/plain; charset=utf-8"
	}

	log.Println(path)
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		w.Header().Add("Content-Type", contentType)
		br := bufio.NewReader(f)
		br.WriteTo(w)
	} else {
		w.WriteHeader(404)
	}
}

type defaultContext struct {
	Title       string
	Section     string
	Year        int
	ErrorMsgs   string
	SuccessMsgs string
}

var staticPages = PopulateStaticPages()

func ServeContent(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]
	t := time.Now()

	if pageAlias == "" {
		pageAlias = "index.html"
	}

	context := defaultContext{}
	context.Title = strings.Title(pageAlias)
	context.Section = pageAlias
	context.Year = t.Year()
	context.ErrorMsgs = ""
	context.SuccessMsgs = ""

	staticPage := staticPages.Lookup(pageAlias)
	if staticPage == nil {
		context.Title = strings.Title("Whoops!")
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, context)
}

func updateHTML(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]

	staticPages = PopulateStaticPages()
	staticPage := staticPages.Lookup(pageAlias)
	if staticPage == nil {
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, nil)
}

// WebServer - Server Function w/ Routers
func WebServer() {
	route := mux.NewRouter()

	route.HandleFunc("/", ServeContent)
	route.HandleFunc("/{pageAlias}", ServeContent)          // Dynamic URL
	route.HandleFunc("/updateHTML/{pageAlias}", updateHTML) //For updating new html files online

	http.HandleFunc("/css/", ServeResource)
	http.HandleFunc("/js/", ServeResource)
	http.HandleFunc("/images/", ServeResource)

	//route.HandleFunc("/book/{Name}/{pageAlias}", serveBuyBook)
	route.HandleFunc("/p/{pageAlias}", GetPopularBooks).Methods("GET")
	route.HandleFunc("/create/{pageAlias}", wannaCreateBook)
	route.HandleFunc("/create/books/{pageAlias}", createBook).Methods("POST")
	route.HandleFunc("/books/{pageAlias}", getAllBooks).Methods("GET")
	route.HandleFunc("/{Name}/book-name.html", getBookByName).Methods("GET")
	route.HandleFunc("/{Author}/book-author.html", getBookByAuthor).Methods("GET")

	http.Handle("/", route)

	portNumber := "8088"
	fmt.Println("------------------------------------------------------------")
	log.Println("Server started at http://localhost:" + portNumber)
	fmt.Println("------------------------------------------------------------")
	http.ListenAndServe(":"+portNumber, nil)

}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "alex"
	dbname   = "book_store"
)

type Book struct {
	Name        string
	Author      string
	Price       float32
	Description string
	ID          int
}

func Open() (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db, nil
}

func GetPopularBooks(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]

	PopularBooks := getPopularBooksFromDb()

	staticPage := staticPages.Lookup(pageAlias)

	if staticPage == nil {
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, PopularBooks)
}

func getPopularBooksFromDb() []Book {
	dbase, err := Open()
	if err != nil {
		panic(err)
	}
	defer dbase.Close()

	popularBooks := []Book{}

	rows, err := dbase.Query("select name, author, price from \"books\" limit 3")
	if err != nil {
		panic(err)
	}

	t := Book{}
	for rows.Next() {
		err := rows.Scan(&t.Name, &t.Author, &t.Price)
		if err != nil {
			fmt.Println(err)
			continue
		}
		popularBooks = append(popularBooks, t)
	}
	return popularBooks
}

func getAllBooks(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]

	Books := getAllBooksFromDb()
	staticPage := staticPages.Lookup(pageAlias)
	if staticPage == nil {
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, Books)
}

func getAllBooksFromDb() []Book {
	dbase, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	Books := []Book{}
	rows, err := dbase.Query("select name, author, price from books")
	if err != nil {
		log.Fatal(err)
	}
	b := Book{}
	for rows.Next() {
		err := rows.Scan(&b.Name, &b.Author, &b.Price)
		if err != nil {
			fmt.Println(err)
		}
		Books = append(Books, b)
	}
	return Books
}

func wannaCreateBook(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]
	staticPage := staticPages.Lookup(pageAlias)
	if staticPage == nil {
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, nil)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]
	price, _ := strconv.ParseFloat(r.FormValue("Price"), 32)
	book := Book{
		Name:        r.FormValue("Name"),
		Author:      r.FormValue("Author"),
		Price:       float32(price),
		Description: r.FormValue("Description"),
	}
	createdBook := createBookInDb(book)
	staticPage := staticPages.Lookup(pageAlias)
	staticPage.Execute(w, createdBook)
}

func createBookInDb(book Book) Book {
	dbase, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	_, err = dbase.Exec("insert into books (name, author, price, description) values($1, $2, $3, $4);", book.Name, book.Author, book.Price, book.Description)
	if err != nil {
		log.Fatal(err)
	}
	row := dbase.QueryRow("select name, author, price from books where id=(select max(id) from books)")
	b := Book{}
	err = row.Scan(&b.Name, &b.Author, &b.Price)
	fmt.Println("Scan success")
	if err != nil {
		fmt.Println(err)
	}
	return b
}

func getBookByName(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	Name := urlParams["Name"]

	Book := getBookByNameFromDb(Name)

	staticPage := staticPages.Lookup("book-name.html")

	if staticPage == nil {
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, Book)
}

func getBookByNameFromDb(name string) Book {
	dbase, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	row := dbase.QueryRow("select * from books where name=$1", name)
	b := Book{}
	err = row.Scan(&b.Name, &b.Author, &b.Price, &b.Description, &b.ID)
	fmt.Println("Scan success")
	return b
}

func getBookByAuthor(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	Name := urlParams["Author"]

	Book := getBookByAuthorFromDb(Name)

	staticPage := staticPages.Lookup("book-author.html")

	if staticPage == nil {
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, Book)
}

func getBookByAuthorFromDb(author string) Book {
	dbase, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	row := dbase.QueryRow("select * from books where author=$1 ", author)
	b := Book{}
	err = row.Scan(&b.Name, &b.Author, &b.Price, &b.Description, &b.ID)
	fmt.Println("Scan success")
	return b
}
