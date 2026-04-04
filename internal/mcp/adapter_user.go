package mcp

import (
	"context"
	"encoding/json"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func (a *Adapter) registerUserResources(reg *Registry) {
	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://profile/me",
		Name:        "profile",
		Title:       "Current User Profile",
		Description: "Profile information of the authenticated user.",
		MimeType:    "application/json",
	}, a.resourceProfile, authModel.ScopeProfileRead)
}

func (a *Adapter) resourceProfile(ctx context.Context, _ string) (*ResourceReadResult, error) {
	v := viewer.MustFromContext(ctx)
	user, err := a.userSvc.GetUserByID(v.UserID())
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(map[string]any{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"avatar":   user.Avatar,
		"is_admin": user.IsAdmin,
	})
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://profile/me", MimeType: "application/json", Text: string(data)}},
	}, nil
}
