package nekotv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bakape/meguca/common"
	"github.com/bakape/meguca/pb"
	"os/exec"
	"regexp"
)

var (
	twitchStreamRegex = regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?twitch\.tv\/(\w+)(?:\/)?`)
	kickStreamRegex   = regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?kick\.com\/(\w+)(?:\/)?`)
)

type TwitchData struct {
	ID          string `json:"id"`
	DisplayID   string `json:"display_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnails  []struct {
		URL        string `json:"url"`
		ID         string `json:"id"`
		Preference int    `json:"preference,omitempty"`
	} `json:"thumbnails"`
	Uploader   string `json:"uploader"`
	UploaderID string `json:"uploader_id"`
	Timestamp  int    `json:"timestamp"`
	Formats    []struct {
		FormatID    string  `json:"format_id"`
		URL         string  `json:"url"`
		ManifestURL string  `json:"manifest_url"`
		Tbr         float64 `json:"tbr"`
		Ext         string  `json:"ext"`
		Protocol    string  `json:"protocol"`
		HasDrm      bool    `json:"has_drm"`
		Vcodec      string  `json:"vcodec"`
		Acodec      string  `json:"acodec"`
		Resolution  string  `json:"resolution"`
		HTTPHeaders struct {
			UserAgent      string `json:"User-Agent"`
			Accept         string `json:"Accept"`
			AcceptLanguage string `json:"Accept-Language"`
			SecFetchMode   string `json:"Sec-Fetch-Mode"`
		} `json:"http_headers"`
		AudioExt string  `json:"audio_ext"`
		VideoExt string  `json:"video_ext"`
		Vbr      int     `json:"vbr"`
		Abr      float64 `json:"abr"`
		Format   string  `json:"format"`
		Width    int     `json:"width,omitempty"`
		Height   int     `json:"height,omitempty"`
	} `json:"formats"`
	IsLive             bool    `json:"is_live"`
	WebpageURL         string  `json:"webpage_url"`
	OriginalURL        string  `json:"original_url"`
	WebpageURLBasename string  `json:"webpage_url_basename"`
	WebpageURLDomain   string  `json:"webpage_url_domain"`
	Extractor          string  `json:"extractor"`
	ExtractorKey       string  `json:"extractor_key"`
	Thumbnail          string  `json:"thumbnail"`
	Fulltitle          string  `json:"fulltitle"`
	UploadDate         string  `json:"upload_date"`
	LiveStatus         string  `json:"live_status"`
	WasLive            bool    `json:"was_live"`
	Epoch              int     `json:"epoch"`
	FormatID           string  `json:"format_id"`
	URL                string  `json:"url"`
	ManifestURL        string  `json:"manifest_url"`
	Tbr                float64 `json:"tbr"`
	Ext                string  `json:"ext"`
	Fps                float64 `json:"fps"`
	Protocol           string  `json:"protocol"`
	HasDrm0            bool    `json:"has_drm"`
	Width              int     `json:"width"`
	Height             int     `json:"height"`
	Vcodec             string  `json:"vcodec"`
	Acodec             string  `json:"acodec"`
	DynamicRange       string  `json:"dynamic_range"`
	Resolution         string  `json:"resolution"`
	AspectRatio        float64 `json:"aspect_ratio"`
	HTTPHeaders        struct {
		UserAgent      string `json:"User-Agent"`
		Accept         string `json:"Accept"`
		AcceptLanguage string `json:"Accept-Language"`
		SecFetchMode   string `json:"Sec-Fetch-Mode"`
	} `json:"http_headers"`
	VideoExt  string `json:"video_ext"`
	AudioExt  string `json:"audio_ext"`
	Format    string `json:"format"`
	Filename  string `json:"_filename"`
	Filename0 string `json:"filename"`
	Type      string `json:"_type"`
	Version   struct {
		Version        string `json:"version"`
		ReleaseGitHead string `json:"release_git_head"`
		Repository     string `json:"repository"`
	} `json:"_version"`
}

func getTwitchData(link string) (*pb.VideoItem, error) {
	match := twitchStreamRegex.FindStringSubmatch(link)
	if match == nil {
		return nil, errors.New("Invalid twitch link")
	}
	twitchURL := "https://www.twitch.tv/" + match[1]
	cmd := exec.Command("yt-dlp", "--dump-json", twitchURL)
	var stdoutbuf bytes.Buffer
	var errorbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	cmd.Stderr = &errorbuf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	jsonBytes := stdoutbuf.Bytes()
	var twitchData TwitchData
	err = json.Unmarshal(jsonBytes, &twitchData)
	if err != nil {
		return nil, err
	}
	title := fmt.Sprintf("%s - %s", twitchData.Uploader, twitchData.Description)
	return &pb.VideoItem{
		Url:      twitchData.WebpageURL,
		Title:    title,
		Duration: common.Float32Infinite,
		Type:     pb.VideoType_TWITCH,
	}, nil
}

func getKickData(url string) (*pb.VideoItem, error) {
	match := kickStreamRegex.FindStringSubmatch(url)
	if match == nil {
		return nil, errors.New("Invalid kick link")
	}
	kickUsername := match[1]
	return &pb.VideoItem{
		Url:      "https://player.kick.com/" + kickUsername,
		Title:    kickUsername,
		Duration: common.Float32Infinite,
		Type:     pb.VideoType_IFRAME,
	}, nil
}
