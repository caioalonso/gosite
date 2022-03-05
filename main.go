package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

func readFile(fileName string) string {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

type Post struct {
	Filename    string
	Title       string
	Date        time.Time
	Markdown    string
	HTMLContent string
	HTML        string
	Aliases     []string
}

var postsAliases map[string]Post
var index string
var head string
var footer string
var fsHandle http.Handler

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	post := vars["post"]

	_, err := fmt.Fprintf(w, postsAliases["/posts/"+post].HTML)
	if err != nil {
		log.Fatal(err)
	}
}

func PostsWithDateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alias := "/" + vars["year"] + "/" + vars["month"] + "/" + vars["day"] + "/" + vars["post"]

	_, err := fmt.Fprintf(w, postsAliases[alias].HTML)
	if err != nil {
		log.Fatal(err)
	}
}

func catchAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		_, err := fmt.Fprintf(w, index)
		if err != nil {
			log.Fatal(err)
		}
	} else if strings.Contains(r.URL.Path, "/.") {
		w.WriteHeader(http.StatusNotFound)
	} else {
		fsHandle.ServeHTTP(w, r)
	}
}

func readPosts() (posts []Post) {
	err := filepath.Walk("posts", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			posts = append(posts, Post{Filename: path})
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	markdown := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)

	for i := range posts {
		posts[i].Markdown = readFile(posts[i].Filename)

		var buf bytes.Buffer
		context := parser.NewContext()
		if err := markdown.Convert([]byte(posts[i].Markdown), &buf, parser.WithContext(context)); err != nil {
			log.Fatal(err)
		}
		metadata := meta.Get(context)
		posts[i].Title = fmt.Sprintf("%v", metadata["title"])
		posts[i].HTMLContent = buf.String()
		layout := "2006-01-02"
		date, err := time.Parse(layout, fmt.Sprintf("%v", metadata["date"]))
		if err != nil {
			log.Fatal(err)
		}
		posts[i].Date = date
		posts[i].HTML = assemblePostPage(&posts[i])
		switch reflect.TypeOf(metadata["aliases"]).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(metadata["aliases"])
			for j := 0; j < s.Len(); j++ {
				posts[i].Aliases = append(posts[i].Aliases, fmt.Sprintf("%v", s.Index(j)))
			}
		}
	}

	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Date.Before(posts[j].Date)
	})

	return
}

func assemblePage(title, content string) string {
	headWithTitle := strings.Replace(head, "$TITLE", title, 1)
	return headWithTitle + content + footer
}

func assemblePostPage(post *Post) string {
	content := "<article>"
	content += "<h2>" + post.Title + "</h2>"
	content += fmt.Sprintf("<time datetime=%v>%v</time>", post.Date.Format("2006-01-02"), post.Date.Format("January 2, 2006"))
	content += post.HTMLContent
	content += "</article>"
	return assemblePage(post.Title, content)
}

func atomFeed(posts []Post) (feed string) {
	feed = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	feed += "<feed xmlns=\"http://www.w3.org/2005/Atom\">\n"
	feed += "<title>Caio Alonso</title>\n"
	feed += "<link href=\"https://caioalonso.com/feed.xml\" rel=\"self\" />\n"
	feed += "<link href=\"https://caioalonso.com\" />\n"
	feed += "<updated>" + posts[0].Date.Format("2006-01-02T15:04:05Z") + "</updated>\n"
	feed += "<id>https://caioalonso.com/</id>\n"
	feed += "<author>\n"
	feed += "<name>Caio Alonso</name>\n"
	feed += "</author>\n"
	for i := range posts {
		feed += "<entry>\n"
		feed += "<title>" + posts[i].Title + "</title>\n"
		feed += "<link href=\"https://caioalonso.com" + posts[i].Aliases[0] + "\"/>\n"
		feed += "<id>https://caioalonso.com" + posts[i].Aliases[0] + "</id>\n"
		feed += "<updated>" + posts[i].Date.Format("2006-01-02T15:04:05Z") + "</updated>\n"
		feed += "<summary>" + html.EscapeString(posts[i].HTMLContent) + "</summary>\n"
		feed += "</entry>\n"
	}
	feed += "</feed>\n"

	return
}

func main() {
	r := mux.NewRouter()

	head = readFile("head.html")
	footer = readFile("footer.html")
	learning := assemblePage("Learning", readFile("learning.html"))
	posts := readPosts()

	postsList := "<ul class=posts>"
	for _, post := range posts {
		postsList += fmt.Sprintf(
			"<li>\n<time datetime=%v>%v</time>\n<a href=%v>%v</a>\n</li>", post.Date.Format("2006-01-02"), post.Date.Format("January 2, 2006"), post.Aliases[0], post.Title)
	}
	postsList += "</ul>"
	index = assemblePage("Caio Alonso", postsList)

	atom := atomFeed(posts)

	postsAliases = make(map[string]Post)
	for _, post := range posts {
		for _, alias := range post.Aliases {
			postsAliases[alias] = post
		}
	}

	r.HandleFunc("/posts/{post}", PostsHandler)
	r.HandleFunc("/posts/{post}/", PostsHandler)
	// Legacy URLs I want to maintain
	r.HandleFunc("/{year}/{month}/{day}/{post}", PostsWithDateHandler)

	r.HandleFunc("/learning", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, learning)
		if err != nil {
			log.Fatal(err)
		}
	})

	r.HandleFunc("/index.xml", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, atom)
		if err != nil {
			log.Fatal(err)
		}
	})

	r.HandleFunc("/feed.xml", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, atom)
		if err != nil {
			log.Fatal(err)
		}
	})

	r.HandleFunc("/learning/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, learning)
		if err != nil {
			log.Fatal(err)
		}
	})

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	staticDir := filepath.Join(homeDir, "public_html")
	fsHandle = http.FileServer(http.Dir(staticDir))

	r.PathPrefix("/").HandlerFunc(catchAllHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
