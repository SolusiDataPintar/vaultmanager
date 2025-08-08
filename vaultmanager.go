package vaultmanager

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/hashicorp/vault/api"
)

type VaultManager interface {
	GetClient() *api.Client
	WriteKVv2(ctx context.Context, mountPath, secretPath string, data map[string]interface{}) error
	ReadKVv2(ctx context.Context, mountPath, secretPath string) (map[string]any, error)
	ListWithContext(ctx context.Context, path string) (*api.Secret, error)
	DeleteKVv2(ctx context.Context, mountPath, secretPath string) error
	ManageTokenLifecycle(ctx context.Context) error
}

type vaultmanager struct{ apiClient *api.Client }

func NewVaultManager(cfg *api.Config) (VaultManager, error) {
	c, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &vaultmanager{apiClient: c}, nil

}

func (vm *vaultmanager) GetClient() *api.Client { return vm.apiClient }

func (vm *vaultmanager) WriteKVv2(ctx context.Context, mountPath, secretPath string, data map[string]interface{}) error {
	_, err := vm.GetClient().KVv2(mountPath).Put(ctx, secretPath, data)
	if err != nil {
		return err
	}
	return nil
}

func (vm *vaultmanager) ReadKVv2(ctx context.Context, mountPath, secretPath string) (map[string]any, error) {
	s, err := vm.GetClient().KVv2(mountPath).Get(ctx, secretPath)
	if err != nil {
		uwerr := errors.Unwrap(err)
		if uwerr != nil {
			return map[string]any{}, uwerr
		}
		return map[string]any{}, err
	}
	return s.Data, nil
}

func (vm *vaultmanager) DeleteKVv2(ctx context.Context, mountPath, secretPath string) error {
	return vm.GetClient().KVv2(mountPath).Delete(ctx, secretPath)
}

func (vm *vaultmanager) ListWithContext(ctx context.Context, path string) (*api.Secret, error) {
	return vm.GetClient().Logical().ListWithContext(ctx, path)
}

func (vm *vaultmanager) ManageTokenLifecycle(ctx context.Context) error {
	ta := vm.GetClient().Auth().Token()

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
		slog.Debug("vault token is not renewable")
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
		slog.Debug("vault token is not renewable")
		return nil, nil
	}

	ttl, err := s.TokenTTL()
	if err != nil {
		return nil, err
	}

	var dur time.Duration
	if ttl <= 1*time.Hour {
		dur = ttl / 2
	} else {
		dur = 1 * time.Hour
	}

	slog.Debug("vault token ttl", slog.Duration("ttl", dur), slog.Duration("renewIn", dur))

	timer := time.NewTimer(dur)
	select {
	case <-ctx.Done():
		return nil, nil
	case <-timer.C:
		slog.Debug("vault token ttl increment", slog.Duration("ttl", dur))
		newS, err := tokenAuth.RenewSelfWithContext(ctx, 0)
		if err != nil {
			slog.Error("error renew vault token", slog.Any("err", err))
			return nil, err
		}
		return newS, nil
	}
}
