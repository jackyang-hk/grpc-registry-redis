// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package redis

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gsvc"
)

// Register registers `service` to Registry.
// Note that it returns a new Service if it changes the input Service with custom one.
func (r *Registry) Register(ctx context.Context, service gsvc.Service) (gsvc.Service, error) {
	service = NewService(service)
	var (
		key   = getRedisKey(r.Group, service.GetKey())
		value = service.GetValue()
	)
	if usingList {
		_, _ = g.Redis().RPush(ctx, getRedisGroupKey(r.Group), key)
	}
	_, err := g.Redis().Set(ctx, key, value)
	if err != nil {
		return nil, gerror.Wrapf(
			err,
			`redis put failed with key "%s", value "%s"`,
			key, value,
		)
	}
	_, _ = g.Redis().Expire(ctx, key, DefaultKeepAliveTTL)
	go r.doKeepAlive(key)
	return service, nil
}

// Deregister off-lines and removes `service` from the Registry.
func (r *Registry) Deregister(ctx context.Context, service gsvc.Service) error {
	if usingList {
		toRemoveKey := getRedisKey(r.Group, service.GetKey())
		_, _ = g.Redis().LRem(ctx, getRedisGroupKey(r.Group), 0, toRemoveKey)
	}
	_, err := g.Redis().Del(ctx, getRedisKey(r.Group, service.GetKey()))
	return err
}

// doKeepAlive continuously keeps alive the key from Redis.
func (r *Registry) doKeepAlive(key string) {
	var ctx = context.Background()
	for {
		_, err := g.Redis().Expire(ctx, key, DefaultKeepAliveTTL)
		if err != nil {
			return
		}
		time.Sleep((DefaultKeepAliveTTL / 2) * time.Second)
	}
}

func getRedisKey(group string, serviceKey string) string {
	return fmt.Sprintf("%s%s", group, serviceKey)
}

func getRedisGroupKey(group string) string {
	return fmt.Sprintf("%s_allkeys", group)
}
