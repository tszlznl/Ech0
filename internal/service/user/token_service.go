package service

import (
	model "github.com/lin-snow/ech0/internal/model/user"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

func (userService *UserService) issueUserToken(user model.User) (string, error) {
	return jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
}
