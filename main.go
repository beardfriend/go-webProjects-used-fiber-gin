package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	var star int
	var url string

	flag.StringVar(&url, "url", "", "Enter package name e.g. gocolly/colly")
	flag.IntVar(&star, "star", 0, "Enter min star for crawl")
	flag.Parse()

	if url == "" || star == 0 {
		fmt.Println("error")
		os.Exit(1)
	}
	c := colly.NewCollector()

	var nextPageLinks []string
	var isFinish bool

	c.OnHTML(".paginate-container .BtnGroup a.BtnGroup-item:last-child", func(e *colly.HTMLElement) {
		nextLink := e.Attr("href")
		nextPageLinks = append(nextPageLinks, nextLink)
		if strings.Contains(e.Attr("disabled"), "disabled") {
			isFinish = true
		}
	})

	c.OnHTML(".application-main .Box", func(e *colly.HTMLElement) {
		e.ForEach(".Box-row", func(i int, h *colly.HTMLElement) {
			text := h.ChildText(".d-flex span:first-child")
			star, _ := strconv.Atoi(text)
			if star >= 1 {
				url := h.ChildAttr(".f5 a.text-bold", "href")
				fmt.Println(url)
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == 429 {
			log.Println("Received 429 response. Retrying after 5 seconds...")
			time.Sleep(5 * time.Second)
			c.Visit(r.Request.URL.String())
		} else {
			log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		}
	})

	err := c.Visit("https://github.com/gin-gonic/gin/network/dependents")
	if err != nil {
		log.Fatal(err)
	}

	for !isFinish {
		nextLink := nextPageLinks[len(nextPageLinks)-1]
		err := c.Visit(nextLink)
		if err != nil {
			log.Fatal(err)
		}
	}

	c.Wait()
}
