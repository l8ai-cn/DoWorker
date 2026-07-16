package preview

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrBootstrapConsumed = errors.New("preview bootstrap already consumed")
	ErrSessionInactive   = errors.New("preview session inactive")
	ErrStoreUnavailable  = errors.New("preview session store unavailable")
)

type SessionRecord struct {
	ID        string    `json:"id"`
	PodKey    string    `json:"pod_key"`
	UserID    int64     `json:"user_id"`
	OrgID     int64     `json:"org_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionStore struct {
	redis *redis.Client
}

func NewSessionStore(redisClient *redis.Client) *SessionStore {
	return &SessionStore{redis: redisClient}
}

func (s *SessionStore) Redeem(ctx context.Context, bootstrapID string, record SessionRecord, bootstrapTTL time.Duration) error {
	if s == nil || s.redis == nil {
		return ErrStoreUnavailable
	}
	if bootstrapID == "" || record.ID == "" || record.PodKey == "" || record.UserID == 0 || record.OrgID == 0 {
		return fmt.Errorf("preview session identity is required")
	}
	sessionTTL := time.Until(record.ExpiresAt)
	if bootstrapTTL <= 0 || sessionTTL <= 0 {
		return fmt.Errorf("preview session expiry is invalid")
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	result, err := s.redis.Eval(
		ctx,
		redeemScript,
		[]string{bootstrapKey(bootstrapID), sessionKey(record.ID), userKey(record.UserID)},
		record.PodKey,
		bootstrapTTL.Milliseconds(),
		payload,
		sessionTTL.Milliseconds(),
		record.ID,
	).Int()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}
	if result == 0 {
		return ErrBootstrapConsumed
	}
	return nil
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (SessionRecord, error) {
	if s == nil || s.redis == nil {
		return SessionRecord{}, ErrStoreUnavailable
	}
	payload, err := s.redis.Get(ctx, sessionKey(sessionID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return SessionRecord{}, ErrSessionInactive
	}
	if err != nil {
		return SessionRecord{}, fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}
	var record SessionRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return SessionRecord{}, fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}
	if time.Now().After(record.ExpiresAt) {
		return SessionRecord{}, ErrSessionInactive
	}
	return record, nil
}

func (s *SessionStore) RevokeUser(ctx context.Context, userID int64) error {
	if s == nil || s.redis == nil {
		return ErrStoreUnavailable
	}
	key := userKey(userID)
	ids, err := s.redis.SMembers(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}
	keys := make([]string, 0, len(ids)+1)
	for _, id := range ids {
		keys = append(keys, sessionKey(id))
	}
	keys = append(keys, key)
	if err := s.redis.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrStoreUnavailable, err)
	}
	return nil
}

func bootstrapKey(id string) string { return "preview:bootstrap:consumed:" + id }
func sessionKey(id string) string   { return "preview:session:" + id }
func userKey(userID int64) string   { return "preview:session:user:" + strconv.FormatInt(userID, 10) }

const redeemScript = `
if redis.call("EXISTS", KEYS[1]) == 1 then
  return 0
end
redis.call("PSETEX", KEYS[1], ARGV[2], ARGV[1])
redis.call("PSETEX", KEYS[2], ARGV[4], ARGV[3])
redis.call("SADD", KEYS[3], ARGV[5])
local userTTL = redis.call("PTTL", KEYS[3])
if userTTL < tonumber(ARGV[4]) then
  redis.call("PEXPIRE", KEYS[3], ARGV[4])
end
return 1
`
