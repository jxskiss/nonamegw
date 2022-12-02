package errcode

import "github.com/jxskiss/gopkg/errcode"

var reg = errcode.New()

var (
	IllegalAuthToken    = reg.Register(100_001, "illegal auth token")
	UnknownTokenVersion = reg.Register(100_002, "unknown token version")
)
