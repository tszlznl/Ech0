package viewer

type NoopViewer struct{}

func NewNoopViewer() *NoopViewer { return &NoopViewer{} }
func (v *NoopViewer) UserID() string {
	return ""
}

var _ Context = (*NoopViewer)(nil)
