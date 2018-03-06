package redis_lock

import (
	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
	"time"
)

var (
	LockScript = redis.NewScript(`
    return redis.call('SET', KEYS[1], ARGV[1], 'NX', 'PX', ARGV[2])
  `)

	ExtendLockScript = redis.NewScript(`
    if redis.call("get", KEYS[1]) == ARGV[1] then
      return redis.call('PEXPIRE', KEYS[1], ARGV[2])
    else
      return 0
    end
  `)

	UnlockScript = redis.NewScript(`
    if redis.call("get", KEYS[1]) == ARGV[1] then
      return redis.call("DEL", KEYS[1])
    else
      return 0
    end
  `)
)

func LockSession() string {
	return uuid.NewV4().String()
}

func RedisLock(client *redis.Client, key, session string, exp time.Duration) (bool, error) {
	if res, err := LockScript.Run(client, []string{key}, session, exp.Nanoseconds()/int64(time.Millisecond)).Result(); err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	} else {
		return res.(string) == "OK", nil
	}
}

func RedisExtendLock(client *redis.Client, key, session string, exp time.Duration) (bool, error) {
	if res, err := ExtendLockScript.Run(client, []string{key}, session, exp.Nanoseconds()/int64(time.Millisecond)).Result(); err != nil {
		return false, err
	} else {
		return res.(int64) == 1, nil
	}
}

func RedisUnlock(client *redis.Client, key, session string) (bool, error) {
	if res, err := UnlockScript.Run(client, []string{key}, session).Result(); err != nil {
		return false, err
	} else {
		return res.(int64) == 1, nil
	}
}
