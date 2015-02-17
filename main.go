package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/PuerkitoBio/goquery"
)

const url = "http://www.metal-archives.com"

type band struct {
	Name string
	URL  string
	Date string
}

var (
	bandsAdded   = make(map[string]*band)
	bandsUpdated = make(map[string]*band)

	twitter *anaconda.TwitterApi

	silent *bool
)

func getBands(doc *goquery.Document, cache map[string]*band, selector string) map[string]*band {
	bs := make(map[string]*band)
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		td := s.Find("td")
		info := td.First().Find("a").First()
		url, _ := info.Attr("href")
		b := &band{
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

func bandsJob(tweet bool, doc *goquery.Document, cache map[string]*band, selector, status string) {
	band := getBands(doc, cache, selector)
	for _, b := range band {
		if !*silent {
			log.Printf("band %s: %s (%s) %s\n", status, b.Name, b.URL, b.Date)
		}
		if tweet && twitter != nil {
			tw, err := twitter.PostTweet(fmt.Sprintf("Band %s: %s %s", status, b.Name, b.URL), nil)
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

func work(tweet bool) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	go bandsJob(tweet, doc, bandsAdded, "#additionBands tr", "added")
	go bandsJob(tweet, doc, bandsUpdated, "#updatedBands tr", "updated")
}

func main() {
	silent = flag.Bool("s", false, "silent mode")
	key := flag.String("key", "", "Twitter API consumer key")
	secret := flag.String("secret", "", "Twitter API consumer secret")
	accessToken := flag.String("access_token", "", "Twitter access key")
	accessSecret := flag.String("access_secret", "", "Twitter access secret")
	flag.Parse()

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

	println("Watching...")
	work(false)

	for {
		select {
		case <-time.Tick(30 * time.Second):
			work(true)
		}
	}
}
