package imager

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bakape/meguca/common"
	"github.com/go-playground/log"
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
	Message         string `json:"message,omitempty"`
	Code            int    `json:"code,omitempty"`
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

func getTiktokUsername(filename string) (string, error) {
	tokID := common.GetTokID(filename)
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
		if data.Code != 0 {
			resp.Body.Close()
			return "", errors.New("tiktok video not found")
		}
		resp.Body.Close()
		if len(data.AuthorUniqueID) == 0 {
			result, err := GetSourceEncode(*tokID)
			if err != nil {
				return "", errors.New("GetSourceEncode failed")
			}
			return result["unique_id"].(string), nil
		}

		return data.AuthorUniqueID, nil
	}

	log.Error("No redirect found for URL: ", url)
	return "", errors.New("no redirect found")
}

func GetSourceEncode(videoID string) (map[string]interface{}, error) {
	submitURL := "https://tikwm.com/api/video/task/submit"
	resultURL := "https://tikwm.com/api/video/task/result?task_id="

	client := &http.Client{Timeout: 15 * time.Second}

	// Submit task
	var taskID string

	for attempt := 0; attempt < 5; attempt++ {
		data := url.Values{}
		data.Set("web", "1")
		data.Set("url", videoID)

		req, err := http.NewRequest("POST", submitURL, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := client.Do(req)
		if err == nil {
			var result struct {
				Code int `json:"code"`
				Data struct {
					TaskID string `json:"task_id"`
				} `json:"data"`
			}

			if resp.StatusCode == http.StatusOK {
				err = json.NewDecoder(resp.Body).Decode(&result)
				resp.Body.Close()

				if err == nil && result.Code == 0 {
					taskID = result.Data.TaskID
					break
				}
			} else {
				resp.Body.Close()
			}
		}

		time.Sleep(2 * time.Second)
	}

	if taskID == "" {
		return nil, fmt.Errorf("failed to submit TikWM task: %s", videoID)
	}

	// Check task result
	for attempt := 0; attempt < 60; attempt++ {
		resp, err := client.Get(resultURL + url.QueryEscape(taskID))
		if err == nil {
			var result struct {
				Code int `json:"code"`
				Data struct {
					Status int `json:"status"`
					Detail struct {
						PlayURL string `json:"play_url"`
						Size    int64  `json:"size"`
						Author  struct {
							UniqueID string `json:"unique_id"`
						} `json:"author"`
					} `json:"detail"`
				} `json:"data"`
			}

			if resp.StatusCode == http.StatusOK {
				err = json.NewDecoder(resp.Body).Decode(&result)
				resp.Body.Close()

				if err == nil && result.Code == 0 {
					switch result.Data.Status {
					case 2:
						return map[string]interface{}{
							"play_url":  result.Data.Detail.PlayURL,
							"size":      result.Data.Detail.Size,
							"unique_id": result.Data.Detail.Author.UniqueID,
						}, nil
					case 3:
						return nil, nil
					}
				}
			} else {
				resp.Body.Close()
			}
		}

		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("failed to get source quality video: %s", videoID)
}
