// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	model "github.com/lin-snow/ech0/internal/model/setting"
)

// fakeProvider 是一个脚本化的 Provider 测试替身：每次 Stream 调用依次消费 scripts[calls]
// 里的事件序列后关闭 channel，并记录每轮收到的 Request（用于断言 Tools / Messages）。
// Complete 不被 runLoop 触达，留空实现。
type fakeProvider struct {
	scripts [][]Event // 每次 Stream() 调用消费一段事件
	calls   int       // Stream 调用计数
	gotReqs []Request // 每轮捕获的 Request（断言强制收尾轮 Tools==nil、图片折入下一轮 Messages）
}

func (p *fakeProvider) Complete(_ context.Context, _ Request) (Response, error) {
	return Response{}, errors.New("not used")
}

func (p *fakeProvider) Stream(ctx context.Context, req Request) (<-chan Event, error) {
	p.gotReqs = append(p.gotReqs, req)
	var events []Event
	if p.calls < len(p.scripts) {
		events = p.scripts[p.calls]
	}
	p.calls++

	ch := make(chan Event)
	go func() {
		defer close(ch)
		for _, ev := range events {
			select {
			case ch <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

// drain 收集 out 的全部事件直到关闭。
func drain(out <-chan AgentEvent) []AgentEvent {
	var evs []AgentEvent
	for ev := range out {
		evs = append(evs, ev)
	}
	return evs
}

// countingTool 构造一个记录执行次数的工具；output / err 为固定返回。
func countingTool(name string, output ToolOutput, err error) (Tool, *int) {
	calls := 0
	t := Tool{
		Def: ToolDef{Name: name, Description: "test tool", Parameters: json.RawMessage(`{"type":"object"}`)},
		Execute: func(_ context.Context, _ json.RawMessage) (ToolOutput, error) {
			calls++
			return output, err
		},
	}
	return t, &calls
}

// toolCallEvent / textEvent / doneEvent / errEvent 是构造脚本事件的便捷函数。
func toolCallEvent(id, name, args string) Event {
	return Event{Kind: EventToolCall, ToolCall: ToolCall{ID: id, Name: name, Args: json.RawMessage(args)}}
}
func textEvent(s string) Event { return Event{Kind: EventTextDelta, Text: s} }
func doneEvent() Event         { return Event{Kind: EventDone} }
func errEvent(err error) Event { return Event{Kind: EventError, Err: err} }

// runLoopSync 同步跑一轮 runLoop，返回收集到的事件（断言 calls / gotReqs 用 provider 字段）。
func runLoopSync(ctx context.Context, provider Provider, req RunRequest) []AgentEvent {
	return drain(runChan(ctx, provider, req))
}

// kinds 抽取事件类型序列，便于断言顺序。
func kinds(evs []AgentEvent) []AgentEventKind {
	ks := make([]AgentEventKind, len(evs))
	for i, e := range evs {
		ks[i] = e.Kind
	}
	return ks
}

func countKind(evs []AgentEvent, k AgentEventKind) int {
	n := 0
	for _, e := range evs {
		if e.Kind == k {
			n++
		}
	}
	return n
}

// enabledSetting 是一个能通过 validate 的最小可用配置（runLoop 不实际调 validate，
// 但 RunRequest 需要一个 Setting；这里保持真实形状）。
func enabledSetting() model.AgentSetting {
	return model.AgentSetting{Enable: true, Protocol: "openai", Model: "gpt-test", ApiKey: "k"}
}

// 多轮 happy path：第一轮调工具，第二轮作答。事件序列应为 Searching→ToolResult→Delta→Done。
func TestRunLoop_MultiRoundHappyPath(t *testing.T) {
	tool, execs := countingTool("search_echos", ToolOutput{Content: "hit", Meta: "meta"}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{
		Setting: enabledSetting(),
		Tools:   []Tool{tool},
	})

	want := []AgentEventKind{AgentSearching, AgentToolResult, AgentDelta, AgentDone}
	got := kinds(evs)
	if len(got) != len(want) {
		t.Fatalf("event kinds = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event[%d] kind = %d, want %d (full: %v)", i, got[i], want[i], got)
		}
	}
	if *execs != 1 {
		t.Fatalf("tool executed %d times, want 1", *execs)
	}
	if fp.calls != 2 {
		t.Fatalf("provider.Stream called %d times, want 2", fp.calls)
	}
	if evs[2].Text != "answer" {
		t.Fatalf("delta text = %q, want %q", evs[2].Text, "answer")
	}
}

// 工具去重：跨轮同 Name+Args 的调用只执行一次，第二次走 seen 短路（不发 Searching/ToolResult）。
func TestRunLoop_ToolDedup(t *testing.T) {
	tool, execs := countingTool("search_echos", ToolOutput{Content: "hit"}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"same"}`), doneEvent()},
		{toolCallEvent("c2", "search_echos", `{"q":"same"}`), doneEvent()}, // 同 Name+Args → 去重
		{textEvent("done"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if *execs != 1 {
		t.Fatalf("tool executed %d times, want 1 (dedup)", *execs)
	}
	if n := countKind(evs, AgentSearching); n != 1 {
		t.Fatalf("AgentSearching count = %d, want 1", n)
	}
	if n := countKind(evs, AgentToolResult); n != 1 {
		t.Fatalf("AgentToolResult count = %d, want 1", n)
	}
	if n := countKind(evs, AgentDone); n != 1 {
		t.Fatalf("AgentDone count = %d, want 1", n)
	}
}

// maxRounds 用尽：工具轮跑满后强制一轮「不给工具」收尾，第二次 Stream 的 Tools 必须为 nil。
func TestRunLoop_MaxRoundsForcesFinalNoToolRound(t *testing.T) {
	tool, _ := countingTool("search_echos", ToolOutput{Content: "hit"}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()}, // 唯一一轮就调工具
		{textEvent("forced answer"), doneEvent()},                       // 强制收尾轮
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{
		Setting:   enabledSetting(),
		Tools:     []Tool{tool},
		MaxRounds: 1,
	})

	if fp.calls != 2 {
		t.Fatalf("provider.Stream called %d times, want 2 (1 tool round + 1 forced)", fp.calls)
	}
	if len(fp.gotReqs) != 2 {
		t.Fatalf("captured %d requests, want 2", len(fp.gotReqs))
	}
	if len(fp.gotReqs[0].Tools) != 1 {
		t.Fatalf("first round Tools len = %d, want 1", len(fp.gotReqs[0].Tools))
	}
	if fp.gotReqs[1].Tools != nil {
		t.Fatalf("forced final round must pass nil Tools, got %v", fp.gotReqs[1].Tools)
	}
	if evs[len(evs)-1].Kind != AgentDone {
		t.Fatalf("last event = %d, want AgentDone", evs[len(evs)-1].Kind)
	}
}

// 工具执行错误：包装成 tool 结果回喂模型自愈，不中止；下一轮正常作答收尾。
func TestRunLoop_ToolExecErrorFedBack(t *testing.T) {
	tool, execs := countingTool("search_echos", ToolOutput{}, errors.New("boom"))
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()},
		{textEvent("recovered"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if *execs != 1 {
		t.Fatalf("tool executed %d times, want 1", *execs)
	}
	if n := countKind(evs, AgentError); n != 0 {
		t.Fatalf("AgentError count = %d, want 0 (exec error must not abort)", n)
	}
	if n := countKind(evs, AgentSearching); n != 1 {
		t.Fatalf("AgentSearching count = %d, want 1", n)
	}
	if n := countKind(evs, AgentToolResult); n != 0 {
		t.Fatalf("AgentToolResult count = %d, want 0 (failed exec emits no ToolResult)", n)
	}
	if evs[len(evs)-1].Kind != AgentDone {
		t.Fatalf("last event = %d, want AgentDone", evs[len(evs)-1].Kind)
	}
}

// 未知工具：模型调了不存在的工具，短路（不发 Searching），不 panic，继续到下一轮收尾。
func TestRunLoop_UnknownTool(t *testing.T) {
	tool, _ := countingTool("search_echos", ToolOutput{Content: "hit"}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "nope", `{}`), doneEvent()}, // 未注册的工具名
		{textEvent("answer"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if n := countKind(evs, AgentSearching); n != 0 {
		t.Fatalf("AgentSearching count = %d, want 0 (unknown tool short-circuits)", n)
	}
	if evs[len(evs)-1].Kind != AgentDone {
		t.Fatalf("last event = %d, want AgentDone", evs[len(evs)-1].Kind)
	}
}

// 传输/协议错误：单 AgentError 后关闭，不再调 Stream。
func TestRunLoop_TransportError(t *testing.T) {
	fp := &fakeProvider{scripts: [][]Event{
		{errEvent(errors.New("network down"))},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting()})

	if len(evs) != 1 || evs[0].Kind != AgentError {
		t.Fatalf("events = %v, want single AgentError", kinds(evs))
	}
	if evs[0].Err == nil {
		t.Fatalf("AgentError must carry the error")
	}
	if fp.calls != 1 {
		t.Fatalf("provider.Stream called %d times, want 1", fp.calls)
	}
}

// ctx 取消：预取消后 runLoop 必须干净收口——关闭 out、不死锁、不泄漏 goroutine。
// （注意：emit 用 select{out / ctx.Done} 监听取消，预取消下二者皆 ready，是否还发出 Done 由
// Go 随机选择决定，故这里只断言「能终止」这一真实不变式，不断言具体事件。）
func TestRunLoop_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 预取消

	fp := &fakeProvider{scripts: [][]Event{
		{textEvent("never delivered"), doneEvent()},
	}}

	done := make(chan struct{})
	go func() {
		// drain 直到 out 关闭：若 runLoop 不收口，drain 会阻塞，由 go test 超时兜底报错。
		drain(runChan(ctx, fp, RunRequest{Setting: enabledSetting()}))
		close(done)
	}()

	<-done // 必须返回，证明 runLoop 在取消下关闭了 out（不死锁、不泄漏）。
}

// runChan 启动 runLoop 并返回其事件 channel（不立即 drain，供取消场景断言收口）。
func runChan(ctx context.Context, provider Provider, req RunRequest) <-chan AgentEvent {
	out := make(chan AgentEvent)
	go runLoop(ctx, provider, req, out)
	return out
}

// per-run 超时：Timeout>0 时 Run 给整轮套上超时；极短超时下应干净收口（关闭 out、不死锁）。
// （走真实 provider，但 ctx 立即过期使其在发起网络请求前即返回错误，故快速且无真实网络。）
func TestRun_TimeoutAborts(t *testing.T) {
	out, err := Run(context.Background(), RunRequest{
		Setting: enabledSetting(),
		Timeout: time.Nanosecond, // 立即过期
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	drain(out) // 必须关闭（超时 ctx 经 emit 收口），否则 go test 超时报错
}

// RunStrings 留空字段回退到中文默认；非空字段保留。
func TestRunStrings_WithDefaults(t *testing.T) {
	if got := (RunStrings{}).withDefaults(); got != defaultRunStrings {
		t.Fatalf("empty RunStrings should equal defaults, got %+v", got)
	}
	got := RunStrings{DedupNote: "EN dedup", ImageNote: "EN image"}.withDefaults()
	if got.DedupNote != "EN dedup" || got.ImageNote != "EN image" {
		t.Fatalf("provided fields must be preserved, got %+v", got)
	}
	if got.UnknownTool != defaultRunStrings.UnknownTool || got.ToolError != defaultRunStrings.ToolError {
		t.Fatalf("empty fields must fall back to defaults, got %+v", got)
	}
}

// 自定义 Strings.ImageNote 应出现在带图的下一轮 user 消息里（替代中文默认）。
func TestRunLoop_CustomImageNote(t *testing.T) {
	const enNote = "(custom english image note)"
	tool, _ := countingTool("search_echos", ToolOutput{Content: "hit", Images: []ImagePart{{MediaType: "image/png", Base64: "abc"}}}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{
		Setting: enabledSetting(),
		Tools:   []Tool{tool},
		Strings: RunStrings{ImageNote: enNote},
	})

	found := false
	for _, m := range fp.gotReqs[1].Messages {
		if m.Role == RoleUser && m.Content == enNote {
			found = true
		}
	}
	if !found {
		t.Fatalf("custom ImageNote %q should appear in next round messages", enNote)
	}
}

// 多模态：工具带出图片时，下一轮 Messages 应追加一条带图的 RoleUser（Content==toolImageNote）。
func TestRunLoop_ToolImageNoteAppended(t *testing.T) {
	img := ImagePart{MediaType: "image/png", Base64: "abc"}
	tool, _ := countingTool("search_echos", ToolOutput{Content: "hit", Images: []ImagePart{img}}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if len(fp.gotReqs) != 2 {
		t.Fatalf("captured %d requests, want 2", len(fp.gotReqs))
	}
	var found *Message
	for i := range fp.gotReqs[1].Messages {
		m := fp.gotReqs[1].Messages[i]
		if m.Role == RoleUser && m.Content == toolImageNote {
			found = &fp.gotReqs[1].Messages[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("next round messages should contain the toolImageNote user message")
	}
	if len(found.Images) != 1 || found.Images[0].Base64 != "abc" {
		t.Fatalf("toolImageNote message should carry the tool's image, got %+v", found.Images)
	}
}
