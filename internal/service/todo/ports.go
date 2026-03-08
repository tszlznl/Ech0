package service

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/todo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	GetTodoList(userid string) ([]model.Todo, error)
	AddTodo(userid string, todo *model.Todo) error
	UpdateTodo(userid string, id string) error
	DeleteTodo(userid string, id string) error
}

type CommonService = commonService.Service

type Repository interface {
	GetTodosByUserID(ctx context.Context, userid string) ([]model.Todo, error)
	CreateTodo(ctx context.Context, todo *model.Todo) error
	GetTodoByID(ctx context.Context, id string) (*model.Todo, error)
	UpdateTodo(ctx context.Context, todo *model.Todo) error
	DeleteTodo(ctx context.Context, id string) error
}
