package websockets

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bakape/meguca/imager"
	"github.com/bakape/meguca/websockets/feeds"
	"github.com/go-playground/log"
	"math"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/bakape/meguca/common"
	"github.com/bakape/meguca/config"
	"github.com/bakape/meguca/db"
	"github.com/bakape/meguca/parser"
	"github.com/bakape/meguca/util"
)

var (
	errNoPostOpen    = errors.New("no post open")
	errEmptyPost     = errors.New("post body empty")
	errTooManyLines  = errors.New("too many lines in post body")
	errSpliceTooLong = errors.New("splice text too long")
	errSpliceNOOP    = errors.New("splice NOOP")
	errTextOnly      = errors.New("text only board")
	errHasImage      = errors.New("post already has image")
	preImageJson     = `{"type":"image","source":{"type":"base64","media_type":"image/webp","data":"`
	postImageJson    = `"}}`
)

// Error created, when client supplies invalid splice coordinates to server
type errInvalidSpliceCoords struct {
	body string
	req  spliceRequestString
}

func (e errInvalidSpliceCoords) Error() string {
	return fmt.Sprintf("invalid splice coordinates: %#v", e)
}

// Like spliceRequest, but with a string Text field. Used for internal
// conversions between []rune and string.
type spliceRequestString struct {
	spliceCoords
	Text string `json:"text"`
}

// Common part of a splice request and a splice response
type spliceCoords struct {
	//Start uint `json:"start"`
	//Len   uint `json:"len"`
	Start uint16 `json:"start"`
	Len   uint16 `json:"len"`
}

// Response to a spliceRequest. Sent to all listening clients.
type spliceMessage struct {
	ID uint64 `json:"id"`
	spliceRequestString
}

// Request or to replace the current line's text starting at an exact position
// in the current line
type spliceRequest struct {
	spliceCoords
	Text []rune
}

// Custom unmarshaling of string -> []rune
func (s *spliceRequest) UnmarshalJSON(buf []byte) error {
	var tmp spliceRequestString
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	*s = spliceRequest{
		spliceCoords: tmp.spliceCoords,
		Text:         []rune(tmp.Text),
	}
	return nil
}

// Custom marshaling of []rune -> string
func (s spliceRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(spliceRequestString{
		spliceCoords: s.spliceCoords,
		Text:         string(s.Text),
	})
}

// putPostIDToMsg embeds the postID into the provided byte array, representing it as a float64.
// Using float64 improves decoding performance in JavaScript clients
func putPostIDToMsg(msg []byte, postID uint64) {
	binary.LittleEndian.PutUint64(msg, math.Float64bits(float64(postID)))
}

// Append a rune to the body of the open post
func (c *Client) appendRune(data []byte) (err error) {
	has, err := c.hasPost()
	switch {
	case err != nil:
		return
	case !has:
		return
	case c.post.len+1 > common.MaxLenBody:
		return common.ErrBodyTooLong
	}

	var char rune
	//err = decodeMessage(data, &char)
	char, runeLen := utf8.DecodeRune(data)
	switch {
	case err != nil || char == utf8.RuneError || runeLen != len(data):
		return
	case char == 0:
		return common.ErrContainsNull
	case char == '\n':
		c.post.lines++
		if c.post.lines > common.MaxLinesBody {
			return errTooManyLines
		}
	}
	err = parser.IsPrintable(char, true)
	if err != nil {
		return
	}

	//var msg = [len(data) + 16]byte
	msg := make([]byte, len(data)+9)
	//write c.post.id into msg
	putPostIDToMsg(msg, c.post.id)
	copy(msg[8:], data)
	msg[8+len(data)] = byte(common.MessageAppend)
	//msg, err := common.EncodeMessage(
	//	common.MessageAppend,
	//	[2]uint64{c.post.id, uint64(char)},
	//)

	if err != nil {
		return
	}

	c.post.body = append(c.post.body, data...)
	c.post.len++
	return c.updateBodyBinary(msg, 1)
}

