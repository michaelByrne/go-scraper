package main

import (
	"fmt"
	"strings"
	"github.com/gocolly/colly"
	"time"
	"strconv"
	"sync"
	"math/rand"
	"os"
)

func main() {

	numShows := 55
	showNumber := 0
	rand.Seed( time.Now().UnixNano() )

	d, err := os.Getwd()
	if err != nil {
		fmt.Println( err )
	}
	filename := d + "/hann.txt"

	f, err := os.OpenFile( filename, os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0666 )

	if err != nil {
		fmt.Println( err )
	}

	var w *sync.WaitGroup = new(sync.WaitGroup)

	pages := make([]string, numShows)
	pages = generateRiverUrls(numShows)

	v := getTranscriptUrls(pages)

	for k, _ := range v{
		in := make(chan string)
		go scrapePage(k, f, w, showNumber, in)
		showNumber++
		result := <-in
		WriteToFile(result, f, w, showNumber)
	}

	w.Wait()

}

func scrapePage(url string, f *os.File, w *sync.WaitGroup, showNum int, ret chan string){
	w.Add(1)
	defer w.Done()

	isHannity := false

	g := colly.NewCollector()
	text := ""

	g.OnHTML("p", func(e *colly.HTMLElement) {
		words := strings.Fields(e.Text)
		if len(words) > 0 {
			if strings.ToUpper(words[0]) == words[0] {
				if strings.Compare(words[0], "HANNITY:") == 0 {
					words = append(words[:0], words[1:]...)
					isHannity = true
				} else {
					isHannity = false
				}
			}
			if isHannity{
				text += fmt.Sprintln(strings.Join(words, " "))
			}
		}
	})

	g.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	g.Visit(url)
	ret <- text
}


func WriteToFile( i string, f *os.File, w *sync.WaitGroup, showNum int ){

	fmt.Fprintf( f, "show#: %d  ***************************************************\n %s\n", showNum, i )


}

func generateRiverUrls(num int) []string{
	t := time.Now().UnixNano() / int64(time.Millisecond)
	times := make([]string, num)
	for i := 0; i < num; i++ {
		t = t - int64(1209600000)
		ts := strconv.Itoa(int(t))
		url := fmt.Sprintf("http://www.foxnews.com/category/shows/hannity/transcript/_jcr_content/category-river.results.more.%s.html", ts)
		times = append(times, url)
	}
	return times
}

func getTranscriptUrls(pages []string) map[string]bool{
	c := colly.NewCollector()
	visited := map[string]bool{}
	count := 0

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		path := e.Attr("href")
		link := "http://www.foxnews.com"
		link += path
		if strings.Contains(link, "transcript") && !visited[link]{
			//fmt.Printf("Link found:%s\n", link)
			visited[link] = true
		}
	})
	c.OnResponse(func(r *colly.Response){
		count++
		fmt.Println(len(r.Body))
		fmt.Println(count)
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	for _, e := range pages{
		c.Visit(e)
	}
	return visited
}