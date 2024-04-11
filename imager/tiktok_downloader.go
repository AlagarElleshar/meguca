package imager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/log"
	"golang.org/x/text/unicode/norm"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type TWMTikTokData struct {
	ID          string `json:"id"`
	Region      string `json:"region"`
	Title       string `json:"title"`
	Cover       string `json:"cover"`
	OriginCover string `json:"origin_cover"`
	Duration    int    `json:"duration"`
	Play        string `json:"play"`
	Wmplay      string `json:"wmplay"`
	Size        int    `json:"size"`
	WmSize      int    `json:"wm_size"`
	Music       string `json:"music"`
	MusicInfo   struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Play     string `json:"play"`
		Cover    string `json:"cover"`
		Author   string `json:"author"`
		Original bool   `json:"original"`
		Duration int    `json:"duration"`
		Album    string `json:"album"`
	} `json:"music_info"`
	PlayCount     int  `json:"play_count"`
	DiggCount     int  `json:"digg_count"`
	CommentCount  int  `json:"comment_count"`
	ShareCount    int  `json:"share_count"`
	DownloadCount int  `json:"download_count"`
	CollectCount  int  `json:"collect_count"`
	CreateTime    int  `json:"create_time"`
	IsAd          bool `json:"is_ad"`
	CommerceInfo  struct {
		AdvPromotable          bool `json:"adv_promotable"`
		AuctionAdInvited       bool `json:"auction_ad_invited"`
		BrandedContentType     int  `json:"branded_content_type"`
		WithCommentFilterWords bool `json:"with_comment_filter_words"`
	} `json:"commerce_info"`
	CommercialVideoInfo string `json:"commercial_video_info"`
	ItemCommentSettings int    `json:"item_comment_settings"`
	Author              struct {
		ID       string `json:"id"`
		UniqueID string `json:"unique_id"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	} `json:"author"`
}

var (
	twmRequestChannel  chan *string
	twmResponseChannel chan *TWMTikTokData
	twmErrChannel      chan error
)

type TikWMResponse struct {
	Code          int            `json:"code"`
	Msg           string         `json:"msg"`
	ProcessedTime float64        `json:"processed_time"`
	Data          *TWMTikTokData `json:"data"`
}

func downloadToTemp(url string, file string) (fileSize int64, err error) {
	outputFile, err := os.Create(file)
	if err != nil {
		return 0, err
	}
	defer outputFile.Close()

	response, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	fileSize, err = io.Copy(outputFile, response.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return 0, err
	}
	return fileSize, nil
}

func getFilename(id string, desc string) string {
	prepend := id + " "
	remainingLen := 200 - len(prepend)
	normalize := norm.NFC.String(desc)
	if len(normalize) <= remainingLen {
		return prepend + normalize
	}
	descBytes := []byte(normalize)
	pos := 0
	for {
		nextPos := norm.NFC.NextBoundary(descBytes[pos:], false)
		if nextPos == -1 {
			break
		}
		if nextPos > remainingLen {
			break
		}
		pos += nextPos
	}
	return prepend + string(descBytes[:pos])
}

func DownloadTikTok(input string) (token string, filename string, err error) {
	twmRequestChannel <- &input
	tokData := <-twmResponseChannel
	err = <-twmErrChannel
	if err != nil || tokData == nil {
		return
	}
	tmpFilename := fmt.Sprintf("tmp/%s.mp4", tokData.ID)
	size, err := downloadToTemp(tokData.Play, tmpFilename)
	if err != nil {
		return
	}
	tmpFile, err := os.Open(tmpFilename)
	defer tmpFile.Close()
	defer os.Remove(tmpFilename)
	if err != nil {
		return
	}
	filename = getFilename(tokData.ID, strings.Trim(tokData.Title, " "))
	res := <-requestThumbnailing(tmpFile, filename, int(size), &tokData.Author.UniqueID)
	if res.err != nil {
		err = res.err
	}
	return res.imageID, filename, nil
}

func initTiktokDownloader() {
	twmRequestChannel = make(chan *string)
	twmResponseChannel = make(chan *TWMTikTokData)
	twmErrChannel = make(chan error)
	go func() {
		for {
			input := <-twmRequestChannel
			resp, err := getTiktokVideoURL(input)
			twmResponseChannel <- resp
			twmErrChannel <- err
			time.Sleep(1500)
		}
	}()

}

func getTiktokVideoURL(twmInput *string) (tiktokData *TWMTikTokData, err error) {
	client := &http.Client{}

	data := url.Values{
		"url": {*twmInput},
		"web": {"0"},
		"hd":  {"0"},
	}

	req, err := http.NewRequest("POST", "https://tikwm.com/api/", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	var response TikWMResponse
	err = json.Unmarshal(bodyText, &response)
	if err != nil {
		return
	}
	if response.Code == 0 {
		tiktokData = response.Data
	} else {
		err = errors.New(response.Msg)
	}
	return
}
