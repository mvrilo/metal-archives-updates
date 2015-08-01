package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mvrilo/malog"
)

var (
	twitter *anaconda.TwitterApi
	silent  *bool
	tweet   *bool
)

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

	first := true
	res, err := malog.Fetch()
	for {
		select {
		case e := <-err:
			if !*silent {
				log.Println(e)
			}
		case r := <-res:
			if first {
				continue
			}
			if !*silent {
				log.Printf("%s %s: %s (%s) %s\n", r.Title, r.Type, r.Name, r.URL, r.Date)
			}
			if *tweet && twitter != nil {
				tw, err := twitter.PostTweet(fmt.Sprintf("%s %s: %s %s", r.Title, r.Type, r.Name, r.URL), nil)
				if *silent {
					continue
				}
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("Tweet posted - id: %s\n", tw.IdStr)
			}
		case <-time.Tick(30 * time.Second):
			if first {
				first = false
			}
			res, err = malog.Fetch()
		}
	}
}
