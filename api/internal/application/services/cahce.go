package services

import (
	"context"
	"time"
)

/*
 interface that defines the funcitons of what can be done with redis
 an actual service is not needed
*/

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
