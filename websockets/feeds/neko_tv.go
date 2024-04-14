package feeds

import (
	common "github.com/bakape/meguca/common"
	"github.com/bakape/meguca/pb"
	"github.com/bakape/meguca/websockets/feeds/nekotv"
	"github.com/golang/protobuf/proto"
	"sync"
)

type NekoTVFeed struct {
	baseFeed
	videoTimer  *nekotv.VideoTimer
	videoList   *nekotv.VideoList
	thread      uint64
	clientCount uint32
	mu          sync.Mutex
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
	f.thread = thread

	go func() {
		for {
			select {
			case c := <-f.add:
				f.addClient(c)
				f.clientCount += 1
				f.sendConnectedMessage(c)
			case c := <-f.remove:
				f.removeClient(c)
				f.clientCount -= 1
			}
		}
	}()

	return
}

func (e *NekoTVFeed) GetCurrentState() pb.ServerState {
	return pb.ServerState{
		VideoList:      e.videoList.GetItems(),
		IsPlaylistOpen: true,
		ItemPos:        0,
		Timer:          nil,
	}
}

func (f *NekoTVFeed) sendConnectedMessage(c common.Client) {
	conMessage := pb.ConnectedEvent{
		VideoList:      f.videoList.GetItems(),
		ItemPos:        int32(f.videoList.Pos),
		IsPlaylistOpen: false,
		GetTime:        f.videoTimer.GetTimeData(),
	}
	data, err := proto.Marshal(&conMessage)
	if err != nil {
		return
	}
	c.SendBinary(data)
}

func (f *NekoTVFeed) AddVideo(v *pb.VideoItem, atEnd bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.videoList.AddItem(v, atEnd)
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_AddVideoEvent{AddVideoEvent: &pb.AddVideoEvent{
		Item:  v,
		AtEnd: atEnd,
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
}

// RemoveVideo removes a video from the playlist
func (f *NekoTVFeed) RemoveVideo(url string) {
	f.mu.Lock()
	defer f.mu.Unlock()

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
}

// SkipVideo skips to the next video in the playlist
func (f *NekoTVFeed) SkipVideo(url string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.videoList.Length() == 0 {
		return
	}

	currentItem, err := f.videoList.CurrentItem()
	if err != nil || currentItem.Url != url {
		return
	}

	f.videoList.SkipItem()
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_SkipVideoEvent{SkipVideoEvent: &pb.SkipVideoEvent{
		Url: url,
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
}

// Pause pauses the current video
func (f *NekoTVFeed) Pause(time float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.videoList.Length() == 0 {
		return
	}

	f.videoTimer.SetTime(time)
	f.videoTimer.Pause()
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_PauseEvent{PauseEvent: &pb.PauseEvent{
		Time: time,
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
}

// Play plays the current video or resumes if paused
func (f *NekoTVFeed) Play(time float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.videoList.Length() == 0 {
		return
	}

	f.videoTimer.SetTime(time)
	f.videoTimer.Play()
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_PlayEvent{PlayEvent: &pb.PlayEvent{
		Time: time,
	}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
}

// SetTime sets the current playback time
func (f *NekoTVFeed) SetTime(time float32) {
	f.mu.Lock()
	defer f.mu.Unlock()

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
}

// ClearPlaylist clears the playlist
func (f *NekoTVFeed) ClearPlaylist() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.videoList.Clear()
	f.videoTimer.Stop()
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_ClearPlaylistEvent{ClearPlaylistEvent: &pb.ClearPlaylistEvent{}}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
}

func (f *NekoTVFeed) SendTimeSyncMessage() {
	msg := pb.WebSocketMessage{MessageType: &pb.WebSocketMessage_GetTimeEvent{GetTimeEvent: f.videoTimer.GetTimeData()}}
	data, _ := proto.Marshal(&msg)
	data = append(data, uint8(common.MessageNekoTV))
	f.sendToAllBinary(data)
}
