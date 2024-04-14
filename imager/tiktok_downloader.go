package imager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bakape/meguca/common"
	"github.com/go-playground/log"
	transloadit "github.com/transloadit/go-sdk"
	"golang.org/x/text/unicode/norm"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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
	HDPlay      string `json:"hdplay"`
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
	PlayCount     int             `json:"play_count"`
	DiggCount     int             `json:"digg_count"`
	CommentCount  int             `json:"comment_count"`
	ShareCount    int             `json:"share_count"`
	DownloadCount int             `json:"download_count"`
	CollectCount  int             `json:"collect_count"`
	CreateTime    int             `json:"create_time"`
	IsAd          bool            `json:"is_ad"`
	Images        json.RawMessage `json:"images"`
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
	twmRequestChannel  chan *common.PostCommand
	twmResponseChannel chan *TWMTikTokData
	twmErrChannel      chan error
	transloaditClient  transloadit.Client
)

type TikWMResponse struct {
	Code          int            `json:"code"`
	Msg           string         `json:"msg"`
	ProcessedTime float64        `json:"processed_time"`
	Data          *TWMTikTokData `json:"data"`
}

func rotateVideoFile(filename string, rotation int) error {
	cmd := exec.Command("exiftool", fmt.Sprintf("-rotation=%d", rotation), filename)

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func downloadHDToTemp(url string, file string, rotation int) (fileSize int64, status int, err error) {
	assembly := transloadit.NewAssembly()
	//assembly.AddFile("video", url)
	assembly.AddStep("imported", map[string]interface{}{
		"robot": "/http/import",
		"url":   url,
	})
	encodeStep := map[string]interface{}{
		"use":          "imported",
		"robot":        "/video/encode",
		"ffmpeg_stack": "v6.0.0",
		"preset":       "hls-1080p",
		"ffmpeg": map[string]interface{}{
			"vcodec": "libx264",
			"crf":    20,
			"c:a":    "copy",
			"preset": "faster",
		},
		"width":  "${file.meta.width}",
		"height": "${file.meta.height}",
	}
	if rotation > 0 {
		encodeStep["rotate"] = rotation
	}
	assembly.AddStep("encoded-video", encodeStep)
	assembly.TemplateID = "4e4e97b7977b442a836041bf0dd3ba71"

	// Start the upload
	info, err := transloaditClient.StartAssembly(context.Background(), assembly)
	if err != nil {
		panic(err)
	}
	info, err = transloaditClient.WaitForAssembly(context.Background(), info)
	if err != nil {
		panic(err)
	}
	out_url := info.Results["encoded-video"][0].URL
	outputFile, err := os.Create(file)
	if err != nil {
		return
	}
	defer outputFile.Close()

	response, err := http.Get(out_url)
	if err != nil {
		return 0, status, err
	} else if response.StatusCode != http.StatusOK {
		return 0, response.StatusCode, fmt.Errorf("status code %d", response.StatusCode)
	}
	defer response.Body.Close()

	fileSize, err = io.Copy(outputFile, response.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}
	return
}
func downloadToTemp(url string, file string) (fileSize int64, status int, err error) {
	outputFile, err := os.Create(file)
	if err != nil {
		return
	}
	defer outputFile.Close()

	response, err := http.Get(url)
	if err != nil {
		return 0, status, err
	} else if response.StatusCode != http.StatusOK {
		return 0, response.StatusCode, fmt.Errorf("status code %d", response.StatusCode)
	}
	defer response.Body.Close()

	fileSize, err = io.Copy(outputFile, response.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		return
	}
	return
}

func getFilename(id string, desc string) string {
	if len(desc) == 0 {
		return id
	}
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

func DownloadTikTok(input *common.PostCommand) (token string, filename string, err error) {
	twmRequestChannel <- input
	tokData := <-twmResponseChannel
	err = <-twmErrChannel
	if err != nil || tokData == nil {
		return
	}
	if tokData.Images != nil {
		err = errors.New("Image URL Unsupported")
		return
	}
	tmpFilename := fmt.Sprintf("tmp/%s.mp4", tokData.ID)
	var size int64
	if tokData.Duration > 20 {
		input.HD = false
	}
	if input.HD {
		size, _, err = downloadHDToTemp(tokData.HDPlay, tmpFilename, input.Rotation)
		if err != nil {
			defer os.Remove(tmpFilename)
			return
		}
	} else {
		var status int
		size, status, err = downloadToTemp(tokData.Play, tmpFilename)
		if err != nil {
			if status == http.StatusNotFound {
				size, status, err = downloadToTemp(tokData.Wmplay, tmpFilename)
				if err != nil {
					defer os.Remove(tmpFilename)
					return
				}
			} else {
				defer os.Remove(tmpFilename)
				return
			}
		}
		if input.Rotation != 0 {
			err = rotateVideoFile(tmpFilename, input.Rotation)
		}
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
	options := transloadit.DefaultConfig

	var config struct {
		TransloaditAPIKey    string `json:"transloadit_api_key"`
		TransloaditAPISecret string `json:"transloadit_api_secret"`
	}

	configBytes, err := os.ReadFile("config.json")
	if err == nil {
		err = json.Unmarshal(configBytes, &config)
		if err == nil {
			options.AuthKey = config.TransloaditAPIKey
			options.AuthSecret = config.TransloaditAPISecret
		} else {
			fmt.Println(err.Error())
		}
	}
	transloaditClient = transloadit.NewClient(options)
	twmRequestChannel = make(chan *common.PostCommand)
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

func getTiktokVideoURL(twmInput *common.PostCommand) (tiktokData *TWMTikTokData, err error) {
	client := &http.Client{}

	data := url.Values{
		"url": {twmInput.Input},
		"web": {"0"},
		"hd":  {"1"},
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
