package nekotv

import (
	"errors"
	"github.com/bakape/meguca/pb"
	"math/rand"
)

type VideoList struct {
	items  []*pb.VideoItem
	Pos    int
	isOpen bool
}

func NewVideoList() *VideoList {
	return &VideoList{
		items:  []*pb.VideoItem{},
		Pos:    0,
		isOpen: true,
	}
}

func (v *VideoList) Length() int {
	return len(v.items)
}

func (v *VideoList) CurrentItem() (*pb.VideoItem, error) {
	if v.Pos < 0 || v.Pos >= len(v.items) {
		return nil, errors.New("invalid position")
	}
	return v.items[v.Pos], nil
}

func (v *VideoList) GetItem(i int) (*pb.VideoItem, error) {
	if i < 0 || i >= len(v.items) {
		return nil, errors.New("invalid index")
	}
	return v.items[i], nil
}

func (v *VideoList) SetItem(i int, item *pb.VideoItem) error {
	if i < 0 || i >= len(v.items) {
		return errors.New("invalid index")
	}
	v.items[i] = item
	return nil
}

func (v *VideoList) GetItems() []*pb.VideoItem {
	return v.items
}

func (v *VideoList) SetItems(items []*pb.VideoItem) {
	v.Clear()
	v.items = append(v.items, items...)
}

func (v *VideoList) SetPos(i int) {
	if i < 0 || i > len(v.items)-1 {
		i = 0
	}
	v.Pos = i
}

func (v *VideoList) Exists(f func(item *pb.VideoItem) bool) bool {
	for _, item := range v.items {
		if f(item) {
			return true
		}
	}
	return false
}

func (v *VideoList) FindIndex(f func(item *pb.VideoItem) bool) int {
	for i, item := range v.items {
		if f(item) {
			return i
		}
	}
	return -1
}

func (v *VideoList) AddItem(item *pb.VideoItem, atEnd bool) {
	if atEnd {
		v.items = append(v.items, item)
	} else {
		v.items = append(v.items[:v.Pos+1], v.items[v.Pos:]...)
		v.items[v.Pos+1] = item
	}
}

func (v *VideoList) SetNextItem(nextPos int) error {
	if nextPos < 0 || nextPos >= len(v.items) {
		return errors.New("invalid next position")
	}
	next := v.items[nextPos]
	v.items = append(v.items[:nextPos], v.items[nextPos+1:]...)
	if nextPos < v.Pos {
		v.Pos--
	}
	v.items = append(v.items[:v.Pos+1], v.items[v.Pos:]...)
	v.items[v.Pos+1] = next
	return nil
}

func (v *VideoList) ToggleItemType(pos int) error {
	if pos < 0 || pos >= len(v.items) {
		return errors.New("invalid position")
	}
	v.items[pos].IsTemp = !v.items[pos].IsTemp
	return nil
}

func (v *VideoList) RemoveItem(index int) error {
	if index < 0 || index >= len(v.items) {
		return errors.New("invalid index")
	}
	if index < v.Pos {
		v.Pos--
	}
	v.items = append(v.items[:index], v.items[index+1:]...)
	if v.Pos >= len(v.items) {
		v.Pos = 0
	}
	return nil
}

func (v *VideoList) SkipItem() (done bool) {
	if !v.items[v.Pos].IsTemp {
		v.Pos++
	} else {
		v.items = append(v.items[:v.Pos], v.items[v.Pos+1:]...)
	}
	if v.Pos >= len(v.items) {
		v.Pos = 0
		done = true
	}
	return
}

func (v *VideoList) Clear() {
	v.items = []*pb.VideoItem{}
	v.Pos = 0
}

func (v *VideoList) Shuffle() {
	current := v.items[v.Pos]
	v.items = append(v.items[:v.Pos], v.items[v.Pos+1:]...)
	shuffleArray(v.items)
	v.items = append([]*pb.VideoItem{current}, v.items...)
}

func shuffleArray(arr []*pb.VideoItem) {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}
