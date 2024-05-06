package nekotv

import (
	"regexp"
)

const (
	twitchStreamRegex = `(?:https?:\/\/)?(?:www\.)?twitch\.tv\/(\w+)(?:\/)?`
)

func isTwitchStream(link string) bool {
	match, _ := regexp.MatchString(twitchStreamRegex, link)
	return match
}
