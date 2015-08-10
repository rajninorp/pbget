package main

import (
	"time"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"regexp"
	"strconv"
	"os"
	"bytes"
	"flag"
)

type Post struct {
	postDate time.Time
	subjectString string
	bodyString string
	mediaLink []string
}

const (
	TIME_FORMAT = "2006/01/02 15:04"
)

func errNotNilToPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func postDate(sel *goquery.Selection) (time.Time, error) {
	pd := time.Unix(0, 0)
	var err error = nil
	sel.Find("span").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if class, exist := s.Attr("class"); exist && class == "comment-created-ts" {
			pd, err = time.Parse(TIME_FORMAT, s.Text())
			errNotNilToPanic(err)
			return false
		}
		return true
	})
	jst := time.FixedZone("JST", 9*60*60)
	return pd.In(jst), err
}

func mediaLink(sel *goquery.Selection) ([]string, error) {
	ml := []string{}
	var err error = nil
	sel.Find("img").Each(func(_ int, s *goquery.Selection) {
		if class, exist := s.Attr("class"); exist && class == "img-responsive" {
			if src, exist := s.Attr("src"); exist {
				ml = append(ml, src)
			}
		}
	})
	return ml, err
}

func post(sel *goquery.Selection) (Post) {
	pd, err := postDate(sel)
	errNotNilToPanic(err)
	ml, err := mediaLink(sel)
	errNotNilToPanic(err)
	return Post{postDate:pd, mediaLink:ml}
}

func posts(url string, lastModified time.Time) ([]Post) {
	doc, err := goquery.NewDocument(url)
	errNotNilToPanic(err)
	lastPage := 0
	doc.Find("ul").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if class, exist := s.Attr("class"); exist && class == "pagination" {
			if href, exist := s.Find("li").Find("a").Last().Attr("href"); exist {
				reg := regexp.MustCompile(".*page_num=([0-9]+)$")
				if m := reg.FindStringSubmatch(href); len(m) > 1 {
					lastPage, _ = strconv.Atoi(m[1])
				}
			}
			return false
		}
		return true
	})
	pList := []Post{}
	for page := 1; page <= lastPage; page++ {
		doc, err := goquery.NewDocument(url + "?comment_order=DESC&page_num=" + strconv.Itoa(page))
		errNotNilToPanic(err)
		doc.Find("div").EachWithBreak(func(_ int, s *goquery.Selection) bool {
			if class, exist := s.Attr("class"); exist && class == "post-sla" {
				p := post(s)
				if !lastModified.Before(p.postDate) {
					return false
				}
				pList = append(pList, p)
			}
			return true
		})
	}
	return pList
}

var dlDir string
var dataFile string
var links []string

func init() {
	flag.StringVar(&dlDir, "dir", ".", "Download directory (default current directory)")
	flag.StringVar(&dataFile, "datafile", ".data", "Data file (default Download directory/.data)")
	flag.Parse()
	if dataFile == ".data" {
		dataFile = dlDir + "/" + dataFile
	}
	links = flag.Args()
	if len(links) == 0 {
		flag.Usage = func() {
			fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] URL1 URL2 ...
Options:
`, os.Args[0])
			flag.PrintDefaults()
		}
		flag.Usage()
	}
}

func main() {
	for _, link := range links {
		_, err := os.Stat(dataFile)
		fileExist := !os.IsNotExist(err)
		lastModified := time.Unix(0, 0)
		if fileExist {
			fp, err := os.Open(dataFile)
			errNotNilToPanic(err)
			defer fp.Close()
			buf := make([]byte, 10)
			fp.Read(buf)
			b := bytes.NewBuffer(buf)
			unixTime, err := strconv.ParseInt(b.String(), 10, 64)
			errNotNilToPanic(err)
			lastModified = time.Unix(unixTime, 0)
		}
		pList := posts(link, lastModified)
		for _, post := range pList {
			for _, media := range post.mediaLink {
				client := &Target{
					Url:media,
					Overwrite:true,
				}
				if err := client.Get(); err != nil {
					fmt.Println(err)
				}
			}
		}
		if len(pList) > 1 {
			fp, err := os.Create(dataFile)
			errNotNilToPanic(err)
			defer fp.Close()
			fp.WriteString(strconv.FormatInt(pList[0].postDate.Unix(), 10))
		}
	}
}
