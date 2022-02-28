package main

import (
	"bytes"
	"fmt"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"
)

func readPosts() (posts []Post) {

	return
}

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
	Date        string
	LegibleDate string
	Markdown    string
	HTML        string
	Aliases     []string
}

func main() {
	head := readFile("head.html")
	footer := readFile("footer.html")
	index := readFile("index.html")
	learning := readFile("learning.html")

	learning = head + learning + footer

	var posts []Post
	err := filepath.Walk("posts", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			posts = append(posts, Post{Filename: path})
		}
		return nil
	})
	if err != nil {
		panic(err)
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
			panic(err)
		}
		posts[i].HTML = head + buf.String() + footer
		metadata := meta.Get(context)
		posts[i].Title = fmt.Sprintf("%v", metadata["title"])
		posts[i].Date = fmt.Sprintf("%v", metadata["date"])
		layout := "2006-01-02"
		date, err := time.Parse(layout, posts[i].Date)
		if err != nil {
			panic(err)
		}
		posts[i].LegibleDate = date.Format("January 2, 2006")
		switch reflect.TypeOf(metadata["aliases"]).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(metadata["aliases"])
			for j := 0; j < s.Len(); j++ {
				posts[i].Aliases = append(posts[i].Aliases, fmt.Sprintf("%v", s.Index(j)))
			}
		}
	}

	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})

	index = head
	index += "<ul class=posts>"
	for _, post := range posts {
		index += fmt.Sprintf(
			"<li>\n<time datetime=%v>%v</time>\n<a href=%v>%v</a>\n</li>", post.Date, post.LegibleDate, post.Aliases[0], post.Title)
	}
	index += "</ul>" + footer

	for _, post := range posts {
		for _, alias := range post.Aliases {
			http.HandleFunc(alias, func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, post.HTML)
			})
		}
	}

	http.HandleFunc("/learning", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, learning)
	})

	http.HandleFunc("/learning/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, learning)
	})

	fsHandle := http.FileServer(http.Dir("/home/caio/public_html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			fmt.Fprintf(w, index)
		} else {
			fsHandle.ServeHTTP(w, r)
		}
	})

	http.ListenAndServe(":5050", nil)
}
