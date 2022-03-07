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
	Comments    []Comment
}

type Comment struct {
	Name string
	Body string
	Date time.Time
}

var postsAliases map[string]*Post
var index string
var head string
var footer string
var fsHandle http.Handler
var markdown goldmark.Markdown

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

func NewCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Reject if the request is not a POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	post := postsAliases["/posts/"+vars["post"]]

	name := r.FormValue("name")
	body := r.FormValue("body")

	if body == "" || name == "" {
		http.Error(w, "Missing name or body", http.StatusBadRequest)
		return
	}

	comment := Comment{
		Name: name,
		Body: body,
		Date: time.Now(),
	}

	post.Comments = append(post.Comments, comment)
	post.HTML = assemblePostPage(post)
	http.Redirect(w, r, post.Aliases[0], http.StatusSeeOther)
	return
}

func NewCommentWithDateHandler(w http.ResponseWriter, r *http.Request) {
	// Reject if the request is not a POST
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	postAlias := "/" + vars["year"] + "/" + vars["month"] + "/" + vars["day"] + "/" + vars["post"]
	post := postsAliases[postAlias]

	name := r.FormValue("name")
	body := r.FormValue("body")

	if body == "" || name == "" {
		http.Error(w, "Missing name or body", http.StatusBadRequest)
		return
	}

	comment := Comment{
		Name: r.FormValue("name"),
		Body: r.FormValue("body"),
		Date: time.Now(),
	}

	post.Comments = append(post.Comments, comment)
	post.HTML = assemblePostPage(post)
	http.Redirect(w, r, post.Aliases[0], http.StatusSeeOther)
	return
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

func parseMarkdown(markdownContent string) (HTML string, metadata map[string]interface{}) {
	var buf bytes.Buffer
	context := parser.NewContext()
	if err := markdown.Convert([]byte(markdownContent), &buf, parser.WithContext(context)); err != nil {
		log.Fatal(err)
	}
	metadata = meta.Get(context)
	HTML = buf.String()
	return
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

	for i := range posts {
		posts[i].Markdown = readFile(posts[i].Filename)
		var metadata map[string]interface{}
		posts[i].HTMLContent, metadata = parseMarkdown(posts[i].Markdown)
		posts[i].Title = fmt.Sprintf("%v", metadata["title"])
		layout := "2006-01-02"
		date, err := time.Parse(layout, fmt.Sprintf("%v", metadata["date"]))
		if err != nil {
			log.Fatal(err)
		}
		posts[i].Date = date
		switch reflect.TypeOf(metadata["aliases"]).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(metadata["aliases"])
			for j := 0; j < s.Len(); j++ {
				posts[i].Aliases = append(posts[i].Aliases, fmt.Sprintf("%v", s.Index(j)))
			}
		}
		posts[i].HTML = assemblePostPage(&posts[i])
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

func assembleGenericPage(title, content string) string {
	content = "<article><h2>" + title + "</h2>" + content + "</article>"
	return assemblePage(title, content)
}

func assemblePostPage(post *Post) string {
	content := "<article>"
	content += "<h2>" + post.Title + "</h2>"
	content += fmt.Sprintf("<time datetime=%v>%v</time>", post.Date.Format("2006-01-02"), post.Date.Format("January 2, 2006"))
	content += post.HTMLContent
	content += "</article>"
	content += "<h2>Comments</h2>"
	content += "<div id=comments>"
	for i, comment := range post.Comments {
		content += "<div class=comment>"
		content += fmt.Sprintf("<p><strong>#%v %v</strong> <time datetime=%v>%v</time></p>", i+1, comment.Name, comment.Date.Format("2006-01-02"), comment.Date.Format("January 2, 2006"))
		content += fmt.Sprintf("<p>%v</p>", comment.Body)
		content += "</div>"
	}
	content += "</div>"
	content += "<form action=\"" + post.Aliases[0] + "/comment\" method=post>"
	content += "<input type=text name=name placeholder=Name><br>"
	content += "<textarea name=body placeholder=Comment rows=10 cols=40></textarea><br>"
	content += "<input type=submit value=Comment>"
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
	markdown = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)

	r := mux.NewRouter()

	head = readFile("head.html")
	footer = readFile("footer.html")
	learning, _ := parseMarkdown(readFile("learning.md"))
	learning = assembleGenericPage("Learning", learning)
	posts := readPosts()

	postsList := "<ul class=posts>"
	for _, post := range posts {
		postsList += fmt.Sprintf(
			"<li>\n<time datetime=%v>%v</time>\n<a href=%v>%v</a>\n</li>", post.Date.Format("2006-01-02"), post.Date.Format("January 2, 2006"), post.Aliases[0], post.Title)
	}
	postsList += "</ul>"
	index = assemblePage("Caio Alonso", postsList)

	atom := atomFeed(posts)

	postsAliases = make(map[string]*Post)
	for _, post := range posts {
		for _, alias := range post.Aliases {
			postsAliases[alias] = &post
		}
	}

	r.HandleFunc("/posts/{post}", PostsHandler)
	r.HandleFunc("/posts/{post}/", PostsHandler)
	r.HandleFunc("/posts/{post}/comment", NewCommentHandler)
	// Legacy URLs I want to maintain
	r.HandleFunc("/{year}/{month}/{day}/{post}", PostsWithDateHandler)
	r.HandleFunc("/{year}/{month}/{day}/{post}", NewCommentWithDateHandler)

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

	r.HandleFunc("/learning", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, learning)
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
