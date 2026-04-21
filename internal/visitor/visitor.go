// Package visitor 实现进程内的 PV/UV 统计。
//
// 设计要点:
//   - Actor 模型:所有状态仅被 Tracker.run 这一个 goroutine 读写;外部通过 channel
//     交互,天然线程安全,无需互斥锁。
//   - 日粒度以 UTC 为准(dateLayout = YYYY-MM-DD),和 Tasker 调度器、DB 落盘键保持
//     一致;跨时区部署不会产生错位。
//   - 纯内存结构,真正的持久化由 internal/task.Tasker 定时 upsert 到 visitor_daily_stats
//     表完成;本包只负责统计与快照恢复。
package visitor

import (
	"encoding/hex"
	"hash/fnv"
	"net/http"
	"time"
)

const (
	// 统一使用的 UTC 日期格式,作为 byDay 的 key 与 DB 主键。
	dateLayout = "2006-01-02"
	// 内存中保留的最大天数(含当天),与前端"最近 7 天"视图对齐。
	keepDays = 7
	// 同一 IP 在该窗口内重复请求不再计入 PV,抑制刷新器/脚本造成的虚高。
	pvWindow = 5 * time.Minute
	// 记录事件缓冲队列容量。队列满时 Record 直接丢弃事件(见 Tracker.Record),
	// 以保证请求主路径永远不会被阻塞。
	recordBuf = 4096
)

// DayStat 是单日的 PV/UV 快照,对外返回 JSON 时使用。
type DayStat struct {
	Date string `json:"date"`
	PV   int64  `json:"pv"`
	UV   int64  `json:"uv"`
}

// Tracker 是 PV/UV 统计器的对外句柄。
// 所有字段都是 channel — 方法本身不持有任何可变状态,状态全在 run goroutine 里。
type Tracker struct {
	recordCh chan recordEvent    // 请求上报 → run
	queryCh  chan chan []DayStat // "最近 7 天" 查询:外部发一个回传 channel,run 填完再回
	todayCh  chan chan DayStat   // "今天一条" 查询,模式同上
	loadCh   chan loadRequest    // 启动时从 DB 回填历史快照
}

// NewTracker 启动后台 goroutine 并返回句柄。
// 注意:当前没有对应的 Stop/Close — tracker 的生命周期等同于进程。
func NewTracker() *Tracker {
	s := &Tracker{
		recordCh: make(chan recordEvent, recordBuf),
		queryCh:  make(chan chan []DayStat),
		todayCh:  make(chan chan DayStat),
		loadCh:   make(chan loadRequest),
	}
	go s.run()
	return s
}

// Record 上报一次访问。由 HTTP 中间件在每次 GET 页面请求时调用。
// 只统计 GET,避免 POST/OPTIONS 等非页面流量污染 PV。
func (s *Tracker) Record(r *http.Request, ip string) {
	if r == nil || r.Method != http.MethodGet {
		return
	}
	event := recordEvent{
		ipHash: hashIP(ip),
		at:     time.Now().UTC(),
	}
	// 非阻塞写入:队列满时直接丢弃,宁可丢统计也不拖慢请求。
	select {
	case s.recordCh <- event:
	default:
	}
}

// Last7Days 返回按日期升序排列的最近 7 天快照(今天在最后)。
// 缺失的日期会用 PV/UV 为 0 的占位行补齐,保证前端图表永远是 7 个点。
func (s *Tracker) Last7Days() []DayStat {
	resp := make(chan []DayStat, 1)
	s.queryCh <- resp
	return <-resp
}

// TodayStat 返回当前 UTC 日期的 PV/UV。Tasker 定时快照任务会调用它取值再落库。
func (s *Tracker) TodayStat() DayStat {
	resp := make(chan DayStat, 1)
	s.todayCh <- resp
	return <-resp
}

