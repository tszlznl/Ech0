// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package subscriber_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/event/subscriber"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/test/mocks/embeddingmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("embedding boom")

// TestHandleEchoCreated_Index covers the IndexEcho retry semantics driven by
// HandleEchoCreated: first-try success, fail-then-succeed (one retry), and the
// terminal all-fail path that warns and returns the last error after 3 tries.
//
// NOTE: assertions are call-count based (mock .Once()/.Times(3)) — never timing.
// The production withRetry sleeps between attempts, so the all-fail case is slow
// by design; we still assert on counts, not elapsed time.
func TestHandleEchoCreated_Index(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		idx := embeddingmock.NewMockIndexer(t)
		e := helpers.NewEcho(func(x *echoModel.Echo) { x.ID = "echo-c1" })
		idx.EXPECT().IndexEcho(mock.Anything, e).Return(nil).Once()

		ep := subscriber.NewEmbeddingProcessor(idx)
		require.NoError(t, ep.HandleEchoCreated(helpers.CtxAnonymous(), event.EchoCreated{Echo: e}))
	})

	t.Run("retry then success", func(t *testing.T) {
		idx := embeddingmock.NewMockIndexer(t)
		// First call fails, second call succeeds — withRetry must swallow the
		// transient error and return nil. Exactly two IndexEcho calls expected.
		idx.EXPECT().IndexEcho(mock.Anything, mock.Anything).Return(errBoom).Once()
		idx.EXPECT().IndexEcho(mock.Anything, mock.Anything).Return(nil).Once()

		ep := subscriber.NewEmbeddingProcessor(idx)
		require.NoError(t, ep.HandleEchoCreated(helpers.CtxAnonymous(), event.EchoCreated{Echo: helpers.NewEcho()}))
	})

	t.Run("all attempts fail returns last error after 3 tries", func(t *testing.T) {
		idx := embeddingmock.NewMockIndexer(t)
		// Exactly 3 attempts; a 4th would make the mock fail (count guard).
		idx.EXPECT().IndexEcho(mock.Anything, mock.Anything).Return(errBoom).Times(3)

		ep := subscriber.NewEmbeddingProcessor(idx)
		err := ep.HandleEchoCreated(helpers.CtxAnonymous(), event.EchoCreated{Echo: helpers.NewEcho()})
		require.ErrorIs(t, err, errBoom)
	})
}

// TestEmbeddingProcessor_Registrations checks the processor advertises its three
// echo-lifecycle subscriptions (build-only; not bound to a live bus here).
func TestEmbeddingProcessor_Registrations(t *testing.T) {
	idx := embeddingmock.NewMockIndexer(t)
	ep := subscriber.NewEmbeddingProcessor(idx)
	regs := ep.Registrations()
	require.Len(t, regs, 3)
	for i, r := range regs {
		assert.NotNil(t, r, "registration %d should be non-nil", i)
	}
}

// TestHandleEchoUpdated_Index proves HandleEchoUpdated routes through the same
// IndexEcho retry path.
func TestHandleEchoUpdated_Index(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		idx := embeddingmock.NewMockIndexer(t)
		e := helpers.NewEcho(func(x *echoModel.Echo) { x.ID = "echo-u1" })
		idx.EXPECT().IndexEcho(mock.Anything, e).Return(nil).Once()

		ep := subscriber.NewEmbeddingProcessor(idx)
		require.NoError(t, ep.HandleEchoUpdated(helpers.CtxAnonymous(), event.EchoUpdated{Echo: e}))
	})

	t.Run("all fail propagates error", func(t *testing.T) {
		idx := embeddingmock.NewMockIndexer(t)
		idx.EXPECT().IndexEcho(mock.Anything, mock.Anything).Return(errBoom).Times(3)

		ep := subscriber.NewEmbeddingProcessor(idx)
		require.ErrorIs(t, ep.HandleEchoUpdated(helpers.CtxAnonymous(), event.EchoUpdated{Echo: helpers.NewEcho()}), errBoom)
	})
}

// TestHandleEchoDeleted_Remove proves the delete path calls RemoveEcho with the
// echo ID and — crucially — does NOT retry on failure (single call, error
// propagated). The single-call expectation is the no-retry guard.
func TestHandleEchoDeleted_Remove(t *testing.T) {
	t.Run("success calls RemoveEcho with id", func(t *testing.T) {
		idx := embeddingmock.NewMockIndexer(t)
		var gotID string
		idx.EXPECT().RemoveEcho(mock.Anything, "echo-del-1").
			Run(func(_ context.Context, echoID string) { gotID = echoID }).
			Return(nil).Once()

		ep := subscriber.NewEmbeddingProcessor(idx)
		e := helpers.NewEcho(func(x *echoModel.Echo) { x.ID = "echo-del-1" })
		require.NoError(t, ep.HandleEchoDeleted(helpers.CtxAnonymous(), event.EchoDeleted{Echo: e}))
		assert.Equal(t, "echo-del-1", gotID)
	})

	t.Run("error propagated without retry", func(t *testing.T) {
		idx := embeddingmock.NewMockIndexer(t)
		// Exactly one call — if HandleEchoDeleted retried, the second call would
		// be unexpected and fail the mock.
		idx.EXPECT().RemoveEcho(mock.Anything, mock.Anything).Return(errBoom).Once()

		ep := subscriber.NewEmbeddingProcessor(idx)
		err := ep.HandleEchoDeleted(helpers.CtxAnonymous(), event.EchoDeleted{Echo: helpers.NewEcho()})
		require.ErrorIs(t, err, errBoom)
	})
}