// Send message to thread update feed and writes the open post's buffer to the
// embedded database. Requires locking of c.openPost.
// n specifies the number of characters updated.
func (c *Client) updateBody(msg []byte, n int) error {
	c.feed.SetOpenBody(c.post.id, string(c.post.body), msg)
	c.incrementSpamScore(uint(n) * config.Get().CharScore)
	return db.SetOpenBody(c.post.id, c.post.body)
}

// Special case of UpdateBody for appending a single rune to a message
func (c *Client) updateBodyBinary(msg []byte, n int) error {
	c.feed.UpdateBody(c.post.id, string(c.post.body), msg)
	c.incrementSpamScore(uint(n) * config.Get().CharScore)
	return db.SetOpenBody(c.post.id, c.post.body)
}

// Increment the spam score for this IP by score. If the client requires a new
// solved captcha, send a notification.
func (c *Client) incrementSpamScore(score uint) {
	db.IncrementSpamScore(c.captchaSession, c.ip, score)
}

// Remove one character from the end of the line in the open post
func (c *Client) backspace() error {
	has, err := c.hasPost()
	switch {
	case err != nil:
		return err
	case !has:
		return nil
	case c.post.len == 0:
		return errEmptyPost
	}

	msg := make([]byte, 9)
	putPostIDToMsg(msg, c.post.id)
	msg[8] = byte(common.MessageBackspace)

	r, lastRuneLen := utf8.DecodeLastRune(c.post.body)
	c.post.body = c.post.body[:len(c.post.body)-lastRuneLen]
	if r == '\n' {
		c.post.lines--
	}
	c.post.len--

	return c.updateBodyBinary(msg, 1)
}

// Close an open post and parse the last line, if needed.
func (c *Client) closePost() (err error) {
	if c.post.id == 0 {
		return errNoPostOpen
	}
	var (
		links       []common.Link
		com         []common.Command
		postCommand *string
	)
	var claude *common.ClaudeState = nil
	if c.post.len != 0 {
		links, com, claude, postCommand, err = parser.ParseBody(c.post.body, c.post.board, c.post.op,
			c.post.id, c.ip, false)
		if err != nil {
			return
		}
		if len(links) != 0 &&
			bytes.Contains(c.post.body, []byte("#steal")) {
			var (
				from = links[len(links)-1].ID
				img  *common.Image
			)
			img, err = db.TransferImage(from, c.post.id, c.post.op)
			if err != nil {
				return
			}
			if img != nil {
				c.incrementSpamScore(config.Get().ImageScore)

				var msg []byte
				msg, err = common.EncodeMessage(
					common.MessageStoleImageFrom,
					from,
				)
				if err != nil {
					return
				}
				c.feed.Send(msg)

				msg, err = common.EncodeMessage(
					common.MessageStoleImageTo,
					struct {
						ID    uint64        `json:"id"`
						Image *common.Image `json:"image"`
					}{
						ID:    c.post.id,
						Image: img,
					},
				)
				if err != nil {
					return
				}
				c.feed.Send(msg)
			}
		}
	}
	claudeOk := true
	if claude != nil {
		claudeOk = db.CheckIfClaudeAllowed(c.ip)
		if !claudeOk {
			claude.Status = common.Error
			claude.Response.WriteString("Rate limit reached, try again later.")
		}
	}
	cid, err := db.ClosePost(c.post.id, c.post.op, string(c.post.body), links, com, claude)
	if err != nil {
		return
	}
	if claude != nil && claudeOk {
		//Include thumbnail of post
		id := c.post.id
		feed := c.feed
		imgSha1, err := db.GetPostSha1(id)
		var image *[]byte = nil
		if err == nil && imgSha1 != nil {
			*imgSha1 += ".webp"
			file := filepath.Join("images/thumb/", *imgSha1)
			cwd, _ := os.Getwd()
			log.Info("CWD: ", cwd)
			log.Info("img: ", file)
			fileData, err := os.ReadFile(file)
			if err == nil {
				size := len(preImageJson) + base64.StdEncoding.EncodedLen(len(fileData)) + len(postImageJson)
				buf := make([]byte, size)
				offset := 0
				offset += copy(buf[offset:], preImageJson)
				base64.StdEncoding.Encode(buf[offset:], fileData)
				offset = size - len(postImageJson)
				copy(buf[offset:], postImageJson)

				image = &buf
			}
		}
		go StreamMessages(Claude3Haiku, DefaultSystemPrompt, 255, claude, image,
			func() {
				claude.Status = common.Generating
				db.UpdateClaude(cid, claude)
			},
			func(token string) {
				feed.SendClaudeToken(id, token)
			},
			func() {
				isError := claude.Status == common.Error
				feed.SendClaudeComplete(id, isError, &claude.Response)
				db.UpdateClaude(cid, claude)
			})
	}
	if postCommand != nil {
		hasImage, err := c.hasImage()
		if err == nil && !hasImage {
			handlePostCommand(c.post.id, c.post.op, postCommand)
		}
	}
	c.post = openPost{}
	return
}
func handlePostCommand(id uint64, op uint64, input *string) {
	go func() {
		token, filename, err := imager.DownloadTikTok(*input)
		if err != nil {
			log.Error("Error downloading tiktok: ", *input)
			log.Error("Error: ", err)
			return
		}
		formatImageName(&filename)

		var msg []byte
		err = db.InTransaction(false, func(tx *sql.Tx) (err error) {
			msg, err = db.InsertImage(tx, id, token, filename,
				false)
			return
		})
		if err != nil {
			log.Error("Error downloading tiktok: ", *input)
			log.Error("Error: ", err)
			return
		}
		feeds.SendIfExists(op, func(feed *feeds.Feed) error {
			feed.InsertImage(op, false, common.PrependMessageType(common.MessageInsertImage, msg))
			return nil
		})

		return
	}()
}