// LoadHistory 把 DB 中已持久化的历史快照回填进内存。
// 进程启动时由 Tasker 调用一次;方法同步等待 run goroutine 完成加载再返回。
func (s *Tracker) LoadHistory(history []DayStat) {
	done := make(chan struct{}, 1)
	s.loadCh <- loadRequest{
		history: history,
		done:    done,
	}
	<-done
}

type recordEvent struct {
	ipHash string
	at     time.Time
}

type loadRequest struct {
	history []DayStat
	done    chan struct{}
}

// runtimeState 是 run goroutine 独占的状态。
// 所有读写都发生在 run 的 select 里,因此无需加锁。
type runtimeState struct {
	// byDay 保存每个 UTC 日期的累计 PV/UV;最多保留 keepDays 天,更早的由 gc 清理。
	byDay map[string]DayStat
	// today 是 run 上次观察到的 UTC 日期;用于侦测跨日并重置 todayUV。
	today string
	// todayUV 是"今天"已出现过的 ipHash 集合,用于 UV 去重。跨日时会被清空。
	// 注意:该集合不落盘 — 重启后会从空集开始,可能造成极少量 UV 重复计数
	// (见 loadHistory 上的注释)。
	todayUV map[string]struct{}
	// lastPVAt 记录每个 ipHash 最后一次计入 PV 的时间,配合 pvWindow 做 PV 去重。
	lastPVAt map[string]time.Time
}

// run 是整个包的唯一"写者"。它串行处理四类消息:
//   - recordCh: 写入一次访问
//   - queryCh:  返回最近 7 天
//   - todayCh:  返回今天单日
//   - loadCh:   启动时回填历史
//
// 每次取出消息时都会先 rotateDay + gc,保证 UTC 跨日和过期数据被及时处理,
// 即使这段时间没有 Record 进来也不会残留脏状态。
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
			now := time.Now().UTC()
			rotateDay(&state, now.Format(dateLayout))
			gc(&state, now)
			resp <- snapshotLast7Days(state.byDay, now)
		case resp := <-s.todayCh:
			now := time.Now().UTC()
			today := now.Format(dateLayout)
			rotateDay(&state, today)
			gc(&state, now)
			stat, ok := state.byDay[today]
			if !ok {
				// 今天还没人访问,返回一个显式的零值(而不是零值的 DayStat{}),
				// 让调用方也能拿到正确的 Date 字段去 upsert。
				stat = DayStat{Date: today}
			}
			resp <- stat
		case req := <-s.loadCh:
			now := time.Now().UTC()
			rotateDay(&state, now.Format(dateLayout))
			gc(&state, now)
			loadHistory(&state, req.history, now.Format(dateLayout))
			req.done <- struct{}{}
		}
	}
}

// loadHistory 把 DB 中的历史快照写入内存。
//
// 关键点:今天那条**也**要回填。否则重启后 byDay[today] 为空,下一次 upsert 会用
// 0 覆盖 DB 里已经保存的今日累计,造成"重启一次当天清零一次"的 bug。
// 代价:todayUV 的 IP 集合不持久化,重启前后都访问过的 IP 会被重复计 UV;
// 相比 PV/UV 被清零,这个误差可以接受。
func loadHistory(state *runtimeState, history []DayStat, _ string) {
	for _, stat := range history {
		if stat.Date == "" {
			continue
		}
		state.byDay[stat.Date] = DayStat{
			Date: stat.Date,
			PV:   stat.PV,
			UV:   stat.UV,
		}
	}
}

// recordIP 处理一次访问事件:更新 PV/UV,并顺手做跨日检测和过期清理。
func recordIP(state *runtimeState, ev recordEvent) {
	atUTC := ev.at.UTC()
	today := atUTC.Format(dateLayout)
	rotateDay(state, today)
	gc(state, atUTC)

	ipHash := ev.ipHash
	stat := state.byDay[today]
	// 当 byDay[today] 还不存在时,上面的 map 取值返回 DayStat 零值;
	// 这里显式补上 Date,保证后续写回时 key 一致。
	stat.Date = today
	if canCountPV(state, ipHash, atUTC) {
		stat.PV++
	}
	// UV 用 todayUV 做集合去重,一个 IP 每天只计一次。
	if _, ok := state.todayUV[ipHash]; !ok {
		state.todayUV[ipHash] = struct{}{}
		stat.UV++
	}
	state.byDay[today] = stat
}

