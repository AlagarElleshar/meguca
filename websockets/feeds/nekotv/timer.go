package nekotv

import (
	"github.com/bakape/meguca/pb"
	"time"
)

type VideoTimer struct {
	isStarted      bool
	startTime      time.Time
	pauseStartTime time.Time
	rateStartTime  time.Time
	rate           float32
}

func NewVideoTimer() *VideoTimer {
	return &VideoTimer{
		rate: 1.0,
	}
}

func (t *VideoTimer) Start() {
	t.isStarted = true
	t.startTime = time.Now()
	t.pauseStartTime = time.Time{}
	t.rateStartTime = time.Now()
}

func (t *VideoTimer) Stop() {
	t.isStarted = false
	t.startTime = time.Time{}
	t.pauseStartTime = time.Time{}
}

func (t *VideoTimer) Pause() {
	t.startTime = t.startTime.Add(t.rateTime() - time.Duration(float32(t.rateTime())/t.rate))
	t.pauseStartTime = time.Now()
	t.rateStartTime = time.Time{}
}

func (t *VideoTimer) Play() {
	if !t.isStarted {
		t.Start()
	} else {
		t.startTime = t.startTime.Add(t.pauseTime())
		t.pauseStartTime = time.Time{}
		t.rateStartTime = time.Now()
	}
}

func (t *VideoTimer) GetTime() float32 {
	if t.startTime.IsZero() {
		return 0
	}
	elapsed := time.Since(t.startTime).Seconds()
	result := elapsed - t.rateTime().Seconds() + float64(t.rateTime())/float64(t.rate) - t.pauseTime().Seconds()
	return float32(result)
}

func (t *VideoTimer) SetTime(secs float32) {
	t.startTime = time.Now().Add(-time.Duration(secs * float32(time.Second)))
	t.rateStartTime = time.Now()
	if t.IsPaused() {
		t.Pause()
	}
}

func (t *VideoTimer) IsPaused() bool {
	return !t.isStarted || !t.pauseStartTime.IsZero()
}

func (t *VideoTimer) GetRate() float32 {
	return t.rate
}

func (t *VideoTimer) SetRate(rate float32) {
	if !t.IsPaused() {
		t.startTime = t.startTime.Add(t.rateTime() - time.Duration(float32(t.rateTime())/t.rate))
		t.rateStartTime = time.Now()
	}
	t.rate = rate
}

func (t *VideoTimer) pauseTime() time.Duration {
	if t.pauseStartTime.IsZero() {
		return 0
	}
	return time.Since(t.pauseStartTime) - t.rateTime()
}

func (t *VideoTimer) rateTime() time.Duration {
	if t.rateStartTime.IsZero() {
		return 0
	}
	return time.Since(t.rateStartTime) - t.pauseTime()
}

func (t *VideoTimer) GetTimeData() *pb.GetTimeEvent {
	return &pb.GetTimeEvent{
		Time:   t.GetTime(),
		Paused: t.IsPaused(),
		Rate:   t.GetRate(),
	}
}
