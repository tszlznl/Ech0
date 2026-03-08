package service

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/todo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	GetTodoList(userid uint) ([]model.Todo, error)
	AddTodo(userid uint, todo *model.Todo) error
	UpdateTodo(userid uint, id int64) error
	DeleteTodo(userid uint, id int64) error
}

type CommonService = commonService.Service

type Repository interface {
	GetTodosByUserID(ctx context.Context, userid uint) ([]model.Todo, error)
	CreateTodo(ctx context.Context, todo *model.Todo) error
	GetTodoByID(ctx context.Context, id int64) (*model.Todo, error)
	UpdateTodo(ctx context.Context, todo *model.Todo) error
	DeleteTodo(ctx context.Context, id int64) error
}
