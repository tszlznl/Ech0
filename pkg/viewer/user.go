package viewer

type UserViewer struct {
	userID string
}

func NewUserViewer(userID string) *UserViewer { return &UserViewer{userID: userID} }
func (v *UserViewer) UserID() string          { return v.userID }

var _ Context = (*UserViewer)(nil)
