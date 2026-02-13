package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nebius/gosdk"
	"github.com/nebius/gosdk/auth"
	commonv1 "github.com/nebius/gosdk/proto/nebius/common/v1"
	"google.golang.org/grpc/codes"
)

func NewSDK(ctx context.Context, saConfig ServiceAccountConfig, parentID string, apiEndpoint string) (*gosdk.SDK, error) {
	sa, err := resolveServiceAccount(ctx, saConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve service account: %w", err)
	}

	opts := []gosdk.Option{
		gosdk.WithCredentials(gosdk.ServiceAccount(sa)),
		gosdk.WithParentID(parentID),
	}

	if apiEndpoint != "" {
		opts = append(opts, gosdk.WithDomain(apiEndpoint))
	}

	sdk, err := gosdk.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create nebius sdk: %w", err)
	}

	return sdk, nil
}

func WaitFinishOperation(ctx context.Context, sdk *gosdk.SDK, operationID string) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("timeout while waiting for operation %s to finish", operationID)
			}
			return ctx.Err()
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

func resolveServiceAccount(ctx context.Context, saConfig ServiceAccountConfig) (auth.ServiceAccount, error) {
	sa, err := auth.NewPrivateKeyFileParser(
		nil,
		saConfig.PrivateKeyFile,
		saConfig.PublicKeyID,
		saConfig.AccountID,
	).ServiceAccount(ctx)
	if err != nil {
		return auth.ServiceAccount{}, fmt.Errorf("failed to parse service account key: %w", err)
	}

	return sa, nil
}

func WaitFinishOperationWithTimeout(ctx context.Context, sdk *gosdk.SDK, operationID string, timeout time.Duration) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return WaitFinishOperation(ctxWithTimeout, sdk, operationID)
}
