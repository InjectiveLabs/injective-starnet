package pulumi

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

const ROLLBACK_TIMEOUT = 15 * time.Minute

func Rollback(stack auto.Stack, err error) error {

	if err == nil {
		return nil
	}

	log.Printf("❌ One of the steps failed, rolling back the infrastructure: %v", err)
	ctx, cancel := context.WithTimeout(context.Background(), ROLLBACK_TIMEOUT)
	defer cancel()

	_, destroyErr := stack.Destroy(ctx)
	if destroyErr != nil {
		return fmt.Errorf("rollback failed: %w (original error: %v)", destroyErr, err)
	}

	return fmt.Errorf("rolled back due to error: %w", err)
}
