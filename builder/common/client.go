package common

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nebius/gosdk"
	"github.com/nebius/gosdk/auth"
	commonv1 "github.com/nebius/gosdk/proto/nebius/common/v1"
	"google.golang.org/grpc/codes"
)

func NewSDK(ctx context.Context, saConfig ServiceAccountConfig, parentID string) (*gosdk.SDK, error) {
	sa, err := resolveServiceAccountFromEnv(ctx, saConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service account from env: %w", err)
	}

	sdk, err := gosdk.New(
		ctx,
		gosdk.WithCredentials(gosdk.ServiceAccount(sa)),
		gosdk.WithDomain("api.testing.nebius.cloud:443"), // TODO: remove when in production
		gosdk.WithParentID(parentID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nebius sdk: %w", err)
	}

	return sdk, nil
}

func WaitFinishOperation(ctx context.Context, sdk *gosdk.SDK, operationID string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.After(10 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for operation %s to finish", operationID)
		case <-ticker.C:
			resp, err := sdk.Services().Compute().V1().Image().GetOperation(ctx, &commonv1.GetOperationRequest{Id: operationID})
			if err != nil {
				return fmt.Errorf("failed to get operation %s: %w", operationID, err)
			}

			if !resp.Done() {
				continue
			}

			if resp.Status() != nil && resp.Status().Code() == codes.OK {
				return nil
			}

			if resp.Status().Err() != nil {
				return fmt.Errorf("operation %s failed: %w", operationID, resp.Status().Err())
			}

			return fmt.Errorf("operation %s failed with unknown error", operationID)
		}
	}
}

func resolveServiceAccountFromEnv(ctx context.Context, saConfig ServiceAccountConfig) (auth.ServiceAccount, error) {
	privateKeyFile := os.Getenv(saConfig.PrivateKeyFileEnv)
	publicKeyID := os.Getenv(saConfig.PublicKeyIDEnv)
	accountID := os.Getenv(saConfig.AccountIDEnv)

	if privateKeyFile == "" {
		return auth.ServiceAccount{}, fmt.Errorf("environment variable %s is not set", saConfig.PrivateKeyFileEnv)
	}
	if publicKeyID == "" {
		return auth.ServiceAccount{}, fmt.Errorf("environment variable %s is not set", saConfig.PublicKeyIDEnv)
	}
	if accountID == "" {
		return auth.ServiceAccount{}, fmt.Errorf("environment variable %s is not set", saConfig.AccountIDEnv)
	}

	sa, err := auth.NewPrivateKeyFileParser(nil, privateKeyFile, publicKeyID, accountID).ServiceAccount(ctx)
	if err != nil {
		return auth.ServiceAccount{}, fmt.Errorf("failed to parse service account key: %w", err)
	}

	return sa, nil
}
