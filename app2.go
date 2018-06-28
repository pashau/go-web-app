package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pashau/lottery-generator"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	println("## save()")
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	println("## loadPage()")
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

//var templates = template.Must(template.ParseGlob("tmpl/*.html"))
// https://stackoverflow.com/a/44222211 output Body as HTML in template
var templates = template.Must(template.New("main").Funcs(template.FuncMap{
	"safeHTML": func(b []byte) template.HTML {
		return template.HTML(b)
	},
}).ParseGlob("tmpl/*.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	println("## renderTemplate()")
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	println("## makeHandler()")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// set the page route to index
			pr := "index"
			fn(w, r, pr)
		} else {
			m := validPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
				http.NotFound(w, r)
				return
			}
			fn(w, r, m[2])
		}
	}
}

func getFileList(path string) []string {
	println("## getFileList()")
	files, err := filepath.Glob(path)
	if err != nil {
		log.Fatal(err)
	}
	return files
}

func stringToByte(stringArray []string) []byte {
	//s := "\x00" + strings.Join(stringArray, "\x20\x00") // x20 = space and x00 = null
	s := strings.Join(stringArray, "")
	//fmt.Println([]byte(s))
	//fmt.Println(string([]byte(s)))
	return []byte(s)
}

func generateIndexData() {
	println("## generateIndexData()")
	var dataList []string = getFileList("data/*") // [data/new.txt data/test.txt]
	for k, v := range dataList {
		//fmt.Printf("key:%v value:%v\n", k, v)
		dataList[k] = "<input type=\"button\" value=\"" + v + "\" onclick=\"location.href='/view/" + v[5:len(v)-4] + "';\" /><br>"
		//dataList[k] = "<a href=\"/view/" + v[5:len(v)-4] + "\">" + v + "</a><br>"
	}
	body := []byte(stringToByte(dataList))
	p := &Page{Title: "index", Body: body}
	p.save()
}

func indexHandler(w http.ResponseWriter, r *http.Request, title string) {
	println("## indexHandler()")
	generateIndexData()
	title = "index"
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, title, p)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	println("## viewHandler()")
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	println("## editHandler()")
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	println("## saveHandler()")
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	//generateIndexData()
	http.HandleFunc("/", makeHandler(indexHandler))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Print(lotto.uniquerandnumbers(5, 50))
}
