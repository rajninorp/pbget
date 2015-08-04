package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"regexp"
	"strconv"
	"time"
)

type SiteData struct {
	url string 
	title string
	lastpage int
	lastmodified time.Time
}

func siteData(url string) (SiteData, error) {
	data := SiteData{url:url, title:"", lastpage:0}
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return data, err
	}
	data.title = doc.Find("title").Text()
	r := regexp.MustCompile(`page_num=([\d]+)`)
	doc.Find("a").Each(func(_ int, s*goquery.Selection) {
		href, exist := s.Attr("href")
		if exist {
			m := r.FindStringSubmatch(href)
			if len(m) > 0 {
				pn, err := strconv.Atoi(m[1])
				if err != nil {
					log.Fatal(err)
				}
				if pn > data.lastpage {
					data.lastpage = pn
				}
			}
		}
	})
	doc.Find("div").Each(func(_ int, s *goquery.Selection) {
		class, exist := s.Attr("class")
		if exist && class == "post-sla" {
			s.Find("span").Each(func(_ int, s*goquery.Selection) {
				class, exist := s.Attr("class")
				if exist && class == "comment-created-ts" {
					t, err := time.Parse("2006/01/02 15:04", s.Text())
					if err != nil {
						log.Fatal(err)
					}
					data.lastmodified = t
					return
				}
			})
		}

	})
	return data, nil
}

func imageLink(url string) ([]string, error) {
	link := []string{}
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return link, err
	}
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		class, exist := s.Attr("class")
		if exist && class == "img-responsive" {
			src, exist := s.Attr("src")
			if exist {
				fmt.Println(src)
				link = append(link, src)
			}
		}
	})
	return link, nil
}

func main() {
	data, err := siteData("http://109815.peta2.jp/646757.html")
	if err != nil {
		log.Fatal(err)
	}
	for pn := 1; pn <= data.lastpage; pn++ {
		fmt.Println(data.url + "?comment_order=DESC&page_num=" + strconv.Itoa(pn))
		link, err := imageLink(data.url + "?comment_order=DESC&page_num=" + strconv.Itoa(pn))
		if err != nil {
			log.Fatal(err)
		}
		for _, l := range link {
			fmt.Println(l)
		}
	}
}
