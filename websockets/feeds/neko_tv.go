package feeds

import (
	"errors"
	"github.com/bakape/meguca/common"
	"github.com/bakape/meguca/db"
	"github.com/bakape/meguca/pb"
	"github.com/bakape/meguca/websockets/feeds/nekotv"
	"github.com/go-playground/log"
	"github.com/golang/protobuf/proto"
	"strconv"
	"strings"
	"sync"
	"time"
)

type NekoTVFeed struct {
	baseFeed
	videoTimer *nekotv.VideoTimer
	videoList  *nekotv.VideoList
	thread     uint64
	mu         sync.Mutex
	isRunning  bool
}

func NewNekoTVFeed() *NekoTVFeed {
	nf := NekoTVFeed{
		videoTimer: nekotv.NewVideoTimer(),
		videoList:  nekotv.NewVideoList(),
	}
	nf.baseFeed.init()
	return &nf
}

func (f *NekoTVFeed) start(thread uint64) (err error) {
	log.Info("Starting NekoTV feed for thread ", thread)
	f.thread = thread
	f.isRunning = true
	state, dbErr := db.GetNekoTVState(thread)
	if dbErr == nil {
		if state.Timer != nil {
			f.videoList.SetItems(state.VideoList)
			f.videoList.SetPos(int(state.ItemPos))
			f.videoTimer.FromProto(state.Timer)
		}
	}

	go func() {
		for {
			select {
			case c := <-f.add:
				f.addClient(c)
				f.sendConnectedMessage(c)
				log.Info("Client added")
			case c := <-f.remove:
				if f.removeClient(c) {
					f.isRunning = false
					log.Info("shutting down feed for thread ", thread)
					return
				}
			}
		}
	}()
	go func() {
		for {
			if !f.isRunning {
				return
			}
			item, err := f.videoList.CurrentItem()
			if err != nil {

				time.Sleep(1000 * time.Millisecond)
				continue
			}
			maxTime := item.Duration - 0.01
			if f.videoTimer.GetTime() > maxTime {
				f.videoTimer.Pause()
				f.videoTimer.SetTime(maxTime)
				skipUrl := item.Url
				time.AfterFunc(500*time.Millisecond, func() {
					if f.videoList.Length() == 0 {
						return
					}
					currentItem, err := f.videoList.CurrentItem()
					if err != nil || currentItem.Url != skipUrl {
						return
					}
					f.SkipVideo()
				})

				continue
			}

			f.SendTimeSyncMessage()
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	return
}

func (e *NekoTVFeed) GetCurrentState() *pb.ServerState {
	return &pb.ServerState{
		VideoList: e.videoList.GetItems(),
		ItemPos:   int32(e.videoList.Pos),
		Timer:     e.videoTimer.ToProto(),
	}
}

func (f *NekoTVFeed) WriteStateToDb() {
	db.SetNekoTVState(f.thread, f.GetCurrentState())
}

func (f *NekoTVFeed) sendConnectedMessage(c common.Client) {
	conMessage := pb.ConnectedEvent{
		VideoList:      f.videoList.GetItems(),
		ItemPos:        int32(f.videoList.Pos),
		IsPlaylistOpen: false,
		GetTime:        f.videoTimer.GetTimeData(),
	}
	wsMessage := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_ConnectedEvent{ConnectedEvent: &conMessage}}
	data, err := proto.Marshal(&wsMessage)
	data = append(data, uint8(common.MessageNekoTV))
	if err != nil {
		return
	}
	c.SendBinary(data)
	log.Info("Sent connected message to client.")
}

func (f *NekoTVFeed) AddVideo(v *pb.VideoItem, atEnd bool) {

	if f.videoList.Exists(func(item *pb.VideoItem) bool {
		return item.Url == v.Url
	}) {
		return
	}
	f.videoList.AddItem(v, atEnd)
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_AddVideoEvent{AddVideoEvent: &pb.AddVideoEvent{
		Item:  v,
		AtEnd: atEnd,
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
	if f.videoList.Length() == 1 {
		f.videoTimer.Start()
	}
	f.WriteStateToDb()
}

// RemoveVideo removes a video from the playlist
func (f *NekoTVFeed) RemoveVideo(url string) {

	index := f.videoList.FindIndex(func(item *pb.VideoItem) bool {
		return item.Url == url
	})
	if index == -1 {
		return
	}

	f.videoList.RemoveItem(index)
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_RemoveVideoEvent{RemoveVideoEvent: &pb.RemoveVideoEvent{
		Url: url,
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
	f.WriteStateToDb()
}

// SkipVideo skips to the next video in the playlist
func (f *NekoTVFeed) SkipVideo() {

	if f.videoList.Length() == 0 {
		return
	}

	currentItem, err := f.videoList.CurrentItem()
	if err != nil {
		return
	}

	finished := f.videoList.SkipItem()
	f.videoTimer.SetTime(0)
	if !finished {
		f.Play()
	} else {
		f.ClearPlaylist()
	}
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_SkipVideoEvent{SkipVideoEvent: &pb.SkipVideoEvent{
		Url: currentItem.Url,
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
}

// Pause pauses the current video
func (f *NekoTVFeed) Pause() {

	if f.videoList.Length() == 0 {
		return
	}

	f.videoTimer.Pause()
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_PauseEvent{PauseEvent: &pb.PauseEvent{
		Time: f.videoTimer.GetTime(),
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
	f.WriteStateToDb()
}

// Play plays the current video or resumes if paused
func (f *NekoTVFeed) Play() {

	if f.videoList.Length() == 0 {
		return
	}

	time := f.videoTimer.GetTime()
	f.videoTimer.Play()
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_PlayEvent{PlayEvent: &pb.PlayEvent{
		Time: time,
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
	f.WriteStateToDb()
}

// SetTime sets the current playback time
func (f *NekoTVFeed) SetTime(time float32) {

	if f.videoList.Length() == 0 {
		return
	}

	f.videoTimer.SetTime(time)
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_SetTimeEvent{SetTimeEvent: &pb.SetTimeEvent{
		Time: time,
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
	f.WriteStateToDb()
}

// UpdatePlaylist updates the playlist
func (f *NekoTVFeed) UpdatePlaylist() {
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_UpdatePlaylistEvent{UpdatePlaylistEvent: &pb.UpdatePlaylistEvent{
		VideoList: &pb.VideoItemList{
			Items: f.videoList.GetItems(),
		},
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
	f.WriteStateToDb()
}

// ClearPlaylist clears the playlist
func (f *NekoTVFeed) ClearPlaylist() {

	f.videoList.Clear()
	f.videoTimer.Stop()
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_ClearPlaylistEvent{ClearPlaylistEvent: &pb.ClearPlaylistEvent{}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
	f.WriteStateToDb()
}

func (f *NekoTVFeed) SendTimeSyncMessage() {
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_GetTimeEvent{GetTimeEvent: f.videoTimer.GetTimeData()}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
}
func parseTimestamp(timestamp string) (time float32, err error) {
	if strings.Contains(timestamp, ":") {
		parts := strings.Split(timestamp, ":")
		if len(parts) == 2 {
			var minutes int
			minutes, err = strconv.Atoi(parts[0])
			if err != nil {
				return
			}
			time = float32(minutes * 60)
			var seconds float64
			seconds, err = strconv.ParseFloat(parts[1], 32)
			time += float32(seconds)
			return
		} else if len(parts) == 3 {
			var hours int
			hours, err = strconv.Atoi(parts[0])
			if err != nil {
				return
			}
			time = float32(hours * 60 * 60)
			var minutes int
			minutes, err = strconv.Atoi(parts[0])
			if err != nil {
				return
			}
			time += float32(minutes * 60)
			var seconds float64
			seconds, err = strconv.ParseFloat(parts[1], 32)
			time += float32(seconds)
			return
		} else {
			err = errors.New("invalid timestamp")
			return
		}
	}
	var seconds float64
	seconds, err = strconv.ParseFloat(timestamp, 32)
	time = float32(seconds)
	return
}

func HandleMediaCommand(thread uint64, c *common.MediaCommand) {
	feeds.mu.RLock()
	ntv, ok := feeds.nekotvFeeds[thread]
	feeds.mu.RUnlock()
	if !ok {
		return
	}
	switch c.Type {
	case common.AddVideo:
		log.Info("Adding video to the playlist")
		videoData, err := nekotv.GetVideoData(c.Args)
		if err == nil {
			log.Infof("Video data retrieved: %v", videoData)
			ntv.AddVideo(&videoData, true)
		} else {
			log.Errorf("Failed to get video data: %v", err)
		}
		break
	case common.RemoveVideo:
		log.Infof("Removing video from the playlist: %s", c.Args)
		ntv.RemoveVideo(c.Args)
	case common.SkipVideo:
		log.Info("Skipping to the next video")
		ntv.SkipVideo()
	case common.Pause:
		log.Info("Pausing the video playback")
		ntv.Pause()
	case common.Play:
		log.Info("Resuming the video playback")
		ntv.Play()
	case common.SetTime:
		log.Infof("Setting video playback time to: %s", c.Args)
		time, err := parseTimestamp(c.Args)
		if err != nil {
			log.Errorf("Failed to parse timestamp: %v", err)
		} else {
			ntv.SetTime(time)
		}
	case common.ClearPlaylist:
		log.Info("Clearing the video playlist")
		ntv.ClearPlaylist()
	default:
		log.Warnf("Unknown media command type: %v", c.Type)
	}
}
