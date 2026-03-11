package viewer

import (
	"context"
	"net/http"
)

type contextKey struct{}

// WithContext returns a new context with the viewer attached.
func WithContext(ctx context.Context, v Context) context.Context {
	return context.WithValue(ctx, contextKey{}, v)
}

// FromContext extracts the viewer from the context.
func FromContext(ctx context.Context) (Context, bool) {
	if ctx == nil {
		return nil, false
	}
	v := ctx.Value(contextKey{})
	if v == nil {
		return nil, false
	}
	vc, ok := v.(Context)
	return vc, ok
}

// MustFromContext extracts the viewer from context and falls back to NoopViewer.
func MustFromContext(ctx context.Context) Context {
	if ctx == nil {
		return NewNoopViewer()
	}
	if v := ctx.Value(contextKey{}); v != nil {
		if vc, ok := v.(Context); ok && vc != nil {
			return vc
		}
	}
	return NewNoopViewer()
}

// WithRequest returns a new request with the viewer attached to context.
func WithRequest(req *http.Request, v Context) *http.Request {
	if req == nil {
		return nil
	}
	reqCtx := req.Context()
	if reqCtx == nil {
		reqCtx = context.Background()
	}
	return req.WithContext(WithContext(reqCtx, v))
}

// AttachToRequest updates request pointer with viewer-attached request context.
func AttachToRequest(req **http.Request, v Context) {
	if req == nil || *req == nil {
		return
	}
	*req = WithRequest(*req, v)
}
