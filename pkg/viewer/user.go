package viewer

type UserViewer struct {
	userID      string
	workspaceID string
	tenantID    string
	roles       []string
	traceID     string
}

type UserViewerOption func(*UserViewer)

func WithWorkspaceID(wsID string) UserViewerOption {
	return func(v *UserViewer) {
		v.workspaceID = wsID
	}
}

func WithTenantID(tenantID string) UserViewerOption {
	return func(v *UserViewer) {
		v.tenantID = tenantID
	}
}

func WithRoles(roles []string) UserViewerOption {
	return func(v *UserViewer) {
		v.roles = roles
	}
}

func WithTraceID(traceID string) UserViewerOption {
	return func(v *UserViewer) {
		v.traceID = traceID
	}
}

func NewUserViewer(userID string, opts ...UserViewerOption) *UserViewer {
	v := &UserViewer{
		userID: userID,
		roles:  []string{},
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func (v *UserViewer) UserID() string           { return v.userID }
func (v *UserViewer) WorkspaceID() string      { return v.workspaceID }
func (v *UserViewer) TenantID() string         { return v.tenantID }
func (v *UserViewer) Roles() []string          { return v.roles }
func (v *UserViewer) TraceID() string          { return v.traceID }
func (v *UserViewer) IsAuthenticated() bool    { return v.userID != "" }
func (v *UserViewer) IsSystemContext() bool    { return false }
func (v *UserViewer) ShouldAudit() bool        { return true }
func (v *UserViewer) IsAdmin() bool            { return v.HasRole("admin") }
func (v *UserViewer) HasRole(role string) bool {
	for _, r := range v.roles {
		if r == role {
			return true
		}
	}
	return false
}

var _ Context = (*UserViewer)(nil)
