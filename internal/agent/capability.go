// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"strconv"
	"strings"
)

// Model capabilities that change what a request may carry.
//
// These are read off the model id because that is all an operator configures,
// and the same id reaches us through every protocol: "gpt-5" behind an
// OpenAI-compatible proxy is still a reasoning model, "anthropic/claude-opus-5"
// on a router still rejects temperature. The checks are deliberately narrow —
// an id nobody recognises keeps the permissive default, because a proxy with a
// custom alias is far more common than a new model family.

// normalizeModelID lowercases the id and strips routing prefixes such as
// "openai/", "anthropic/" or Bedrock's "us.anthropic.", and Vertex's
// "@<date>" suffix, leaving the bare model name the rules are written against.
func normalizeModelID(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))
	if i := strings.LastIndex(m, "/"); i >= 0 {
		m = m[i+1:]
	}
	if i := strings.Index(m, "claude-"); i > 0 {
		m = m[i:]
	}
	if i := strings.Index(m, "@"); i >= 0 {
		m = m[:i]
	}
	return m
}

// isOpenAIReasoningModel reports the o-series and GPT-5 family. They reject
// temperature/top_p other than the default, and max_tokens in favour of
// max_completion_tokens.
func isOpenAIReasoningModel(model string) bool {
	m := normalizeModelID(model)
	for _, p := range []string{"o1", "o3", "o4", "gpt-5"} {
		if strings.HasPrefix(m, p) {
			return true
		}
	}
	return false
}

// claudeVersion parses "claude-<family>-<major>[-<minor>]" and the legacy
// "claude-<major>[-<minor>]-<family>". A dated snapshot suffix is not a minor
// version, which is why only one- and two-digit segments count.
func claudeVersion(model string) (major, minor int, ok bool) {
	m := normalizeModelID(model)
	if !strings.HasPrefix(m, "claude-") {
		return 0, 0, false
	}
	var nums []int
	for seg := range strings.SplitSeq(strings.TrimPrefix(m, "claude-"), "-") {
		if seg == "" || len(seg) > 2 {
			if len(nums) > 0 {
				break
			}
			continue
		}
		n, err := strconv.Atoi(seg)
		if err != nil {
			if len(nums) > 0 {
				break
			}
			continue
		}
		nums = append(nums, n)
		if len(nums) == 2 {
			break
		}
	}
	switch len(nums) {
	case 0:
		return 0, 0, false
	case 1:
		return nums[0], 0, true
	default:
		return nums[0], nums[1], true
	}
}

// acceptsSampling reports whether the model takes temperature at all.
//
// OpenAI's reasoning models fix it at 1; Claude removed sampling parameters
// from Opus 4.7 onward and from every 5-series model (Opus, Sonnet, Fable,
// Mythos), returning 400 when one is sent. Everything else still honours it.
func acceptsSampling(model string) bool {
	if isOpenAIReasoningModel(model) {
		return false
	}
	m := normalizeModelID(model)
	if strings.HasPrefix(m, "claude-fable") || strings.HasPrefix(m, "claude-mythos") {
		return false
	}
	major, minor, ok := claudeVersion(model)
	if !ok {
		return true
	}
	return major < 4 || (major == 4 && minor <= 6)
}

// temperatureFor returns the request's temperature when the model accepts one.
func (r Request) temperatureFor(model string) *float32 {
	if r.Temperature == nil || !acceptsSampling(model) {
		return nil
	}
	return r.Temperature
}
