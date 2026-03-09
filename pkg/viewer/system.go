package viewer

type SystemViewer struct {
	name        string
	workspaceID string
	traceID     string
}

type SystemViewerOption func(*SystemViewer)

func WithSystemWorkspaceID(wsID string) SystemViewerOption {
	return func(v *SystemViewer) {
		v.workspaceID = wsID
	}
}

func WithSystemTraceID(traceID string) SystemViewerOption {
	return func(v *SystemViewer) {
		v.traceID = traceID
	}
}

func NewSystemViewer(name string, opts ...SystemViewerOption) *SystemViewer {
	v := &SystemViewer{name: name}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func (v *SystemViewer) UserID() string           { return "" }
func (v *SystemViewer) WorkspaceID() string      { return v.workspaceID }
func (v *SystemViewer) TenantID() string         { return "" }
func (v *SystemViewer) Roles() []string          { return []string{"system"} }
func (v *SystemViewer) HasRole(role string) bool { return role == "system" }
func (v *SystemViewer) IsAdmin() bool            { return false }
func (v *SystemViewer) TraceID() string          { return v.traceID }
func (v *SystemViewer) IsAuthenticated() bool    { return false }
func (v *SystemViewer) IsSystemContext() bool    { return true }
func (v *SystemViewer) ShouldAudit() bool        { return false }
func (v *SystemViewer) Name() string             { return v.name }

var _ Context = (*SystemViewer)(nil)
