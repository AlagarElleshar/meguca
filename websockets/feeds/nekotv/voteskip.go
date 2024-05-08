package nekotv

import (
	"github.com/bakape/meguca/pb"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

const halfMinute = 30 * time.Second
const fiveSeconds = 5 * time.Second
const tenSeconds = 10 * time.Second

type Poll struct {
	sync.Mutex
	YesVotes uint32
	NoVotes  uint32
	End      time.Time
	EndBy    time.Time
}

func NewPoll() *Poll {
	return &Poll{
		EndBy: time.Now().Add(halfMinute),
		End:   time.Now().Add(tenSeconds),
	}
}
func (p *Poll) UpdateEndTime() {
	newEnd := time.Now().Add(fiveSeconds)
	if newEnd.After(p.End) {
		p.End = newEnd
	}
	if p.End.After(p.EndBy) {
		p.End = p.EndBy
	}
}

func (p *Poll) VoteYes() {
	p.YesVotes++
}

func (p *Poll) VoteNo() {
	p.NoVotes++
}

func (p *Poll) Serialize(id uint64) (result []byte) {
	diff := p.End.Sub(time.Now())
	result, _ = proto.Marshal(&pb.VoteSkip{
		Post:     float64(id),
		YesVotes: p.YesVotes,
		NoVotes:  p.NoVotes,
		Time:     float32(diff.Seconds()),
	})
	return
}
