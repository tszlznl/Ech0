// Package viewer provides a unified viewer context abstraction.
package viewer

// Context defines the current viewer (user or system).
type Context interface {
	UserID() string
	WorkspaceID() string
	TenantID() string
	Roles() []string
	HasRole(role string) bool
	IsAdmin() bool
	TraceID() string
	IsAuthenticated() bool
	IsSystemContext() bool
	ShouldAudit() bool
}
