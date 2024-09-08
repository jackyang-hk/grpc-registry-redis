// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package redis

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/gsvc"
)

var (
	_ gsvc.Watcher = &watcher{}
)

type watcher struct {
	group string
	key   string
	ctx   context.Context
}

func newWatcher(group string, key string) (*watcher, error) {
	w := &watcher{
		group: group,
		key:   key,
	}
	w.ctx = context.Background()
	w.key = key
	w.group = group
	return w, nil
}

// Proceed is used to watch the key.
func (w *watcher) Proceed() ([]gsvc.Service, error) {
	var keys = make([]string, 0)
	list, _ := fetchAllGroupKeys(w.ctx, w.group)
	for _, key := range list {
		if strings.Contains(key, w.key) {
			keys = append(keys, key)
		}
	}
	return fetchServicesFromGroupKeys(w.ctx, w.group, keys)
}

// Close is used to close the watcher.
func (w *watcher) Close() error {
	return nil
}
