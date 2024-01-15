package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
)

const tmplt = "newsaggtemplate.gohtml"

var wg sync.WaitGroup
var HomeDir, execname, workingdir, ans, ExecTimeStamp, fulltmplt string

type NewsMap struct {
	Keyword  string
	Location string
}

type NewsAggPage struct {
	Title string
	News  map[string]NewsMap
}

type Sitemapindex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1> <a href=\"/agg/\">Washington Post Aggregator</a></h1>")
}

func newsRoutine(c chan News, Location string) {
	defer wg.Done()
	var n News
	resp, _ := http.Get(Location)
	// bytes, _ := ioutil.ReadAll(resp.Body)  Deprecated as of Go 1.16.
	bytes, _ := io.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)
	resp.Body.Close()
	c <- n
}

func newsAggHandler(w http.ResponseWriter, r *http.Request) {

	var s Sitemapindex
	resp, _ := http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	// bytes, _ := ioutil.ReadAll(resp.Body) Deprecated as of Go 1.16
	bytes, _ := io.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)
	news_map := make(map[string]NewsMap)
	resp.Body.Close()
	queue := make(chan News, 300)

	for _, Location := range s.Locations {
		wg.Add(1)
		go newsRoutine(queue, Location)
	}
	wg.Wait()
	close(queue)

	for elem := range queue {
		for idx, _ := range elem.Keywords {
			news_map[elem.Titles[idx]] = NewsMap{elem.Keywords[idx], elem.Locations[idx]}
		}
	}

	p := NewsAggPage{Title: "A News Aggregator based on the Washington Post", News: news_map}
	t, err := template.ParseFiles(tmplt)
	//	                                                         t, err := template.ParseFiles("newsaggtemplate.gohtml")
	check(err)
	//	                                                         t, _ := template.ParseFiles("aggregatorfinish.html")
	t.Execute(w, p)
}

func main() {
	if runtime.GOOS == "linux" {
		HomeDir = os.Getenv("HOME")
	} else if runtime.GOOS == "windows" {
		HomeDir = os.Getenv("userprofile")
	} else { // then HomeDir will be empty.
		fmt.Println(" runtime.GOOS does not say linux or windows.  Is this a Mac?")
	}
	workingdir, _ := os.Getwd()

	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	ExecTimeStamp = ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	pathsep := string(os.PathSeparator)
	tmplt1 := workingdir + pathsep + tmplt
	tmplt2 := HomeDir + pathsep + tmplt

	fmt.Println(ExecFI.Name(), "timestamp is", ExecTimeStamp, ".  Full exec is", execname)
	fulltmplt = tmplt1
	_, err := os.Stat(tmplt1)
	if err != nil {
		fulltmplt = tmplt2
		_, err = os.Stat(tmplt2)
		if err != nil {
			fmt.Println(" Template file not found in ", workingdir, " or ", HomeDir, ".  Exiting.")
			os.Exit(1)
		}
	}
	fmt.Println(" Using", fulltmplt, " as template file.")
	fmt.Println()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/agg/", newsAggHandler)
	//	fmt.Print("Hit <enter> to continue   ")
	//	fmt.Scanln(&ans)
	//	if strings.ToUpper(ans) == "QUIT" {
	//		os.Exit(0)
	//	}

	http.ListenAndServe(":8000", nil)
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
