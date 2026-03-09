package viewer

type NoopViewer struct{}

func NewNoopViewer() *NoopViewer               { return &NoopViewer{} }
func (v *NoopViewer) UserID() string           { return "" }
func (v *NoopViewer) WorkspaceID() string      { return "" }
func (v *NoopViewer) TenantID() string         { return "" }
func (v *NoopViewer) Roles() []string          { return []string{} }
func (v *NoopViewer) HasRole(role string) bool { return false }
func (v *NoopViewer) IsAdmin() bool            { return false }
func (v *NoopViewer) TraceID() string          { return "" }
func (v *NoopViewer) IsAuthenticated() bool    { return false }
func (v *NoopViewer) IsSystemContext() bool    { return false }
func (v *NoopViewer) ShouldAudit() bool        { return false }

var _ Context = (*NoopViewer)(nil)
