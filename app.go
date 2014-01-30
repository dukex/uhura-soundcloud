package main

import (
	"encoding/xml"
	"fmt"
	"github.com/gorilla/feeds"
	"github.com/ricallinson/forgery"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Track struct {
	Id          int64  `xml:"id"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"permalink-url"`
	CreatedAt   string `xml:"created-at"`
	StreamUrl   string `xml:"stream-url"`
}

type T struct {
	Tracks []Track `xml:"track"`
}

func getTracks(username string) ([]byte, error) {
	apiUrl := "http://api.soundcloud.com/users/" + username + "/tracks.xml?client_id=9747d5436f4eafe5dcb2c410da9ec009"

	resSoundcloud, err := http.Get(apiUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resSoundcloud.Body.Close()

	return ioutil.ReadAll(resSoundcloud.Body)
}

type rssFeedXml struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel *feeds.RssFeed
}

func generateFeed(tracks T, username string) *feeds.RssFeed {
	items := []*feeds.RssItem{}
	rss := &feeds.RssFeed{
		Title:       "Soundcloud " + username,
		Description: "Soundcloud Musics From " + username,
		Link:        "soundcloud.com/" + username,
	}
	for _, track := range tracks.Tracks {
		items = append(items, &feeds.RssItem{
			Title:       track.Title,
			Link:        track.Link,
			Description: track.Description,
			PubDate:     track.CreatedAt,
			Enclosure:   &feeds.RssEnclosure{Url: track.StreamUrl + "?client_id=9747d5436f4eafe5dcb2c410da9ec009"},
		})
	}
	rss.Items = items
	return rss
}

func main() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		port = 3003
	}

	app := f.CreateServer()

	app.Get("/", func(req *f.Request, res *f.Response, next func()) {
		res.Set("Content-Type", "text/xml")
		//
		res.Send("<duke>d</duke>")
	})

	app.Get("/:username", func(req *f.Request, res *f.Response, next func()) {
		tracksXml, _ := getTracks(req.Params["username"])

		var result T
		xml.Unmarshal(tracksXml, &result)

		feed := generateFeed(result, req.Params["username"])

		res.Set("Content-Type", "text/xml")

		rss := &rssFeedXml{Version: "2.0", Channel: feed}
		fmt.Println(*rss)
		x, _ := xml.Marshal(rss)
		res.Send(x)
	})

	fmt.Println("Starting server on", port)
	app.Listen(port)
}
