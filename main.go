package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/PuerkitoBio/goquery"
)

const url = "http://www.metal-archives.com"

type data struct {
	Name string
	URL  string
	Date string
}

var (
	bandsAdded   = make(map[string]*data)
	bandsUpdated = make(map[string]*data)

	labelsAdded   = make(map[string]*data)
	labelsUpdated = make(map[string]*data)

	artistsAdded   = make(map[string]*data)
	artistsUpdated = make(map[string]*data)

	twitter *anaconda.TwitterApi

	silent *bool
	tweet  *bool
)

func getData(doc *goquery.Document, cache map[string]*data, selector string) map[string]*data {
	bs := make(map[string]*data)
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		td := s.Find("td")
		info := td.First().Find("a").First()
		url, _ := info.Attr("href")
		b := &data{
			Name: info.Text(),
			Date: td.Last().Text(),
			URL:  url,
		}
		bs[b.URL] = b
	})

	for url1, b := range bs {
		for url2, _ := range cache {
			if url1 == url2 {
				delete(bs, url2)
			}
		}
		cache[url1] = b
	}

	return bs
}

func job(title string, t bool, doc *goquery.Document, cache map[string]*data, selector, status string) {
	data := getData(doc, cache, selector)
	for _, b := range data {
		if !*silent {
			log.Printf("%s %s: %s (%s) %s\n", title, status, b.Name, b.URL, b.Date)
		}
		if t && twitter != nil {
			tw, err := twitter.PostTweet(fmt.Sprintf("%s %s: %s %s", strings.Title(title), status, b.Name, b.URL), nil)
			if *silent {
				continue
			}

			if err != nil {
				log.Println(err)
			} else {
				log.Printf("Tweet posted - id: %s\n", tw.IdStr)
			}
		}
	}
}

func bandsWorker(t bool) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	go job("band", t, doc, bandsAdded, "#additionBands tr", "added")
	go job("band", t, doc, bandsUpdated, "#updatedBands tr", "updated")
}

func labelsJob(t bool, url string, cache map[string]*data, status string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	job("label", t, doc, cache, "#additionLabels tr", status)
}

func labelsWorker(t bool) {
	go labelsJob(t, url+"/index/latest-labels", labelsAdded, "added")
	go labelsJob(t, url+"/index/latest-labels/by/modified", labelsUpdated, "updated")
}

func artistsJob(t bool, url string, cache map[string]*data, status string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	job("artist", t, doc, cache, "#additionArtists tr", status)
}

func artistsWorker(t bool) {
	go artistsJob(t, url+"/index/latest-artists", artistsAdded, "added")
	go artistsJob(t, url+"/index/latest-artists/by/modified", artistsUpdated, "updated")
}

func main() {
	tweet = flag.Bool("t", true, "Tweets!")
	silent = flag.Bool("s", false, "Silent mode")

	key := flag.String("key", "", "Twitter API consumer key")
	secret := flag.String("secret", "", "Twitter API consumer secret")
	accessToken := flag.String("access_token", "", "Twitter access key")
	accessSecret := flag.String("access_secret", "", "Twitter access secret")

	flag.Parse()

	if *tweet {
		if *key != "" && *secret != "" {
			anaconda.SetConsumerKey(*key)
			anaconda.SetConsumerSecret(*secret)
		} else {
			log.Fatal("You need to authorize Twitter API by passing the consumer key/secret")
		}

		if *accessToken != "" && *accessSecret != "" {
			twitter = anaconda.NewTwitterApi(*accessToken, *accessSecret)
		} else {
			log.Fatal("You need to authorize Twitter API by passing the access key/secret")
		}
	}

	if !*silent {
		log.Println("Watching...")
	}

	go bandsWorker(false)
	go labelsWorker(false)
	go artistsWorker(false)

	for {
		select {
		case <-time.Tick(30 * time.Second):
			go bandsWorker(*tweet)
			go labelsWorker(*tweet)
			go artistsWorker(*tweet)
		}
	}
}
