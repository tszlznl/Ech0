package visitor

import (
	"encoding/hex"
	"hash/fnv"
	"net/http"
	"time"
)

const (
	dateLayout = "2006-01-02"
	keepDays   = 7
	pvWindow   = 5 * time.Minute
	recordBuf  = 4096
)

type DayStat struct {
	Date string `json:"date"`
	PV   int64  `json:"pv"`
	UV   int64  `json:"uv"`
}

type Tracker struct {
	recordCh chan recordEvent
	queryCh  chan chan []DayStat
}

func NewTracker() *Tracker {
	s := &Tracker{
		recordCh: make(chan recordEvent, recordBuf),
		queryCh:  make(chan chan []DayStat),
	}
	go s.run()
	return s
}

func (s *Tracker) Record(r *http.Request, ip string) {
	if r == nil || r.Method != http.MethodGet {
		return
	}
	event := recordEvent{
		ipHash: hashIP(ip),
		at:     time.Now(),
	}
	// 非阻塞写入，队列满时丢弃事件，避免影响请求主路径。
	select {
	case s.recordCh <- event:
	default:
	}
}

func (s *Tracker) Last7Days() []DayStat {
	resp := make(chan []DayStat, 1)
	s.queryCh <- resp
	return <-resp
}

type recordEvent struct {
	ipHash string
	at     time.Time
}

type runtimeState struct {
	byDay    map[string]DayStat
	today    string
	todayUV  map[string]struct{}
	lastPVAt map[string]time.Time
}

func (s *Tracker) run() {
	state := runtimeState{
		byDay:    make(map[string]DayStat, keepDays),
		todayUV:  make(map[string]struct{}),
		lastPVAt: make(map[string]time.Time),
	}
	for {
		select {
		case ev := <-s.recordCh:
			recordIP(&state, ev)
		case resp := <-s.queryCh:
			now := time.Now()
			rotateDay(&state, now.Format(dateLayout))
			gc(&state, now)
			resp <- snapshotLast7Days(state.byDay, now)
		}
	}
}

func recordIP(state *runtimeState, ev recordEvent) {
	today := ev.at.Format(dateLayout)
	rotateDay(state, today)
	gc(state, ev.at)

	ipHash := ev.ipHash
	stat := state.byDay[today]
	stat.Date = today
	if canCountPV(state, ipHash, ev.at) {
		stat.PV++
	}
	if _, ok := state.todayUV[ipHash]; !ok {
		state.todayUV[ipHash] = struct{}{}
		stat.UV++
	}
	state.byDay[today] = stat
}

func rotateDay(state *runtimeState, today string) {
	if state.today == today {
		return
	}
	state.today = today
	state.todayUV = make(map[string]struct{})
}

func gc(state *runtimeState, now time.Time) {
	today := now.Format(dateLayout)
	cutoff := parseDate(today).AddDate(0, 0, -(keepDays - 1)).Format(dateLayout)
	for day := range state.byDay {
		if day < cutoff {
			delete(state.byDay, day)
		}
	}

	pvCutoff := now.Add(-24 * time.Hour)
	for ipHash, ts := range state.lastPVAt {
		if ts.Before(pvCutoff) {
			delete(state.lastPVAt, ipHash)
		}
	}
}

func parseDate(day string) time.Time {
	t, err := time.ParseInLocation(dateLayout, day, time.Local)
	if err != nil {
		return time.Now()
	}
	return t
}

func hashIP(ip string) string {
	if ip == "" {
		ip = "unknown"
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(ip))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func canCountPV(state *runtimeState, ipHash string, now time.Time) bool {
	last, ok := state.lastPVAt[ipHash]
	if ok && now.Sub(last) < pvWindow {
		return false
	}
	state.lastPVAt[ipHash] = now
	return true
}

func snapshotLast7Days(byDay map[string]DayStat, now time.Time) []DayStat {
	points := make([]DayStat, 0, keepDays)
	for i := keepDays - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format(dateLayout)
		stat, ok := byDay[day]
		if !ok {
			stat = DayStat{Date: day}
		}
		points = append(points, stat)
	}
	return points
}
