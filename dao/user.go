package dao

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/nilorg/naas/model"
	"github.com/nilorg/pkg/db"
)

// Userer ...
type Userer interface {
	SelectByUsername(ctx context.Context, username string) (mu *model.User, err error)
}

type user struct {
}

func (*user) SelectByUsername(ctx context.Context, username string) (mu *model.User, err error) {
	var gdb *gorm.DB
	gdb, err = db.FromContext(ctx)
	if err != nil {
		return
	}
	var dbResult model.User
	err = gdb.Where("username = ?", username).First(&dbResult).Error
	if err != nil {
		return
	}
	mu = &dbResult
	return
}