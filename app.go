package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/feeds"
	"github.com/ricallinson/forgery"
)

type Track struct {
	Id          int64  `xml:"id"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"permalink-url"`
	CreatedAt   string `xml:"created-at"`
	StreamUrl   string `xml:"stream-url"`
}

type TrackBody struct {
	Tracks []Track `xml:"track"`
}

type UserBody struct {
	Avatar      string `xml:"avatar-url"`
	Description string `xml:"description"`
	Username    string `xml:"username"`
}

var API_KEY string

func getTracks(username string) (TrackBody, error) {
	var result TrackBody

	apiUrl := "http://api.soundcloud.com/users/" + username + "/tracks.xml?client_id=" + API_KEY
	log.Println(apiUrl)

	resSoundcloud, err := http.Get(apiUrl)
	defer resSoundcloud.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	tracksXml, err := ioutil.ReadAll(resSoundcloud.Body)
	if err != nil {
		return result, err
	}

	xml.Unmarshal(tracksXml, &result)
	return result, err
}

func getUser(username string) (UserBody, error) {
	var result UserBody

	apiUrl := "http://api.soundcloud.com/users/" + username + ".xml?client_id=" + API_KEY
	log.Println(apiUrl)
	resSoundcloud, err := http.Get(apiUrl)
	defer resSoundcloud.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	userXml, err := ioutil.ReadAll(resSoundcloud.Body)
	if err != nil {
		return result, err
	}
	xml.Unmarshal(userXml, &result)
	return result, err
}

type rssFeedXml struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel *feeds.RssFeed
}

func generateFeed(username string) *feeds.RssFeed {
	user, err := getUser(username)
	if err != nil {
		log.Fatal(err)
	}

	tracks, err := getTracks(username)

	if err != nil {
		log.Fatal(err)
	}
	items := []*feeds.RssItem{}

	log.Println(user)
	rss := &feeds.RssFeed{
		Title:       user.Username,
		Description: user.Description,
		Link:        "soundcloud.com/" + username,
		Image: &feeds.RssImage{
			Url: user.Avatar,
		},
	}
	log.Println(tracks)
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
	API_KEY = os.Getenv("API_KEY")
	if err != nil {
		port = 3003
	}

	app := f.CreateServer()

	app.Get("/", func(req *f.Request, res *f.Response, next func()) {
		res.Set("Content-Type", "text/xml")
		res.Send("<duke></duke>")
	})

	app.Get("/:username", func(req *f.Request, res *f.Response, next func()) {
		feed := generateFeed(req.Params["username"])

		res.Set("Content-Type", "text/xml")

		rss := &rssFeedXml{Version: "2.0", Channel: feed}
		x, _ := xml.Marshal(rss)
		res.Send(x)
	})

	log.Println("Starting server on", port)
	app.Listen(port)
}
