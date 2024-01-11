package vaultmanager

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/hashicorp/vault/api"
)

var apiClient *api.Client

func Open(cfg *api.Config) error {
	c, err := api.NewClient(cfg)
	if err != nil {
		return err
	}
	apiClient = c
	return nil
}

func GetClient() *api.Client { return apiClient }

func WriteKVv2(ctx context.Context, mountPath, secretPath string, data map[string]interface{}) error {
	_, err := GetClient().KVv2(mountPath).Put(ctx, secretPath, data)
	if err != nil {
		return err
	}
	return nil
}

func ReadKVv2(ctx context.Context, mountPath, secretPath string) (map[string]any, error) {
	s, err := GetClient().KVv2(mountPath).Get(ctx, secretPath)
	if err != nil {
		uwerr := errors.Unwrap(err)
		if uwerr != nil {
			return map[string]any{}, uwerr
		}
		return map[string]any{}, err
	}
	return s.Data, nil
}

func ManageTokenLifecycle(ctx context.Context) error {
	ta := apiClient.Auth().Token()

	s, err := ta.LookupSelf()
	if err != nil {
		slog.Error("error vault lookup self", slog.Any("err", err))
		return err
	}

	isRenewable, err := s.TokenIsRenewable()
	if err != nil {
		slog.Error("error check if vault token is renewable", slog.Any("err", err))
		return err
	}

	if !isRenewable {
		slog.Info("vault token is not renewable")
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			s, err = renew(ctx, ta, s)
			if err != nil {
				slog.Error("error renew vault token", slog.Any("err", err))
				return err
			}
		}
	}
}

func renew(ctx context.Context, tokenAuth *api.TokenAuth, s *api.Secret) (*api.Secret, error) {
	isRenewable, err := s.TokenIsRenewable()
	if err != nil {
		slog.Error("error check if vault token is renewable", slog.Any("err", err))
		return nil, err
	}

	if !isRenewable {
		slog.Info("vault token is not renewable")
		return nil, nil
	}

	dur, err := s.TokenTTL()
	if err != nil {
		return nil, err
	}

	slog.Info("vault token ttl", slog.Duration("ttl", dur), slog.Duration("renewIn", dur/2))

	timer := time.NewTimer(dur / 2)
	select {
	case <-ctx.Done():
		return nil, nil
	case <-timer.C:
		slog.Info("vault token ttl increment", slog.Duration("ttl", dur))
		newS, err := tokenAuth.RenewSelfWithContext(ctx, int(dur/2))
		if err != nil {
			slog.Error("error renew vault token", slog.Any("err", err))
			return nil, err
		}
		return newS, nil
	}
}
