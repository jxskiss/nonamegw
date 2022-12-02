package dao

import (
	"github.com/jxskiss/gopkg/exp/kvutil"
)

var km = kvutil.KeyManager{}

var (
	tokenKey = km.NewKey("token:{token}")

	userConnectionsHashKey = km.NewKey("u:h:{app_id}:{user_id}")
	userConnectionsZsetKey = km.NewKey("u:s:{app_id}:{user_id}")

	deviceConnectionsHashKey = km.NewKey("d:h:{app_id}:{device_id}")
	deviceConnectionsZsetKey = km.NewKey("d:s:{app_id}:{device_id}")
)
