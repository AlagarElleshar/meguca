// Package parser parses and verifies user-sent post data
package parser

import (
	"bytes"
	"github.com/rivo/uniseg"
	"regexp"
	"unicode"

	"github.com/bakape/meguca/common"
	"github.com/bakape/meguca/config"
	"github.com/bakape/meguca/util"
)

var (
	linkRegexp = regexp.MustCompile(`^>{2,}(\d+)$`)
)

// Needed to avoid cyclic imports for the 'db' package
func init() {
	common.ParseBody = ParseBody
}

// ParseBody parses the entire post text body for commands and links.
// internal: function was called by automated upkeep task
func ParseBody(body []byte, board string, thread uint64, id uint64, ip string, internal bool) (links []common.Link, com []common.Command, claude *common.ClaudeState, postCommand *common.PostCommand, mediaCommands []common.MediaCommand, err error) {
	err = IsPrintableString(string(body), true)
	if err != nil {
		if internal {
			err = nil
			// Strip any non-printables for automated post closing
			s := make([]byte, 0, len(body))
			for _, r := range []rune(string(body)) {
				if IsPrintable(r, true) == nil {
					s = append(s, string(r)...)
				}
			}
			body = s
		} else {
			return
		}
	}

	start := 0
	lineStart := 0
	pyu := config.GetBoardConfigs(board).Pyu

	// Prevent link duplication
	haveLink := make(map[uint64]bool)
	// Prevent #pyu duplication
	isSlut := false
	// Prevent #autobahn duplication
	isDead := false

	for i, b := range body {
		switch b {
		case '\n', ' ', '\t':
		default:
			if i == len(body)-1 {
				i++
			} else {
				continue
			}
		}

		_, word, _ := util.SplitPunctuation(body[start:i])
		start = i + 1
		if len(word) == 0 {
			goto next
		}

		switch word[0] {
		case '>':
			m := linkRegexp.FindSubmatch(word)
			if m == nil {
				goto next
			}
			var l common.Link
			l, err = parseLink(m)
			switch {
			case err != nil:
				return
			case l.ID != 0:
				if !haveLink[l.ID] {
					haveLink[l.ID] = true
					links = append(links, l)
				}
			}
		case '#':
			// Ignore hash commands in quotes, or #pyu/#pcount if board option disabled
			if body[lineStart] == '>' ||
				(len(word) > 1 && word[1] == 'p' && !pyu) {
				goto next
			}
			m := common.CommandRegexp.FindSubmatch(word)
			if m == nil {
				goto next
			}
			var c common.Command
			c, err = parseCommand(m[1], board, thread, id, ip, &isSlut, &isDead)
			switch err {
			case nil:
				com = append(com, c)
			case errTooManyRolls, errDieTooBig:
				// Consider command invalid
				err = nil
			default:
				return
			}
		}
	next:
		if b == '\n' {
			lineStart = i + 1
		}
	}

	// Handles claude commands
	m := common.ClaudeRegexp.FindSubmatch(body)
	if m != nil {
		claude = &common.ClaudeState{
			common.Waiting,
			string(m[1]),
			bytes.Buffer{},
		}
	}
	mediaCommands = []common.MediaCommand{}
	matches := common.MediaComRegexp.FindAllSubmatch(body, -1)
	if matches != nil {
		for _, m := range matches {
			var cmdStr string
			if m[1] != nil {
				cmdStr = string(m[1])
			} else if m[3] != nil {
				cmdStr = string(m[3])
			}

			var mediaCommand common.MediaCommand

			switch cmdStr {
			case "play":
				mediaCommand.Type = common.AddVideo
			case "remove":
				mediaCommand.Type = common.RemoveVideo
			case "skip":
				mediaCommand.Type = common.SkipVideo
			case "pause":
				mediaCommand.Type = common.Pause
			case "unpause":
				mediaCommand.Type = common.Play
			case "seek":
				mediaCommand.Type = common.SetTime
			case "clear":
				mediaCommand.Type = common.ClearPlaylist
			default:
				mediaCommand.Type = common.NoMediaCommand
			}
			mediaCommand.Args = string(m[2])

			// Append the command to the slice
			mediaCommands = append(mediaCommands, mediaCommand)
		}
	}
	return
}

// IsPrintable checks, if r is printable.
// Also accepts tabs, and newlines, if multiline = true.
func IsPrintable(r rune, multiline bool) error {
	switch r {
	case '\t', '\n', 12288: // Japanese space
		if !multiline {
			return common.ErrNonPrintable(r)
		}
	default:
		if !unicode.IsPrint(r) {
			return common.ErrNonPrintable(r)
		}
	}
	return nil
}

// IsPrintableString checks, if all of s is printable.
// Also accepts tabs, and newlines, if multiline = true.
func IsPrintableString(str string, multiline bool) error {
	gr := uniseg.NewGraphemes(str)
	for gr.Next() {
		runes := gr.Runes()
		if len(runes) == 1 {
			if err := IsPrintable(runes[0], multiline); err != nil {
				return err
			}
		}
	}
	return nil
}
