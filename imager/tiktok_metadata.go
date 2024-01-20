package imager

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/log"
	"net"
	"net/http"
	"time"
)

type TiktokData struct {
	Version         string `json:"version"`
	Type            string `json:"type"`
	Title           string `json:"title"`
	AuthorURL       string `json:"author_url"`
	AuthorName      string `json:"author_name"`
	Width           string `json:"width"`
	Height          string `json:"height"`
	HTML            string `json:"html"`
	ThumbnailWidth  int    `json:"thumbnail_width"`
	ThumbnailHeight int    `json:"thumbnail_height"`
	ThumbnailURL    string `json:"thumbnail_url"`
	ProviderURL     string `json:"provider_url"`
	ProviderName    string `json:"provider_name"`
	AuthorUniqueID  string `json:"author_unique_id"`
	EmbedProductID  string `json:"embed_product_id"`
	EmbedType       string `json:"embed_type"`
}

var client = &http.Client{
	//CheckRedirect: func(req *http.Request, via []*http.Request) error {
	//	return http.ErrUseLastResponse // This prevents the client from following redirects
	//},
	Timeout: time.Second * 2,
}

// getTiktokUsername takes a filename as input and scans it for a tok ID
// Using the tok ID, it constructs a URL to access the TikTok video
// When tiktok redirects this url, it will insert an @[USERNAME] which we detect
const maxRetries = 2
const retryDelay = time.Second

func getTokID(filename string) *string {
	digits := ""
	for _, c := range filename {
		if c >= '0' && c <= '9' {
			digits += string(c)
		} else {
			if isValidTokID(digits) {
				return &digits
			}
			digits = ""
		}
	}
	if isValidTokID(digits) {
		return &digits
	}
	return nil
}

func getTiktokUsername(filename string) (string, error) {
	tokID := getTokID(filename)
	if tokID == nil {
		return "", errors.New("No TokID found")
	}
	// url := "https://www.tiktok.com/@/video/" + *tokID
	url := "https://www.tiktok.com/oembed?url=https://www.tiktok.com/@/video/" + *tokID
	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(url)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Errorf("Timeout error while accessing URL %s: %s", url, netErr.Error())
			} else {
				log.Error("Error accessing URL: ", err)
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(retryDelay) // Wait before retrying
			continue
		}
		if resp.StatusCode == 404 {
			resp.Body.Close()
			return "", errors.New("tiktok video not found")
		}
		var data TiktokData
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Error("Cannot decode json: ")
			resp.Body.Close()
			return "", err
		}
		resp.Body.Close()
		return data.AuthorName, nil
	}

	log.Error("No redirect found for URL: ", url)
	return "", errors.New("no redirect found")
}
