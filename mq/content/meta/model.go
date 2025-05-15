package main

import (
	"context"
	"fansX/internal/model/database"
	"gorm.io/gorm"
)

type Handler struct {
	db      *gorm.DB
	factory StrategyFactory
}

type Strategy interface {
	Handle(ctx context.Context, db *gorm.DB, record *database.InvisibleContentInfo) error
}

type StrategyFactory struct {
	strategySet map[int]Strategy
}

type CheckStrategy struct {
}

type PassStrategy struct {
}

type DeleteStrategy struct {
}

type NotPassStrategy struct {
}
