package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	var results []string
	var i int
	var str string
	for {
		c := NewCollector()
		c.setStarred(10)

		if i == 0 {
			c.setStartUrl("https://github.com/gin-gonic/gin/network/dependents")
		} else {
			c = c.setStartUrl(str)
		}

		str = c.Start()
		results = append(results, c.results...)
		i++

		if str == "" {
			break
		}

	}
}

// github opensource used page collector
type Collector struct {
	engine *colly.Collector

	startUrl string

	nextPageLinks []string

	isFinish bool

	starred int

	results []string

	// for error
	sleepingSecond int
}

func NewCollector() *Collector {
	return &Collector{
		engine:         colly.NewCollector(),
		sleepingSecond: 5,
	}
}

func (c *Collector) Start() string {
	c.engine.Init()
	c.getNextLink()
	c.getProjects()
	c.retryEngine()
	c.engine.Visit(c.startUrl)

	for !c.isFinish {
		nextLink := c.nextPageLinks[len(c.nextPageLinks)-1]
		c.engine.Visit(nextLink)
	}

	return c.nextPageLinks[len(c.nextPageLinks)-1]
}

func (c *Collector) getNextLink() {
	c.engine.OnHTML(".paginate-container .BtnGroup a.BtnGroup-item:last-child", func(e *colly.HTMLElement) {
		nextLink := e.Attr("href")
		c.nextPageLinks = append(c.nextPageLinks, nextLink)
		if strings.Contains(e.Attr("disabled"), "disabled") {
			c.nextPageLinks = append(c.nextPageLinks, "")
			c.isFinish = true
		}
	})
}

func (c *Collector) getProjects() {
	c.engine.OnHTML(".application-main .Box", func(e *colly.HTMLElement) {
		e.ForEach(".Box-row", func(i int, h *colly.HTMLElement) {
			text := h.ChildText(".d-flex span:first-child")
			star, _ := strconv.Atoi(text)
			if star >= c.starred {
				url := h.ChildAttr(".f5 a.text-bold", "href")
				c.results = append(c.results, url)
				fmt.Println(c.results)
			}
		})
	})
}

func (c *Collector) retryEngine() {
	c.engine.OnError(func(r *colly.Response, err error) {
		if r.StatusCode != 200 {
			log.Println("Received 429 response. Retrying after 5 seconds...")
			time.Sleep(time.Duration(c.sleepingSecond) * time.Second)
			c.nextPageLinks = append(c.nextPageLinks, r.Request.URL.String())
			c.isFinish = true
		} else {
			log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		}
	})
}

func (c *Collector) setStartUrl(url string) *Collector {
	c.startUrl = url
	return c
}

func (c *Collector) setStarred(starred int) *Collector {
	c.starred = starred
	return c
}

func (c *Collector) checkStartPossible() error {
	if c.startUrl == "" {
		return errors.New("please set Url")
	}

	if c.starred == 0 {
		return errors.New("please set starred")
	}

	return nil
}
