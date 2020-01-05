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
	"sync"
	"time"

	"github.com/gorilla/sessions"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	WebServer()
}

func PopulateStaticPages() *template.Template {
	result := template.New("templates")
	templatePaths := new([]string)

	wg := sync.WaitGroup{}
	wg.Add(2)

	f := func(basePath string) {
		templateFolder, _ := os.Open(basePath)
		defer templateFolder.Close()
		templatePathsRaw, _ := templateFolder.Readdir(-1)
		for _, pathInfo := range templatePathsRaw {
			log.Println(pathInfo.Name())
			*templatePaths = append(*templatePaths, basePath+"/"+pathInfo.Name())
		}
		wg.Done()
	}

	go f("content")
	go f("includes")
	wg.Wait()

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
	Cookie      interface{}
}

var staticPages = PopulateStaticPages()
var store = sessions.NewCookieStore([]byte("secret"))

func ServeContent(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]
	t := time.Now()

	if pageAlias == "" {
		pageAlias = "index.html"
	}

	session, _ := store.Get(r, "session")
	session.Save(r, w)

	myContext := defaultContext{}
	myContext.Title = strings.Title(pageAlias)
	myContext.Section = pageAlias
	myContext.Year = t.Year()
	myContext.ErrorMsgs = ""
	myContext.SuccessMsgs = ""
	myContext.Cookie = session.Values["login"]
	fmt.Println(myContext.Cookie)

	staticPage := staticPages.Lookup(pageAlias)
	if staticPage == nil {
		myContext.Title = strings.Title("Whoops!")
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, myContext)
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
	route.HandleFunc("/users/{pageAlias}", wannaCreateUser)
	route.HandleFunc("/users/create/{pageAlias}", createUser)

	route.HandleFunc("/p/{pageAlias}", GetPopularBooks).Methods("GET")

	route.HandleFunc("/create/{pageAlias}", wannaCreateBook).Methods("GET")
	route.HandleFunc("/create/books/{pageAlias}", createBook).Methods("POST")
	route.HandleFunc("/login/{pageAlias}", wannaLogin).Methods("GET")
	route.HandleFunc("/login/{pageAlias}", Login).Methods("POST")
	route.HandleFunc("/logout/{pageAlias}", Logout).Methods("GET")

	private := route.PathPrefix("/private").Subrouter()
	private.Use(authenticateUser)

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

type User struct {
	Email             string
	EncryptedPassword string
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

func wannaLogin(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]
	staticPage := staticPages.Lookup(pageAlias)
	if staticPage == nil {
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, nil)
}

func Login(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]
	user := User{
		Email:             r.FormValue("login"),
		EncryptedPassword: r.FormValue("password"),
	}

	if ok := loginCheck(user); ok {
		session, _ := store.Get(r, "session")
		session.Values["login"] = user.Email
		session.Save(r, w)
		staticPage := staticPages.Lookup(pageAlias)
		staticPage.Execute(w, user.Email)
	} else {
		http.Redirect(w, r, "/404.html", 404)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values["login"] = nil
	session.Save(r, w)
	pageAlias := mux.Vars(r)["pageAlias"]
	http.Redirect(w, r, "/"+pageAlias, http.StatusFound)
}

func authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session")
		if err != nil {
			http.Redirect(w, r, "/404.html", 404)
			return
		}

		_, ok := session.Values["login"]
		if !ok {
			http.Redirect(w, r, "/404.html", 404)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loginCheck(user User) bool {
	dbase, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	row := dbase.QueryRow("select encrypted_password from users where login=$1", user.Email)
	var realPassword string
	err = row.Scan(&realPassword)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(realPassword)
	return comparePasswords(realPassword, []byte(user.EncryptedPassword))
}

func comparePasswords(Pwd string, plainPwd []byte) bool {
	byteHash := []byte(Pwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func wannaCreateUser(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]
	staticPage := staticPages.Lookup(pageAlias)
	if staticPage == nil {
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}

	staticPage.Execute(w, nil)
}

func encryptString(s string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return ""
	}

	return string(b)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	pageAlias := urlParams["pageAlias"]
	user := User{
		Email:             r.FormValue("login"),
		EncryptedPassword: encryptString(r.FormValue("password")),
	}

	createdUser := createUserInDb(user)

	staticPage := staticPages.Lookup(pageAlias)
	staticPage.Execute(w, createdUser.Email)
}

func createUserInDb(user User) User {
	dbase, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	_, err = dbase.Exec("insert into users (login, encrypted_password) values ($1, $2);", user.Email, user.EncryptedPassword)
	if err != nil {
		log.Fatal(err)
	}
	row := dbase.QueryRow("select login from users where login=$1", user.Email)
	u := User{}
	err = row.Scan(&u.Email)
	fmt.Println("Scan success")
	if err != nil {
		fmt.Println(err)
	}
	return u
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