// Splice the text in the open post
func (c *Client) spliceText(data []byte) error {
	if has, err := c.hasPost(); err != nil {
		return err
	} else if !has {
		return nil
	}

	var req spliceRequest
	//err := decodeMessage(data, &req)
	//if err != nil {
	//	return err
	//}
	decodeSpliceMessage(data, &req)
	err := parser.IsPrintableRunes(req.Text, true)
	if err != nil {
		return err
	}

	// Validate
	switch {
	case err != nil:
		return err
	case req.Start > common.MaxLenBody,
		req.Len > common.MaxLenBody,
		int(req.Start+req.Len) > c.post.len:
		return &errInvalidSpliceCoords{
			body: string(c.post.body),
			req: spliceRequestString{
				spliceCoords: spliceCoords{
					Start: req.Start,
					Len:   req.Len,
				},
				Text: string(req.Text),
			},
		}
	case req.Len == 0 && len(req.Text) == 0:
		return errSpliceNOOP // This does nothing. Client-side error.
	case len(req.Text) > common.MaxLenBody:
		return errSpliceTooLong // Nice try, kid
	}

	for _, r := range req.Text {
		if r == 0 {
			return common.ErrContainsNull
		}
	}

	var (
		old = []rune(string(c.post.body))
		end = append(req.Text, old[req.Start+req.Len:]...)
	)
	c.post.len += -int(req.Len) + len(req.Text)
	res := spliceMessage{
		ID: c.post.id,
		spliceRequestString: spliceRequestString{
			spliceCoords: req.spliceCoords,
			Text:         string(req.Text),
		},
	}

	// If it goes over the max post length, trim the end
	exceeding := c.post.len - common.MaxLenBody
	if exceeding > 0 {
		end = end[:len(end)-exceeding]
		res.Len = uint16(len(old[int(req.Start):]))
		res.Text = string(end)
		c.post.len = common.MaxLenBody
	}

	//msg, err := common.EncodeMessage(common.MessageSplice, res)
	msg, err := encodeSpliceMessage(res)
	if err != nil {
		return err
	}

	// Need to prevent modifications to the original slice, as there might be
	// concurrent reads in the update feed.
	c.post.body = util.CloneBytes(c.post.body)

	byteStartPos := 0
	for _, r := range old[:req.Start] {
		byteStartPos += utf8.RuneLen(r)
	}
	c.post.body = append(c.post.body[:byteStartPos], string(end)...)

	c.post.countLines()
	if c.post.lines > common.MaxLinesBody {
		return errTooManyLines
	}

	// +1, so you can't spam zero insert splices to infinity
	return c.updateBodyBinary(msg, len(res.Text)+1)
}

