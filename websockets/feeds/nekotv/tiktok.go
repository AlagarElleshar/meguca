package nekotv

import (
	"errors"
	"fmt"
	"github.com/bakape/meguca/common"
	"github.com/bakape/meguca/imager"
	"github.com/bakape/meguca/pb"
	"strings"
)

func GetTiktokData(link string) (*pb.VideoItem, error) {
	id := common.GetTokID(link)
	if id == nil {
		return nil, errors.New("invalid link")
	}
	metadata, err := imager.GetTikTokMetadata(link)
	if err != nil {
		return nil, err
	}
	tokTitle := strings.Trim(metadata.Title, " ")
	title := ""
	if len(tokTitle) == 0 {
		title = fmt.Sprintf("@%s - %s", metadata.Author.UniqueID, metadata.ID)
	} else {
		title = fmt.Sprintf("@%s - %s", metadata.Author.UniqueID, metadata.Title)
	}
	return &pb.VideoItem{
		Url:      fmt.Sprintf("https://www.tiktok.com/@%s/video/%s", metadata.Author.UniqueID, metadata.ID),
		Title:    title,
		Id:       metadata.ID,
		Duration: float32(metadata.Duration + 1),
		Type:     pb.VideoType_TIKTOK,
	}, nil
}
