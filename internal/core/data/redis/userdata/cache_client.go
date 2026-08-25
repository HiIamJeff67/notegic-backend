package userdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	coreconfig "github.com/HiIamJeff67/notegic-backend/internal/core/configs"
	cacheinputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata/inputs"
	redislibraries "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata/libraries"
)

type UserDataCache struct {
	Id                uuid.UUID         `json:"id"`
	PublicId          uuid.UUID         `json:"publicId"`
	Name              string            `json:"name"`
	DisplayName       string            `json:"displayName"`
	Email             string            `json:"email"`
	AccessToken       string            `json:"accessToken"`
	CSRFToken         string            `json:"csrfToken"`
	PreviousCSRFToken string            `json:"previousCSRFToken,omitempty"`
	Role              cenums.UserRole   `json:"role"`
	Plan              cenums.UserPlan   `json:"plan"`
	Status            cenums.UserStatus `json:"status"`
	AvatarURL         string            `json:"avatarURL"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
}

type UserDataCacheClient struct {
	cacheStore *UserDataCacheStore

	cacheExpiresIn                                       time.Duration
	maxRotationRetries                                   int
	batchCheckAndUpdateQuotasByFormattedKeysArgvPerKey   int
	batchCheckAndUpdateQuotasByFormattedKeyBaseNumOfArgv int
}

/* ============================== Constructor ============================== */

func NewUserDataCacheClient(config coreconfig.UserDataCacheConfig, cacheStore *UserDataCacheStore) *UserDataCacheClient {
	return &UserDataCacheClient{
		cacheStore: cacheStore,

		cacheExpiresIn:     config.CacheExpiresIn,
		maxRotationRetries: config.MaxRotationRetries,
		batchCheckAndUpdateQuotasByFormattedKeysArgvPerKey:   4,
		batchCheckAndUpdateQuotasByFormattedKeyBaseNumOfArgv: 4,
	}
}

/* ============================== Auxiliary Methods ============================== */

func (s *UserDataCacheClient) getRedisClient(identifier string) (*redis.Client, int, *cexceptions.Exception) {
	if s == nil || s.cacheStore == nil {
		return nil, 0, cexceptions.New(
			"CacheClientUnavailable",
			"Cache",
			"GetRedisClient",
			"User data cache client is unavailable",
			http.StatusInternalServerError,
			true,
		)
	}
	redisClient, shardIndex, err := s.cacheStore.ClientSet().ClientForKey(identifier)
	if err != nil {
		return nil, 0, cexceptions.New(
			"CacheClientUnavailable",
			"Cache",
			"GetRedisClient",
			"User data cache client is unavailable",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return redisClient, shardIndex, nil
}

func (s *UserDataCacheClient) formatUserDataKey(identifier string) string {
	return fmt.Sprintf("%s:%s", sredis.CachePurpose_UserData.String(), identifier)
}

/* ============================== Extend Methods ============================== */

func (s *UserDataCacheClient) Extend(identifier string) *cexceptions.Exception {
	redisClient, _, exception := s.getRedisClient(identifier)
	if exception != nil {
		return exception
	}

	updated, err := redisClient.Expire(s.formatUserDataKey(identifier), s.cacheExpiresIn).Result()
	if err != nil {
		return cexceptions.New(
			"FailedToExtendTTL",
			"Cache",
			"ExtendUserData",
			"Failed to extend cached user data",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if !updated {
		return cexceptions.New(
			"NotFound",
			"Cache",
			"ExtendUserData",
			"Cached user data was not found",
			http.StatusNotFound,
			true,
		)
	}

	return nil
}

/* ============================== Quota Method ============================== */

func (s *UserDataCacheClient) CheckAndUpdateQuota(
	identifier string,
	input cacheinputs.CheckAndUpdateUserQuotaInput,
) *cexceptions.Exception {
	redisClient, _, exception := s.getRedisClient(identifier)
	if exception != nil {
		return exception
	}

	arguments := []interface{}{
		"FCALL",
		redislibraries.CheckAndUpdateUserQuotaByFormattedKeyFunction,
		1,
		s.formatUserDataKey(identifier),
		input.Field,
		input.ChangeAmount,
		input.MaxLimit,
		int(time.Until(input.ExpiresIn).Seconds()),
	}
	if _, err := redisClient.Do(arguments...).Result(); err != nil {
		return cexceptions.New(
			"FailedToUpdate",
			"Cache",
			"CheckAndUpdateUserQuota",
			"Failed to update the user quota cache",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return nil
}

func (s *UserDataCacheClient) BestEffortBatchCheckAndUpdateQuotas(
	inputs []cacheinputs.BatchCheckAndUpdateUserQuotaInput,
) *cexceptions.Exception {
	if len(inputs) == 0 {
		return nil
	}

	inputsByShard := make(map[int][]cacheinputs.BatchCheckAndUpdateUserQuotaInput)
	for _, input := range inputs {
		_, shardIndex, err := s.cacheStore.ClientSet().ClientForKey(input.Identifier)
		if err != nil {
			continue
		}
		inputsByShard[shardIndex] = append(inputsByShard[shardIndex], input)
	}

	for shardIndex, groupedInputs := range inputsByShard {
		redisClient, exists := s.cacheStore.ClientSet().Client(shardIndex)
		if !exists {
			continue
		}

		keys := make([]interface{}, 0, len(groupedInputs))
		arguments := make([]interface{}, 0, len(groupedInputs)*s.batchCheckAndUpdateQuotasByFormattedKeysArgvPerKey)
		for _, input := range groupedInputs {
			keys = append(keys, s.formatUserDataKey(input.Identifier))
			arguments = append(arguments,
				input.Input.Field,
				input.Input.ChangeAmount,
				input.Input.MaxLimit,
				int(time.Until(input.Input.ExpiresIn).Seconds()),
			)
		}

		command := []interface{}{
			"FCALL",
			redislibraries.BestEffortBatchCheckAndUpdateUserQuotasByFormattedKeysFunction,
			len(keys),
		}
		command = append(command, keys...)
		command = append(command, arguments...)
		if _, err := redisClient.Do(command...).Result(); err != nil {
			return cexceptions.New(
				"FailedToUpdate",
				"Cache",
				"BestEffortBatchCheckAndUpdateUserQuotas",
				"Failed to update the user quota cache",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
	}

	return nil
}

func (s *UserDataCacheClient) BestEffortBatchCheckAndUpdateQuotasByIdentifier(
	identifier string,
	inputs []cacheinputs.CheckAndUpdateUserQuotaInput,
) *cexceptions.Exception {
	if len(inputs) == 0 {
		return nil
	}

	redisClient, _, exception := s.getRedisClient(identifier)
	if exception != nil {
		return exception
	}

	arguments := make([]interface{}, 0, len(inputs)*s.batchCheckAndUpdateQuotasByFormattedKeyBaseNumOfArgv)
	for _, input := range inputs {
		arguments = append(arguments,
			input.Field,
			input.ChangeAmount,
			input.MaxLimit,
			int(time.Until(input.ExpiresIn).Seconds()),
		)
	}

	command := []interface{}{
		"FCALL",
		redislibraries.BestEffortBatchCheckAndUpdateUserQuotasByFormattedKeyFunction,
		1,
		s.formatUserDataKey(identifier),
	}
	command = append(command, arguments...)
	if _, err := redisClient.Do(command...).Result(); err != nil {
		return cexceptions.New(
			"FailedToUpdate",
			"Cache",
			"BestEffortBatchCheckAndUpdateUserQuotasByIdentifier",
			"Failed to update the user quota cache",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return nil
}

/* ============================== CRUD Method ============================== */

func (s *UserDataCacheClient) Get(identifier string) (*UserDataCache, *cexceptions.Exception) {
	redisClient, serverNumber, exception := s.getRedisClient(identifier)
	if exception != nil {
		return nil, exception
	}

	cacheString, err := redisClient.Get(s.formatUserDataKey(identifier)).Result()
	if err != nil {
		return nil, cexceptions.New(
			"NotFound",
			"Cache",
			"GetUserData",
			"Cached user data was not found",
			http.StatusNotFound,
			true,
		).WithOrigin(err)
	}

	var userDataCache UserDataCache
	if err := json.Unmarshal([]byte(cacheString), &userDataCache); err != nil {
		return nil, cexceptions.New(
			"DeserializationFailed",
			"Cache",
			"GetUserData",
			"Failed to decode cached user data",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	slogs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully got cached user data from server %d", serverNumber))
	return &userDataCache, nil
}

func (s *UserDataCacheClient) Set(identifier string, userDataCache UserDataCache) *cexceptions.Exception {
	if userDataCache.PublicId == uuid.Nil ||
		strings.TrimSpace(userDataCache.Name) == "" ||
		strings.TrimSpace(userDataCache.DisplayName) == "" ||
		strings.TrimSpace(userDataCache.Email) == "" ||
		strings.TrimSpace(userDataCache.AccessToken) == "" ||
		!userDataCache.Role.IsValidEnum() ||
		!userDataCache.Plan.IsValidEnum() ||
		!userDataCache.Status.IsValidEnum() {
		return cexceptions.New(
			"InvalidCacheData",
			"Cache",
			"SetUserData",
			"Cached user data is invalid",
			http.StatusInternalServerError,
			true,
		)
	}

	redisClient, serverNumber, exception := s.getRedisClient(identifier)
	if exception != nil {
		return exception
	}

	value, err := json.Marshal(userDataCache)
	if err != nil {
		return cexceptions.New(
			"SerializationFailed",
			"Cache",
			"SetUserData",
			"Failed to encode cached user data",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := redisClient.Set(s.formatUserDataKey(identifier), string(value), s.cacheExpiresIn).Err(); err != nil {
		return cexceptions.New(
			"FailedToCreate",
			"Cache",
			"SetUserData",
			"Failed to store cached user data",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	slogs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully set cached user data in server %d", serverNumber))
	return nil
}

func (s *UserDataCacheClient) Update(identifier string, input cacheinputs.UpdateUserDataCacheInput) *cexceptions.Exception {
	userDataCache, exception := s.Get(identifier)
	if exception != nil {
		return exception
	}

	userDataCache.UpdatedAt = time.Now()
	if err := copier.Copy(userDataCache, &input); err != nil {
		return cexceptions.New(
			"SerializationFailed",
			"Cache",
			"UpdateUserData",
			"Failed to copy cached user data",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	redisClient, serverNumber, exception := s.getRedisClient(identifier)
	if exception != nil {
		return exception
	}

	value, err := json.Marshal(userDataCache)
	if err != nil {
		return cexceptions.New(
			"SerializationFailed",
			"Cache",
			"UpdateUserData",
			"Failed to encode cached user data",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := redisClient.Set(s.formatUserDataKey(identifier), string(value), s.cacheExpiresIn).Err(); err != nil {
		return cexceptions.New(
			"FailedToUpdate",
			"Cache",
			"UpdateUserData",
			"Failed to update cached user data",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	slogs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully updated cached user data in server %d", serverNumber))
	return nil
}

// RotateCSRFToken atomically replaces the current token only when it still
// matches expectedToken. The returned token is always the cache's current
// token, so concurrent requests can converge on the winner's value.
func (s *UserDataCacheClient) RotateCSRFToken(
	identifier string,
	expectedToken string,
	replacementToken string,
) (string, bool, *cexceptions.Exception) {
	if strings.TrimSpace(identifier) == "" ||
		strings.TrimSpace(expectedToken) == "" ||
		strings.TrimSpace(replacementToken) == "" {
		return "", false, cexceptions.New(
			"InvalidInput",
			"Cache",
			"RotateCSRFToken",
			"CSRF token rotation requires an identifier and both token values",
			http.StatusInternalServerError,
		)
	}

	redisClient, _, exception := s.getRedisClient(identifier)
	if exception != nil {
		return "", false, exception
	}

	key := s.formatUserDataKey(identifier)
	for attempt := 0; attempt < s.maxRotationRetries; attempt++ {
		currentToken := ""
		rotated := false
		err := redisClient.Watch(func(tx *redis.Tx) error {
			cacheString, err := tx.Get(key).Result()
			if err != nil {
				return err
			}

			var userDataCache UserDataCache
			if err := json.Unmarshal([]byte(cacheString), &userDataCache); err != nil {
				return err
			}
			currentToken = userDataCache.CSRFToken
			if currentToken != expectedToken {
				return nil
			}

			userDataCache.PreviousCSRFToken = currentToken
			userDataCache.CSRFToken = replacementToken
			userDataCache.UpdatedAt = time.Now()
			value, err := json.Marshal(userDataCache)
			if err != nil {
				return err
			}

			if _, err := tx.TxPipelined(func(pipe redis.Pipeliner) error {
				pipe.Set(key, string(value), s.cacheExpiresIn)
				return nil
			}); err != nil {
				return err
			}
			rotated = true
			return nil
		}, key)

		if err == redis.TxFailedErr {
			continue
		}
		if err != nil {
			return "", false, cexceptions.New(
				"FailedToRotate",
				"Cache",
				"RotateCSRFToken",
				"Failed to rotate the cached CSRF token",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
		if rotated {
			return replacementToken, true, nil
		}
		return currentToken, false, nil
	}

	userDataCache, exception := s.Get(identifier)
	if exception != nil {
		return "", false, exception
	}
	return userDataCache.CSRFToken, false, nil
}

func (s *UserDataCacheClient) Delete(identifier string) *cexceptions.Exception {
	redisClient, serverNumber, exception := s.getRedisClient(identifier)
	if exception != nil {
		return exception
	}

	if err := redisClient.Del(s.formatUserDataKey(identifier)).Err(); err != nil {
		return cexceptions.New(
			"FailedToDelete",
			"Cache",
			"DeleteUserData",
			"Failed to delete cached user data",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	slogs.NotegicLogger.Debug(context.Background(), fmt.Sprintf("Successfully deleted cached user data from server %d", serverNumber))
	return nil
}
