package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/job"
	"github.com/redis/go-redis/v9"
)

const (
	redisKeyPrefix = "engage:job:"
	redisPending   = "engage:jobs:pending"
)

// RedisStore persists jobs in Redis (list + JSON documents).
type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(url string) (*RedisStore, error) {
	if url == "" {
		return nil, fmt.Errorf("redis url required")
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		opt = &redis.Options{Addr: url}
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &RedisStore{client: client}, nil
}

func (s *RedisStore) jobKey(id string) string {
	return redisKeyPrefix + id
}

func (s *RedisStore) Put(j *domain.Job) error {
	if j == nil || j.ID == "" {
		return fmt.Errorf("invalid job")
	}
	ctx := context.Background()
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	if err := s.client.Set(ctx, s.jobKey(j.ID), data, 0).Err(); err != nil {
		return err
	}
	if j.Status == domain.StatusPending {
		return s.client.LPush(ctx, redisPending, j.ID).Err()
	}
	return nil
}

func (s *RedisStore) Get(id string) (*domain.Job, bool) {
	ctx := context.Background()
	data, err := s.client.Get(ctx, s.jobKey(id)).Bytes()
	if err != nil {
		return nil, false
	}
	var j domain.Job
	if json.Unmarshal(data, &j) != nil {
		return nil, false
	}
	return &j, true
}

func (s *RedisStore) ListByStatus(status domain.Status, limit int) ([]*domain.Job, error) {
	ctx := context.Background()
	var ids []string
	if status == domain.StatusPending || status == "" {
		ids, _ = s.client.LRange(ctx, redisPending, 0, -1).Result()
	} else {
		keys, err := s.client.Keys(ctx, redisKeyPrefix+"*").Result()
		if err != nil {
			return nil, err
		}
		ids = make([]string, 0, len(keys))
		for _, k := range keys {
			ids = append(ids, k[len(redisKeyPrefix):])
		}
	}
	var out []*domain.Job
	for _, id := range ids {
		j, ok := s.Get(id)
		if !ok {
			continue
		}
		if status != "" && j.Status != status {
			continue
		}
		out = append(out, j)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	sortJobsByCreated(out)
	return trimLimit(out, limit), nil
}

func (s *RedisStore) ListPending() ([]*domain.Job, error) {
	return s.ListByStatus(domain.StatusPending, 0)
}

func (s *RedisStore) CountByStatus(status domain.Status) (int, error) {
	list, err := s.ListByStatus(status, 0)
	return len(list), err
}

func (s *RedisStore) TryClaim(id string) (*domain.Job, bool) {
	ctx := context.Background()
	claimKey := redisKeyPrefix + "claim:" + id
	ok, err := s.client.SetNX(ctx, claimKey, "1", 24*time.Hour).Result()
	if err != nil || !ok {
		return nil, false
	}
	j, got := s.Get(id)
	if !got || j.Status != domain.StatusPending {
		_, _ = s.client.Del(ctx, claimKey).Result()
		return nil, false
	}
	_, _ = s.client.LRem(ctx, redisPending, 0, id).Result()
	j.Status = domain.StatusRunning
	j.UpdatedAt = time.Now().UTC()
	if err := s.Put(j); err != nil {
		_, _ = s.client.Del(ctx, claimKey).Result()
		return nil, false
	}
	return j, true
}

// BlockingPop waits for a pending job id (used by worker).
func (s *RedisStore) BlockingPop(ctx context.Context, timeout time.Duration) (string, error) {
	res, err := s.client.BRPop(ctx, timeout, redisPending).Result()
	if err != nil {
		return "", err
	}
	if len(res) < 2 {
		return "", fmt.Errorf("empty pop")
	}
	return res[1], nil
}

func (s *RedisStore) Client() *redis.Client {
	return s.client
}
