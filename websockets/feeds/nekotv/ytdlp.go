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
	Streamer    string `json:"display_id"`
	StreamTitle string `json:"description"`
	WebpageURL  string `json:"webpage_url"`
}
type KickData struct {
	Streamer    string `json:"channel""`
	StreamTitle string `json:"fulltitle"`
	WebpageURL  string `json:"webpage_url"`
}

func getTwitchData(link string) (*pb.VideoItem, error) {
	match := twitchStreamRegex.FindStringSubmatch(link)
	if match == nil {
		return nil, errors.New("invalid twitch link")
	}
	twitchURL := "https://www.twitch.tv/" + match[1]
	cmd := exec.Command("/usr/bin/yt-dlp", "--dump-json", twitchURL)
	var stdoutbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
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
	title := fmt.Sprintf("%s - %s", twitchData.Streamer, twitchData.StreamTitle)
	return &pb.VideoItem{
		Url:      twitchData.WebpageURL,
		Title:    title,
		Duration: common.Float32Infinite,
		Type:     pb.VideoType_TWITCH,
	}, nil
}

func getKickData(link string) (*pb.VideoItem, error) {
	match := kickStreamRegex.FindStringSubmatch(link)
	if match == nil {
		return nil, errors.New("invalid kick link")
	}
	kickUsername := match[1]
	kickURL := "https://kick.com/" + kickUsername
	cmd := exec.Command("/usr/bin/yt-dlp", "--dump-json", kickURL)
	var stdoutbuf bytes.Buffer
	cmd.Stdout = &stdoutbuf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	jsonBytes := stdoutbuf.Bytes()
	var kickData KickData
	err = json.Unmarshal(jsonBytes, &kickData)
	if err != nil {
		return nil, err
	}
	title := fmt.Sprintf("%s - %s", kickData.Streamer, kickData.StreamTitle)
	return &pb.VideoItem{
		Url:      kickData.WebpageURL,
		Title:    title,
		Duration: common.Float32Infinite,
		Id:       "https://player.kick.com/" + kickUsername,
		Type:     pb.VideoType_IFRAME,
	}, nil
}
