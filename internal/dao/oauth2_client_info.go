package dao

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/nilorg/naas/internal/model"
	"github.com/nilorg/naas/internal/pkg/contexts"
	"github.com/nilorg/naas/internal/pkg/random"
	"github.com/nilorg/sdk/cache"
	"gorm.io/gorm"
)

// OAuth2ClientInfoer oauth2 client 接口
type OAuth2ClientInfoer interface {
	SelectByClientID(ctx context.Context, clientID model.ID) (mc *model.OAuth2ClientInfo, err error)
	SelectByClientIDFromCache(ctx context.Context, clientID model.ID) (mc *model.OAuth2ClientInfo, err error)
	Insert(ctx context.Context, mc *model.OAuth2ClientInfo) (err error)
	Delete(ctx context.Context, id model.ID) (err error)
	DeleteByClientID(ctx context.Context, clientID model.ID) (err error)
	DeleteInClientIDs(ctx context.Context, clientIDs []model.ID) (err error)
	Update(ctx context.Context, mc *model.OAuth2ClientInfo) (err error)
}

type oauth2ClientInfo struct {
	cache cache.Cacher
}

func (*oauth2ClientInfo) formatOneKey(id model.ID) string {
	return fmt.Sprintf("id:%d", id)
}

func (o *oauth2ClientInfo) formatOneKeys(ids ...model.ID) (keys []string) {
	for _, id := range ids {
		keys = append(keys, o.formatOneKey(id))
	}
	return
}

func (*oauth2ClientInfo) SelectByClientID(ctx context.Context, clientID model.ID) (mc *model.OAuth2ClientInfo, err error) {
	var gdb *gorm.DB
	gdb, err = contexts.FromGormContext(ctx)
	if err != nil {
		return
	}
	var dbResult model.OAuth2ClientInfo
	err = gdb.Where("client_id = ?", clientID).First(&dbResult).Error
	if err != nil {
		return
	}
	mc = &dbResult
	return
}

func (o *oauth2ClientInfo) SelectByClientIDFromCache(ctx context.Context, clientID model.ID) (mc *model.OAuth2ClientInfo, err error) {
	mc = new(model.OAuth2ClientInfo)
	key := o.formatOneKey(clientID)
	err = o.cache.Get(ctx, key, mc)
	if err != nil {
		mc = nil
		if err == redis.Nil {
			mc, err = o.SelectByClientID(ctx, clientID)
			if err != nil {
				return
			}
			err = o.cache.Set(ctx, key, mc, random.TimeDuration(300, 600))
		}
	}
	return
}

func (*oauth2ClientInfo) Insert(ctx context.Context, mc *model.OAuth2ClientInfo) (err error) {
	var gdb *gorm.DB
	gdb, err = contexts.FromGormContext(ctx)
	if err != nil {
		return
	}
	err = gdb.Create(mc).Error
	return
}

func (o *oauth2ClientInfo) Delete(ctx context.Context, id model.ID) (err error) {
	var gdb *gorm.DB
	gdb, err = contexts.FromGormContext(ctx)
	if err != nil {
		return
	}
	err = gdb.Delete(model.OAuth2ClientInfo{}, id).Error
	if err != nil {
		return
	}
	err = o.cache.Remove(ctx, o.formatOneKey(id))
	return
}

func (o *oauth2ClientInfo) DeleteByClientID(ctx context.Context, clientID model.ID) (err error) {
	var gdb *gorm.DB
	gdb, err = contexts.FromGormContext(ctx)
	if err != nil {
		return
	}
	err = gdb.Where("client_id = ?", clientID).Delete(model.OAuth2ClientInfo{}).Error
	if err != nil {
		return
	}
	err = o.cache.Remove(ctx, o.formatOneKey(clientID))
	return
}

func (o *oauth2ClientInfo) DeleteInClientIDs(ctx context.Context, clientIDs []model.ID) (err error) {
	var gdb *gorm.DB
	gdb, err = contexts.FromGormContext(ctx)
	if err != nil {
		return
	}
	err = gdb.Where("client_id in (?)", clientIDs).Delete(model.OAuth2ClientInfo{}).Error
	if err != nil {
		return
	}
	err = o.cache.Remove(ctx, o.formatOneKeys(clientIDs...)...)
	return
}

func (o *oauth2ClientInfo) Update(ctx context.Context, mc *model.OAuth2ClientInfo) (err error) {
	var gdb *gorm.DB
	gdb, err = contexts.FromGormContext(ctx)
	if err != nil {
		return
	}
	err = gdb.Model(mc).Save(mc).Error
	if err != nil {
		return
	}
	err = o.cache.Remove(ctx, o.formatOneKey(mc.ClientID))
	return
}
