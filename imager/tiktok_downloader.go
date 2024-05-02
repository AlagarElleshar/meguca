package imager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bakape/meguca/common"
	"github.com/go-playground/log"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/vansante/go-ffprobe.v2"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"unicode"
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
	twmMutex           sync.Mutex
	//transloaditClient  transloadit.Client
)

type TikWMResponse struct {
	Code          int            `json:"code"`
	Msg           string         `json:"msg"`
	ProcessedTime float64        `json:"processed_time"`
	Data          *TWMTikTokData `json:"data"`
}

func rotateVideoFile(filename string, rotation int) error {
	cmd := exec.Command("exiftool", fmt.Sprintf("-rotation=%d", rotation), "overwrite_original", filename)

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
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

func getFilename(videoID string, caption string) string {
	if len(caption) == 0 {
		return videoID
	}

	idPrefix := videoID + " "
	maxFilenameLength := 200
	remainingLength := maxFilenameLength - len(idPrefix)

	normalizedCaption := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, norm.NFC.String(caption))
	if len(normalizedCaption) <= remainingLength {
		return idPrefix + normalizedCaption
	}

	truncationSuffix := "… "
	remainingLength -= len(truncationSuffix)

	captionBytes := []byte(normalizedCaption)
	truncationPosition := 0
	for {
		nextBoundaryPosition := norm.NFC.NextBoundary(captionBytes[truncationPosition:], false)
		if nextBoundaryPosition == -1 {
			break
		}
		if nextBoundaryPosition+truncationPosition > remainingLength {
			break
		}
		truncationPosition += nextBoundaryPosition
	}

	truncatedCaption := string(captionBytes[:truncationPosition])
	return idPrefix + strings.TrimSpace(truncatedCaption) + truncationSuffix
}

func DownloadTikTok(input *common.PostCommand) (token string, filename string, err error) {
	twmMutex.Lock()
	twmRequestChannel <- input
	tokData := <-twmResponseChannel
	err = <-twmErrChannel
	twmMutex.Unlock()
	if err != nil || tokData == nil {
		return
	}
	if tokData.Images != nil {
		err = errors.New("Image URL Unsupported")
		return
	}
	tmpFilename := fmt.Sprintf("tmp/%s.mp4", tokData.ID)
	var size int64
	if tokData.Duration > 30 {
		input.HD = false
	}
	if input.HD {
		// Test to see if hdplay is actually h264
		var probeData *ffprobe.ProbeData
		probeData, err = ffprobe.ProbeURL(context.Background(), tokData.HDPlay)
		if probeData == nil {
			input.HD = false
		} else if probeData.FirstVideoStream().CodecName == "h264" {
			// Treat hdplay like an SD video
			input.HD = false
			tokData.Play = tokData.HDPlay
		} else {
			size, err = downloadConverted(tokData.HDPlay, &tokData.ID, tmpFilename, input.Rotation)
			if err != nil {
				defer os.Remove(tmpFilename)
				return
			}
		}
	}
	if !input.HD {
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
		return
	}
	return res.imageID, filename, nil
}

func init() {
	//options := transloadit.DefaultConfig
	//
	//var config struct {
	//	TransloaditAPIKey    string `json:"transloadit_api_key"`
	//	TransloaditAPISecret string `json:"transloadit_api_secret"`
	//}
	//
	//configBytes, err := os.ReadFile("config.json")
	//if err == nil {
	//	err = json.Unmarshal(configBytes, &config)
	//	if err == nil {
	//		options.AuthKey = config.TransloaditAPIKey
	//		options.AuthSecret = config.TransloaditAPISecret
	//	} else {
	//		fmt.Println(err.Error())
	//	}
	//}
	//transloaditClient = transloadit.NewClient(options)
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
