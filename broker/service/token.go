package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jxskiss/errors"

	"github.com/jxskiss/nonamegw/pkg/errcode"
	"github.com/jxskiss/nonamegw/proto/cometsvc"
	"github.com/jxskiss/nonamegw/proto/data"
)

const (
	TokenVersion0   = "0"
	TokenExpiration = 10 * time.Minute
)

const (
	v0TokenLength = 33
)

type TokenDao interface {
	GetToken(ctx context.Context, token string) (*data.TokenInfo, error)
	SaveToken(ctx context.Context, info *data.TokenInfo, ttl time.Duration) error
}

type Signer interface {
	SignAuthToken(ctx context.Context, appId, userId, deviceId int64) (*cometsvc.AuthToken, error)
	DecodeAuthToken(ctx context.Context, token string) (*cometsvc.AuthToken, error)
}

func NewSigner(store TokenDao) Signer {
	return &signer{
		store: store,
	}
}

type signer struct {
	store TokenDao
}

func (s *signer) SignAuthToken(ctx context.Context, appId, userId, deviceId int64) (*cometsvc.AuthToken, error) {
	signTimeMsec := time.Now().UnixNano() / 1e6
	token := newTokenUuid()
	info := &data.TokenInfo{
		Id:           token,
		SignTimeMsec: signTimeMsec,
		AppId:        appId,
		UserId:       userId,
		DeviceId:     deviceId,
	}
	err := s.store.SaveToken(ctx, info, TokenExpiration)
	if err != nil {
		return nil, errors.AddStack(err)
	}
	result := &cometsvc.AuthToken{
		Token:        token,
		SignTimeMsec: signTimeMsec,
		AppId:        appId,
		UserId:       userId,
		DeviceId:     deviceId,
	}
	return result, nil
}

func (s *signer) DecodeAuthToken(ctx context.Context, token string) (*cometsvc.AuthToken, error) {
	if len(token) != v0TokenLength {
		return nil, errors.AddStack(errcode.IllegalAuthToken)
	}
	if !strings.HasPrefix(token, TokenVersion0) {
		return nil, errors.AddStack(errcode.UnknownTokenVersion)
	}

	info, err := s.store.GetToken(ctx, token)
	if err != nil {
		return nil, errors.AddStack(err)
	}
	result := &cometsvc.AuthToken{
		Token:        info.Id,
		SignTimeMsec: info.SignTimeMsec,
		AppId:        info.AppId,
		UserId:       info.UserId,
		DeviceId:     info.DeviceId,
	}
	return result, nil
}

func newTokenUuid() string {
	_uid := uuid.New().String()
	version := TokenVersion0
	return version + strings.Replace(_uid, "-", "", -1)
}