func encodeSpliceMessage(res spliceMessage) (msg []byte, err error) {
	//encode res.text to []byte
	size := 13 + len(res.Text)
	msg = make([]byte, size)
	putPostIDToMsg(msg, res.ID)
	binary.LittleEndian.PutUint16(msg[8:], res.Start)
	binary.LittleEndian.PutUint16(msg[10:], res.Len)
	copy(msg[12:], res.Text)
	msg[size-1] = uint8(common.MessageSplice)
	return msg, nil
}

func decodeSpliceMessage(data []byte, s *spliceRequest) {
	//Read two uints, then read the rest of the message as a string

	//Read the first uint
	s.Start = binary.LittleEndian.Uint16(data[:2])
	s.Len = binary.LittleEndian.Uint16(data[2:4])
	s.Text = []rune(string(data[4:]))
}

// Insert and image into an existing open post
// Note: Spam score is now incremented on image thumbnailing, not assignment to
// post.
func (c *Client) insertImage(data []byte) (err error) {
	// Ensure this can not be spammed, as this function can be resolved into a
	// NOP branch. It is generally good to have some spam protection either way.
	c.incrementSpamScore(config.Get().CharScore)

	has, err := c.hasPost()
	switch {
	case err != nil:
		return
	case !has:
		return errNoPostOpen
	}

	hasImage, err := c.hasImage()
	if err != nil {
		return
	}
	if hasImage {
		// Can be caused by network latency - NOP it
		return nil
	}

	// So the poster can reupload a new image, if

	var req ImageRequest
	err = decodeMessage(data, &req)
	if err != nil {
		return
	}

	if config.GetBoardConfigs(c.post.board).TextOnly {
		return errTextOnly
	}

	formatImageName(&req.Name)

	var msg []byte
	err = db.InTransaction(false, func(tx *sql.Tx) (err error) {
		msg, err = db.InsertImage(tx, c.post.id, req.Token, req.Name,
			req.Spoiler)
		return
	})
	if err != nil {
		return
	}
	c.post.isSpoilered = req.Spoiler
	c.feed.InsertImage(c.post.id, req.Spoiler,
		common.PrependMessageType(common.MessageInsertImage, msg))

	return
}

// Check, if post has an image. Done through the DB, so the poster can reupload,
// after his has been stolen.
func (c *Client) hasImage() (has bool, err error) {
	has, err = db.HasImage(c.post.id)
	if err != nil {
		return
	}
	if !has {
		// Allow respoilering
		c.post.isSpoilered = false
	}
	return
}

// Spoiler an already inserted image in an unclosed post
func (c *Client) spoilerImage() (err error) {
	// Ensure this can not be spammed, as this function can be resolved into a
	// NOP branch. It is generally good to have some spam protection either way.
	c.incrementSpamScore(config.Get().CharScore)

	has, err := c.hasPost()
	switch {
	case err != nil:
		return err
	case !has:
		return errNoPostOpen
	}

	hasImage, err := c.hasImage()
	if err != nil {
		return
	}
	if !hasImage {
		return errors.New("post does not have an image")
	}
	if c.post.isSpoilered {
		// Can be caused by network latency - NOP it
		return nil
	}

	err = db.SpoilerImage(c.post.id, c.post.op)
	if err != nil {
		return
	}
	msg, err := common.EncodeMessage(common.MessageSpoiler, c.post.id)
	if err != nil {
		return
	}
	c.feed.SpoilerImage(c.post.id, msg)

	return
}