// rotateDay 侦测到跨 UTC 日时重置 todayUV(UV 集合按自然日归零)。
// byDay 本身不清 — 昨天的累计要继续保留,以便 Last7Days 返回历史值。
func rotateDay(state *runtimeState, today string) {
	if state.today == today {
		return
	}
	state.today = today
	state.todayUV = make(map[string]struct{})
}

// gc 清理超窗数据,防止 map 无限增长:
//  1. byDay 只保留最近 keepDays 天,更早的日期直接删除。
//  2. lastPVAt 中超过 24 小时没再触发的 IP 也删掉 — pvWindow 只有 5 分钟,
//     24 小时是一个足够宽松的上限,避免高峰期 IP 数量撑爆 map。
func gc(state *runtimeState, now time.Time) {
	nowUTC := now.UTC()
	today := nowUTC.Format(dateLayout)
	cutoff := parseDate(today).AddDate(0, 0, -(keepDays - 1)).Format(dateLayout)
	for day := range state.byDay {
		if day < cutoff {
			delete(state.byDay, day)
		}
	}

	pvCutoff := nowUTC.Add(-24 * time.Hour)
	for ipHash, ts := range state.lastPVAt {
		if ts.Before(pvCutoff) {
			delete(state.lastPVAt, ipHash)
		}
	}
}

// parseDate 把 "YYYY-MM-DD" 解析回 time.Time,明确使用 UTC 时区,
// 避免本机 time.Local 污染日期比较。解析失败时兜底返回当前 UTC 时间
// (这分支实际上不会被触发 — 所有调用方传入的都是 dateLayout 格式的字符串)。
func parseDate(day string) time.Time {
	t, err := time.ParseInLocation(dateLayout, day, time.UTC)
	if err != nil {
		return time.Now().UTC()
	}
	return t
}

// hashIP 把明文 IP 哈希为 16 字符十六进制字符串。
//
// 作用:
//  1. 脱敏 — 内存里只存哈希,不存 IP 原文。
//  2. 归一长度 — IPv4/IPv6 统一成定长 key,便于当 map 键。
//
// 选用 FNV-64 是出于性能考虑(非密码学安全但足够做去重计数);
// 空 IP 兜底成 "unknown",避免进一步的空字符串判空。
func hashIP(ip string) string {
	if ip == "" {
		ip = "unknown"
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(ip))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

// canCountPV 判断当前请求是否应计入 PV。
// 同一 IP 在 pvWindow(5 分钟)内的重复请求会被忽略 — 注意这里有一个边界效应:
// 只要更新了 lastPVAt,就算这次不算 PV,5 分钟的计时也会从这一次开始重新计算,
// 所以持续刷新的脚本永远打不开窗口。这是故意的。
func canCountPV(state *runtimeState, ipHash string, now time.Time) bool {
	last, ok := state.lastPVAt[ipHash]
	if ok && now.Sub(last) < pvWindow {
		return false
	}
	state.lastPVAt[ipHash] = now
	return true
}

// snapshotLast7Days 按日期升序(今天放最后)返回 keepDays 天的切片。
// 即使某天从未有访问、byDay 里也没 entry,也会补一个 PV=UV=0 的占位,
// 保证前端图表长度固定,不用在前端再做 padding。
func snapshotLast7Days(byDay map[string]DayStat, now time.Time) []DayStat {
	nowUTC := now.UTC()
	points := make([]DayStat, 0, keepDays)
	for i := keepDays - 1; i >= 0; i-- {
		day := nowUTC.AddDate(0, 0, -i).Format(dateLayout)
		stat, ok := byDay[day]
		if !ok {
			stat = DayStat{Date: day}
		}
		points = append(points, stat)
	}
	return points
}
