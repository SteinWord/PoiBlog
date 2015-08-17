package main

import (
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"text/template"
)

const (
	_ROOTDIR = "/path/to/blog/root"
)

type Article struct {
	Date  string
	Title string
	Body  string
}

type Top struct {
	ArticleList string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	http.Handle("/box-design/", http.StripPrefix("/box-design/", http.FileServer(http.Dir(_ROOTDIR+"/box-design"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(_ROOTDIR+"/img"))))
	http.HandleFunc("/article", ArticleHandler)
	http.HandleFunc("/", MainHandler)
	http.ListenAndServe(":6789", nil)
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	articles, err := ioutil.ReadDir(_ROOTDIR + "/articles/")
	w.Header().Set("Content-Type", "text/html")
	re := regexp.MustCompile("404|.git|([\\s\\S]*)\\.md")
	var result ByArticles
	var tmp, output string
	topTemplate, err := ioutil.ReadFile(_ROOTDIR + "/tmpl/top.html")
	check(err)
	if err != nil {
		output = "<li>No Article</li>"
	} else if len(articles) == 0 {
		output = "<li>No Article</li>"
	} else {
		sort.Sort(ByArticles(articles))
		for i, _ := range articles {
			tmp = re.ReplaceAllString(articles[i].Name(), "$1")
			if tmp != "" {
				result = append(result, articles[i])
			}
		}
	}
	for j, _ := range result {
		output += string(fmt.Sprintf("<li><ul><li class=\"date\">%s</li><li><a href=\"/article?p=%s\">%s</a></li></ul></li>", result[j].ModTime().Format("2006/01/02 15:04"), re.ReplaceAllString(result[j].Name(), "$1"), re.ReplaceAllString(result[j].Name(), "$1")))
	}
	page := Top{ArticleList: string(output)}
	tmpl, _ := template.New("top").Parse(string(topTemplate))
	tmpl.Execute(w, page)
}

type ByArticles []os.FileInfo

func (articles ByArticles) Len() int {
	return len(articles)
}

func (articles ByArticles) Swap(i, j int) {
	articles[i], articles[j] = articles[j], articles[i]
}

func (articles ByArticles) Less(i, j int) bool {
	return articles[i].ModTime().After(articles[j].ModTime())
}

func ArticleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	r.ParseForm()
	articleName, ok := r.Form["p"]
	if !ok {
		articleName = []string{"404"}
	}
	_, err := os.Stat(_ROOTDIR + "/articles/" + string(articleName[0]) + ".md")
	if err != nil {
		articleName = []string{"404"}
	}
	md, err := ioutil.ReadFile(_ROOTDIR + "/articles/" + string(articleName[0]) + ".md")
	var date string
	f := func() {
		tmp, _ := os.Lstat(_ROOTDIR + "/articles/" + articleName[0] + ".md")
		date = tmp.ModTime().Format("2006/01/02 15:04")
	}
	/*tmp, _ := ioutil.ReadDir(_ROOTDIR + "/articles/")
	for i, _ := range tmp {
		if tmp[i].Name() == articleName[0]+".md" {
			date = tmp[i].ModTime().Format("2006/01/02 15:04")
		}
	}*/
	f()
	check(err)
	articleTemplate, err := ioutil.ReadFile(_ROOTDIR + "/tmpl/article.html")
	check(err)
	output := blackfriday.MarkdownCommon([]byte(md))
	page := Article{Body: string(output), Title: articleName[0], Date: string(date)}
	tmpl, _ := template.New("test").Parse(string(articleTemplate))
	if articleName[0] == "404" {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	tmpl.Execute(w, page)
}
