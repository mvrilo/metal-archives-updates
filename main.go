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

	reviewsAdded = make(map[string]*data)

	twitter *anaconda.TwitterApi

	silent *bool
	tweet  *bool
)

func getData(doc *goquery.Document, cache map[string]*data, id string) map[string]*data {
	bands := make(map[string]*data)
	doc.Find(id + " tr").Each(func(i int, s *goquery.Selection) {
		td := s.Find("td")
		url, _ := td.Eq(0).Find("a").First().Attr("href")

		var info string
		if td.Length() == 3 {
			url, _ = td.Eq(1).Find("a").First().Attr("href")
			info += `"` + td.Eq(1).Text() + `", for `
		}
		info += td.Eq(0).Find("a").First().Text()

		b := &data{
			Name: info,
			Date: td.Last().Text(),
			URL:  url,
		}
		bands[b.URL] = b
	})

	for url1, b := range bands {
		for url2 := range cache {
			if url1 == url2 {
				delete(bands, url2)
			}
		}
		cache[url1] = b
	}

	return bands
}

func job(title string, t bool, doc *goquery.Document, cache map[string]*data, divID, status string) {
	data := getData(doc, cache, divID)
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

func rootWorker(t bool) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	go job("band", t, doc, bandsAdded, "#additionBands", "added")
	go job("band", t, doc, bandsUpdated, "#updatedBands", "updated")
	go job("review", t, doc, reviewsAdded, "#lastReviews", "added")
}

func labelsJob(t bool, url string, cache map[string]*data, status string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	go job("label", t, doc, cache, "#additionLabels", status)
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

	go job("artist", t, doc, cache, "#additionArtists", status)
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

	go rootWorker(false)
	go labelsWorker(false)
	go artistsWorker(false)

	for {
		select {
		case <-time.Tick(30 * time.Second):
			go rootWorker(*tweet)
			go labelsWorker(*tweet)
			go artistsWorker(*tweet)
		}
	}
}
