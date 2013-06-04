package main

import (
	"encoding/xml"
	"fmt"
	"github.com/mrjones/oauth"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var Consumer *oauth.Consumer
var Token oauth.AccessToken

const (
	ConsumerKey       = "REDACTED"
	ConsumerSecret    = "REDACTED"
	AccessToken       = "REDACTED"
	AccessTokenSecret = "REDACTED"
	LastFMAPIKey      = "REDACTED"
)

type Response struct {
	XMLName xml.Name     `xml:"lfm"`
	Root    RecentTracks `xml:"recenttracks"`
	Status  string       `xml:"status,attr"`
}

type RecentTracks struct {
	Tracks []Track `xml:"track"`
}

type Track struct {
	NowPlaying string `xml:"nowplaying,attr"`
	Artist     string `xml:"artist"`
	Name       string `xml:"name"`
	Album      string `xml:"album"`
}

func (t Track) IsNowPlaying() bool {
	return t.NowPlaying == "true"
}

func ConstructURL() string {
	url, err := url.Parse("http://ws.audioscrobbler.com/2.0/")
	if err != nil {
		log.Fatal(err)
	}
	args := url.Query()
	args.Add("method", "user.getrecenttracks")
	args.Add("user", "mhh91")
	args.Add("api_key", LastFMAPIKey)
	args.Add("limit", "1")
	url.RawQuery = args.Encode()
	return url.String()
}

func MakeRequest(url string) (read []byte) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	read, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	return
}

func ParseXML(response []byte) (resp Response) {
	err := xml.Unmarshal(response, &resp)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func GetLatestTrack() (track Track) {
	data := ParseXML(MakeRequest(ConstructURL()))
	track = data.Root.Tracks[0]
	return
}

func InitTwitter(key, secret string) {
	p := oauth.ServiceProvider{
		RequestTokenUrl:   "http://api.twitter.com/oauth/request_token",
		AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
		AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
	}
	Consumer = oauth.NewConsumer(key, secret, p)
}

func SetAccessToken(token, secret string) {
	Token = oauth.AccessToken{token, secret}
}

func PostTweet(tweet string) (err error) {
	InitTwitter(ConsumerKey, ConsumerSecret)
	SetAccessToken(AccessToken, AccessTokenSecret)
	_, err = Consumer.Post(
		"https://api.twitter.com/1.1/statuses/update.json",
		map[string]string{"status": tweet},
		&Token)
	return
}

func main() {
	track := GetLatestTrack()
	if track.IsNowPlaying() {
		tweet := fmt.Sprintf("#Nowplaying %v - %v.\n", track.Artist, track.Name)
		PostTweet(tweet)
		fmt.Println("Done.")
	} else {
		fmt.Println("No track playing, aborting.")
	}

}
