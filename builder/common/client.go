package common

import (
	"context"
	"fmt"
	"os"

	"github.com/nebius/gosdk"
	"github.com/nebius/gosdk/auth"
)

func NewSDK(ctx context.Context, saConfig ServiceAccountConfig) (*gosdk.SDK, error) {
	sa, err := resolveServiceAccountFromEnv(ctx, saConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service account from env: %w", err)
	}

	sdk, err := gosdk.New(
		ctx,
		gosdk.WithCredentials(gosdk.ServiceAccount(sa)),
		gosdk.WithDomain("api.testing.nebius.cloud:443"), // TODO: remove when in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nebius sdk: %w", err)
	}

	return sdk, nil
}

func resolveServiceAccountFromEnv(ctx context.Context, saConfig ServiceAccountConfig) (auth.ServiceAccount, error) {
	privateKeyFile := os.Getenv(saConfig.PrivateKeyFileEnv)
	publicKeyID := os.Getenv(saConfig.PublicKeyIDEnv)
	accountID := os.Getenv(saConfig.AccountIDEnv)

	if privateKeyFile == "" || publicKeyID == "" || accountID == "" {
		return auth.ServiceAccount{}, fmt.Errorf("missing env vars: %s / %s / %s",
			saConfig.PrivateKeyFileEnv,
			saConfig.PublicKeyIDEnv,
			saConfig.AccountIDEnv,
		)
	}

	sa, err := auth.NewPrivateKeyFileParser(nil, privateKeyFile, publicKeyID, accountID).ServiceAccount(ctx)
	if err != nil {
		return auth.ServiceAccount{}, fmt.Errorf("failed to parse service account key: %w", err)
	}

	return sa, nil
}
