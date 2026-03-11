package viewer

// NewSystemViewer returns a system-scoped viewer.
// For current simplified model, system and anonymous share the same behavior.
func NewSystemViewer() *NoopViewer { return NewNoopViewer() }
