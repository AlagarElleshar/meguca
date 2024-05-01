package nekotv

import (
	"encoding/json"
	"fmt"
	"github.com/bakape/meguca/config"
	"github.com/bakape/meguca/pb"
	"github.com/go-playground/log"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	matchId          = regexp.MustCompile(`youtube\.com.*v=([A-z0-9_-]+)`)
	matchShort       = regexp.MustCompile(`youtu\.be\/([A-z0-9_-]+)`)
	matchShorts      = regexp.MustCompile(`youtube\.com\/shorts\/([A-z0-9_-]+)`)
	matchEmbed       = regexp.MustCompile(`youtube\.com\/embed\/([A-z0-9_-]+)`)
	matchPlaylist    = regexp.MustCompile(`youtube\.com.*list=([A-z0-9_-]+)`)
	videosUrl        = "https://www.googleapis.com/youtube/v3/videos"
	urlTitleDuration = "?part=snippet,contentDetails&fields=items(snippet/title,contentDetails/duration)"
	matchHours       = regexp.MustCompile(`(\d+)H`)
	matchMinutes     = regexp.MustCompile(`(\d+)M`)
	matchSeconds     = regexp.MustCompile(`(\d+)S`)
)

func convertTime(duration string) float32 {
	total := 0

	if hours := matchHours.FindStringSubmatch(duration); hours != nil {
		h, _ := strconv.Atoi(hours[1])
		total += h * 3600
	}

	if minutes := matchMinutes.FindStringSubmatch(duration); minutes != nil {
		m, _ := strconv.Atoi(minutes[1])
		total += m * 60
	}

	if seconds := matchSeconds.FindStringSubmatch(duration); seconds != nil {
		s, _ := strconv.Atoi(seconds[1])
		total += s
	}

	return float32(total)
}

func extractVideoID(url string) (string, error) {
	patterns := []*regexp.Regexp{
		matchId,
		matchShort,
		matchShorts,
		matchEmbed,
	}

	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(url); len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("no matching video ID found in URL: %s", url)
}

func GetVideoData(url string) (videoItem pb.VideoItem, err error) {
	var id string
	id, err = extractVideoID(url)
	if err != nil {
		return
	}

	dataURL := fmt.Sprintf("%s%s&id=%s&key=%s", videosUrl, urlTitleDuration, id, *config.Server.YoutubeApiKey)
	log.Info(dataURL)
	resp, err := http.Get(dataURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var jsonResp struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		Items []struct {
			Snippet struct {
				Title string `json:"title"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
		} `json:"items"`
	}
	if err = json.Unmarshal(body, &jsonResp); err != nil {
		return
	}

	if jsonResp.Error != nil {
		err = fmt.Errorf("youtube API error: %d %s", jsonResp.Error.Code, jsonResp.Error.Message)
		return
	}

	if len(jsonResp.Items) == 0 {
		return
	}

	for _, item := range jsonResp.Items {
		title := item.Snippet.Title
		duration := convertTime(item.ContentDetails.Duration)
		if duration == 0 {
			videoItem = pb.VideoItem{
				Duration: float32((99 * time.Hour) / time.Second),
				Title:    title,
				Url:      fmt.Sprintf(`<iframe src="https://www.youtube.com/embed/%s" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>`, id),
				IsIframe: true,
			}
			return
		}
		videoItem = pb.VideoItem{
			Duration: duration,
			Title:    title,
			Url:      url,
		}
		return
	}
	return
}
