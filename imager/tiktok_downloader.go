package imager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bakape/meguca/common"
	"github.com/go-playground/log"
	"golang.org/x/text/unicode/norm"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
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
)

type TikWMResponse struct {
	Code          int            `json:"code"`
	Msg           string         `json:"msg"`
	ProcessedTime float64        `json:"processed_time"`
	Data          *TWMTikTokData `json:"data"`
}

func rotateVideoFile(filename string, rotation int) error {
	cmd := exec.Command("exiftool", fmt.Sprintf("-rotation=%d", rotation), "-overwrite_original_in_place", filename)

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func downloadFileWithSizeCheck(url *string, file string, maxSize int64) (fileSize int64, status int, err error) {

	client := &http.Client{
		Timeout: 30 * time.Second, // Timeout for each request
	}

	headReq, err := http.NewRequest("HEAD", *url, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	headResp, err := client.Do(headReq)
	if err != nil {
		return 0, 0, fmt.Errorf("http HEAD error: %w", err)
	}
	headResp.Body.Close()

	status = headResp.StatusCode
	if status != http.StatusOK {
		return 0, status, fmt.Errorf("HEAD request failed with status: %d", status)
	}

	contentLengthStr := headResp.Header.Get("Content-Length")
	var expectedSize int64 = -1

	if contentLengthStr != "" {
		expectedSize, err = strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			fmt.Printf("Warning: Invalid Content-Length '%s' in HEAD response: %v. Proceeding without size check.\n", contentLengthStr, err)
			expectedSize = -1
			err = nil // Not fatal for download attempt itself
		}
	} else {
		fmt.Println("Warning: Content-Length missing in HEAD response. Proceeding without size check.")
	}

	if maxSize > 0 && expectedSize > 0 && expectedSize > maxSize {
		return 0, status, fmt.Errorf("file size (%d bytes from HEAD) exceeds maximum allowed size (%d bytes)", expectedSize, maxSize)
	}

	response, err := client.Get(*url)
	if err != nil {
		return 0, 0, fmt.Errorf("http GET error: %w", err)
	}
	defer response.Body.Close()

	status = response.StatusCode
	if status != http.StatusOK {
		return 0, status, fmt.Errorf("GET request failed with status: %d", status)
	}

	outputFile, err := os.Create(file)
	if err != nil {
		return 0, status, fmt.Errorf("failed to create temp file '%s': %w", file, err)
	}
	shouldCleanupFile := true
	defer func() {
		closeErr := outputFile.Close()
		if err != nil && shouldCleanupFile {
			removeErr := os.Remove(file)
			if removeErr != nil {
				log.Infof("Warning: Failed to remove partially downloaded file '%s': %v\n", file, removeErr)
			}
		} else if err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close output file: %w", closeErr)
		}
	}()

	fileSize, err = io.Copy(outputFile, response.Body)
	if err != nil {
		return 0, status, fmt.Errorf("error saving file during copy: %w", err)
	}

	if expectedSize != -1 && fileSize != expectedSize {
		shouldCleanupFile = true
		return 0, status, fmt.Errorf("download size mismatch: expected %d bytes (from HEAD), got %d bytes", expectedSize, fileSize)
	}

	shouldCleanupFile = false
	return fileSize, status, nil
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

func GetTikTokFilename(videoID string, caption string) string {
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
		return idPrefix + normalizedCaption + ".mp4"
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
	return idPrefix + strings.TrimSpace(truncatedCaption) + truncationSuffix + ".mp4"
}

func GetTikTokMetadata(input string) (tokData *TWMTikTokData, err error) {
	twmMutex.Lock()
	twmRequestChannel <- &common.PostCommand{Input: input}
	tokData = <-twmResponseChannel
	err = <-twmErrChannel
	twmMutex.Unlock()
	return
}

func runYtDlp(tokID, tempFilename string) (int64, error) {
	// Format the target URL using the token data
	url := fmt.Sprintf("https://www.tiktok.com/@/video/%s", tokID)

	// Run yt-dlp and capture combined stdout/stderr
	cmd := exec.Command("yt-dlp", url, "-o", tempFilename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("yt-dlp error: %w, output: %s", err, output)
	}

	// Stat the file to get its size
	info, err := os.Stat(tempFilename)
	if err != nil {
		return 0, fmt.Errorf("could not stat output file: %w", err)
	}

	// Return the size in bytes
	return info.Size(), nil
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
	//var size int64
	//urls := []*string{&tokData.HDPlay, &tokData.Play, &tokData.Wmplay}
	//for _, url := range urls {
	//	size, _, err = downloadFileWithSizeCheck(url, tmpFilename, 104857600)
	//	if err == nil {
	//		break
	//	} else {
	//		log.Error(err)
	//	}
	//}
	size, err := runYtDlp(tokData.ID, tmpFilename)
	if err != nil {
		defer os.Remove(tmpFilename)
		return
	}
	tmpFile, err := os.Open(tmpFilename)
	defer tmpFile.Close()
	log.Info("Rotation: ", input.Rotation)
	if input.Rotation > 0 {
		rotateVideoFile(tmpFilename, input.Rotation)
	}
	if err != nil {
		return
	}
	filename = GetTikTokFilename(tokData.ID, strings.Trim(tokData.Title, " "))
	res := <-requestThumbnailing(tmpFile, filename, int(size), &tokData.Author.UniqueID)
	if res.err != nil {
		err = res.err
		return
	}
	return res.imageID, filename, nil
}

func init() {
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
