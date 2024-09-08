// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Package redis implements service Registry and Discovery using redis.
package redis

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gsvc"
	"strings"
)

var (
	_         gsvc.Registry = &Registry{}
	usingList               = true
)

// Registry implements gsvc.Registry interface.
type Registry struct {
	Group string
}

const (
	// DefaultKeepAliveTTL is the default keepalive TTL.
	DefaultKeepAliveTTL = 10 // time.Second
)

// New creates and returns a new redis registry.
// Support Redis Address format: ip:port,ip:port...,ip:port@username:password
func New(group string) gsvc.Registry {
	if len(group) == 0 {
		group = "GRedisRpc"
	} else {
		group = fmt.Sprintf("%s_GRedisRpc", group)
	}
	return &Registry{Group: group}
}

func extractResponseToServices(key string, value string) ([]gsvc.Service, error) {
	service, err := gsvc.NewServiceWithKV(key, value)
	if err != nil {
		return nil, err
	}
	return []gsvc.Service{service}, nil
}

func fetchAllGroupKeys(ctx context.Context, group string) ([]string, error) {
	var keys = make([]string, 0)
	var err error
	if usingList {
		data, err := g.Redis().LRange(ctx, getRedisGroupKey(group), 0, -1)
		if err == nil && data != nil {
			keys = data.Strings()
		}
	} else {
		keys, err = g.Redis().Keys(ctx, group+"*")
		if err != nil {
			return nil, err
		}
	}
	return keys, err
}

func fetchServicesFromGroupKeys(ctx context.Context, group string, keys []string) ([]gsvc.Service, error) {
	var (
		services         []gsvc.Service
		servicePrefixMap = make(map[string]*Service)
	)
	for _, key := range keys {
		value, err := g.Redis().Get(ctx, key)
		if err == nil && value != nil {
			ss, _ := extractResponseToServices(strings.ReplaceAll(key, group, ""), value.String())
			if len(ss) > 0 {
				for _, s := range ss {
					//services = append(services, s)
					s := NewService(s)
					if v, ok := servicePrefixMap[s.GetPrefix()]; ok {
						v.Endpoints = append(v.Endpoints, s.GetEndpoints()...)
					} else {
						servicePrefixMap[s.GetPrefix()] = s
						services = append(services, s)
					}
				}
			}
		} else if value == nil || value.String() == "" {
			if usingList {
				_, _ = g.Redis().LRem(ctx, getRedisGroupKey(group), 0, key)
			}
		}
	}
	return services, nil
}
