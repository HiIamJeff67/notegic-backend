package repositories

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type UserQuotaRepositoryInterface interface {
	GetRoutineTaskCostUnitUsed(ctx context.Context, userId uuid.UUID, opts ...options.RepositoryOptions) (int64, *exceptions.Exception)
	InitializeMissing(ctx context.Context, now time.Time, opts ...options.RepositoryOptions) *exceptions.Exception
	InitializeMissingForUserIds(ctx context.Context, userIds []uuid.UUID, now time.Time, opts ...options.RepositoryOptions) *exceptions.Exception
	ResetDue(ctx context.Context, now time.Time, opts ...options.RepositoryOptions) (int64, *exceptions.Exception)
	ConsumeRoutineTaskCostUnits(ctx context.Context, consumptionInputs []inputs.ConsumeRoutineTaskCostUnitInput, opts ...options.RepositoryOptions) ([]uuid.UUID, *exceptions.Exception)
}

type UserQuotaRepository struct{}

func NewUserQuotaRepository() UserQuotaRepositoryInterface {
	return &UserQuotaRepository{}
}

func (r *UserQuotaRepository) GetRoutineTaskCostUnitUsed(
	ctx context.Context,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (int64, *exceptions.Exception) {
	if userId == uuid.Nil {
		return 0, exceptions.New(
			"InvalidDto",
			"UserQuota",
			"GetRoutineTaskCostUnitUsed",
			"User quota request is invalid",
			http.StatusBadRequest,
		)
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	routineTaskCostUnitUsed := []int64{}
	result := parsedOptions.DB.
		WithContext(ctx).
		Model(&schemas.UserQuota{}).
		Where("user_id = ?", userId).
		Limit(1).
		Pluck("routine_task_cost_unit_used", &routineTaskCostUnitUsed)
	if result.Error != nil {
		return 0, exceptions.New(
			"FailedToGet",
			"UserQuota",
			"GetRoutineTaskCostUnitUsed",
			"Failed to retrieve routine task quota usage",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if len(routineTaskCostUnitUsed) == 0 {
		return 0, nil
	}

	return routineTaskCostUnitUsed[0], nil
}

func (r *UserQuotaRepository) InitializeMissing(
	ctx context.Context,
	now time.Time,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	return r.initializeMissing(ctx, nil, now, opts...)
}

func (r *UserQuotaRepository) InitializeMissingForUserIds(
	ctx context.Context,
	userIds []uuid.UUID,
	now time.Time,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	if len(userIds) == 0 {
		return nil
	}

	return r.initializeMissing(ctx, userIds, now, opts...)
}

func (r *UserQuotaRepository) initializeMissing(
	ctx context.Context,
	userIds []uuid.UUID,
	now time.Time,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	if !parsedOptions.IsTransactionStarted {
		return exceptions.New(
			"TransactionRequired",
			"UserQuota",
			"InitializeMissing",
			"User quota initialization requires a transaction",
			http.StatusInternalServerError,
		)
	}

	userQuotaQuery := parsedOptions.DB.
		WithContext(ctx).
		Model(&schemas.UserQuota{}).
		Select("user_id")
	if len(userIds) > 0 {
		userQuotaQuery = userQuotaQuery.Where("user_id IN ?", userIds)
	}

	existingUserIds := []uuid.UUID{}
	if result := userQuotaQuery.Pluck("user_id", &existingUserIds); result.Error != nil {
		return exceptions.New(
			"FailedToInitialize",
			"UserQuota",
			"InitializeMissing",
			"Failed to find existing user quotas",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	userQuery := parsedOptions.DB.
		WithContext(ctx).
		Model(&schemas.User{}).
		Select("id, created_at")
	if len(userIds) > 0 {
		userQuery = userQuery.Where("id IN ?", userIds)
	}
	if len(existingUserIds) > 0 {
		userQuery = userQuery.Where("id NOT IN ?", existingUserIds)
	}

	users := []schemas.User{}
	if result := userQuery.Find(&users); result.Error != nil {
		return exceptions.New(
			"FailedToInitialize",
			"UserQuota",
			"InitializeMissing",
			"Failed to find users without quotas",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if len(users) == 0 {
		return nil
	}

	usersById := make(map[uuid.UUID]schemas.User, len(users))
	userIdsToInitialize := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		usersById[user.Id] = user
		userIdsToInitialize = append(userIdsToInitialize, user.Id)
	}

	type activeBillingPlan struct {
		UserId          uuid.UUID  `gorm:"column:user_id"`
		StartDate       time.Time  `gorm:"column:start_date"`
		NextBillingDate *time.Time `gorm:"column:next_billing_date"`
	}

	billingPlans := []activeBillingPlan{}
	if result := parsedOptions.DB.
		WithContext(ctx).
		Model(&schemas.UsersToBillingPlans{}).
		Select("user_id, start_date, next_billing_date").
		Where("user_id IN ?", userIdsToInitialize).
		Where("status = ?", enums.UsersToBillingPlansStatus_Active).
		Order("start_date DESC").
		Find(&billingPlans); result.Error != nil {
		return exceptions.New(
			"FailedToInitialize",
			"UserQuota",
			"InitializeMissing",
			"Failed to find active billing plans",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	latestBillingPlansByUserId := make(map[uuid.UUID]activeBillingPlan, len(billingPlans))
	for _, billingPlan := range billingPlans {
		if _, exists := latestBillingPlansByUserId[billingPlan.UserId]; !exists {
			latestBillingPlansByUserId[billingPlan.UserId] = billingPlan
		}
	}

	quotas := make([]schemas.UserQuota, 0, len(users))
	for _, userId := range userIdsToInitialize {
		user := usersById[userId]
		cycleStartedAt := user.CreatedAt
		nextResetAt := user.CreatedAt.AddDate(0, 0, 30)
		if billingPlan, exists := latestBillingPlansByUserId[userId]; exists {
			if billingPlan.NextBillingDate == nil {
				cycleStartedAt = billingPlan.StartDate
				nextResetAt = billingPlan.StartDate.AddDate(0, 0, 30)
			} else {
				cycleStartedAt = billingPlan.NextBillingDate.AddDate(0, 0, -30)
				nextResetAt = *billingPlan.NextBillingDate
			}
		}

		quotas = append(quotas, schemas.UserQuota{
			UserId:                  userId,
			RoutineTaskCostUnitUsed: 0,
			CycleStartedAt:          cycleStartedAt,
			NextResetAt:             nextResetAt,
			UpdatedAt:               now,
			CreatedAt:               now,
		})
	}

	result := parsedOptions.DB.
		WithContext(ctx).
		Model(&schemas.UserQuota{}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoNothing: true,
		}).
		CreateInBatches(&quotas, parsedOptions.BatchSize)
	if result.Error != nil {
		return exceptions.New(
			"FailedToInitialize",
			"UserQuota",
			"InitializeMissing",
			"Failed to initialize missing user quotas",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	return nil
}

func (r *UserQuotaRepository) ResetDue(
	ctx context.Context,
	now time.Time,
	opts ...options.RepositoryOptions,
) (int64, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	if !parsedOptions.IsTransactionStarted {
		return 0, exceptions.New(
			"TransactionRequired",
			"UserQuota",
			"ResetDue",
			"User quota reset requires a transaction",
			http.StatusInternalServerError,
		)
	}

	nextResetAt := now.AddDate(0, 0, 30)

	result := parsedOptions.DB.
		WithContext(ctx).
		Model(&schemas.UserQuota{}).
		Where("next_reset_at <= ?", now).
		Updates(map[string]any{
			"routine_task_cost_unit_used": 0,
			"cycle_started_at":            now,
			"next_reset_at":               nextResetAt,
			"updated_at":                  now,
		})
	if result.Error != nil {
		return 0, exceptions.New(
			"FailedToReset",
			"UserQuota",
			"ResetDue",
			"Failed to reset due user quotas",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	return result.RowsAffected, nil
}

func (r *UserQuotaRepository) ConsumeRoutineTaskCostUnits(
	ctx context.Context,
	consumptionInputs []inputs.ConsumeRoutineTaskCostUnitInput,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, *exceptions.Exception) {
	if len(consumptionInputs) == 0 {
		return []uuid.UUID{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	if !parsedOptions.IsTransactionStarted {
		return nil, exceptions.New(
			"TransactionRequired",
			"UserQuota",
			"ConsumeRoutineTaskCostUnits",
			"Routine task quota consumption requires a transaction",
			http.StatusInternalServerError,
		)
	}

	valuePlaceholders := make([]string, 0, len(consumptionInputs))
	valueArgs := make([]any, 0, len(consumptionInputs)*5)
	for _, consumptionInput := range consumptionInputs {
		if consumptionInput.RoutineTaskId == uuid.Nil || consumptionInput.UserId == uuid.Nil ||
			consumptionInput.CostUnit < 0 || consumptionInput.ScheduledAt.IsZero() {
			return nil, exceptions.New(
				"InvalidDto",
				"UserQuota",
				"ConsumeRoutineTaskCostUnits",
				"Routine task quota consumption input is invalid",
				http.StatusBadRequest,
			)
		}

		valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::uuid, ?::bigint, ?::integer, ?::timestamptz)")
		valueArgs = append(
			valueArgs,
			consumptionInput.RoutineTaskId,
			consumptionInput.UserId,
			consumptionInput.CostUnit,
			consumptionInput.Priority,
			consumptionInput.ScheduledAt,
		)
	}

	var routineTaskIds []uuid.UUID
	result := parsedOptions.DB.
		WithContext(ctx).
		Model(&schemas.UserQuota{}).
		Raw(fmt.Sprintf(`
		WITH consumption(routine_task_id, user_id, cost_unit, priority, scheduled_at) AS (
			VALUES %s
		),
		eligible_consumption AS (
			SELECT
				ranked_consumption.routine_task_id,
				ranked_consumption.user_id,
				ranked_consumption.cost_unit
			FROM (
				SELECT
					consumption.*,
					sum(cost_unit) OVER (
						PARTITION BY user_id
						ORDER BY priority DESC, scheduled_at ASC, routine_task_id ASC
					) AS accumulated_cost_unit
				FROM consumption
			) AS ranked_consumption
			JOIN "UserQuotaTable" AS user_quota ON user_quota.user_id = ranked_consumption.user_id
			JOIN "UserTable" AS user_table ON user_table.id = ranked_consumption.user_id
			JOIN "PlanLimitationTable" AS plan_limitation ON plan_limitation.key = user_table.plan
			WHERE user_quota.routine_task_cost_unit_used + accumulated_cost_unit <= plan_limitation.max_routine_task_cost_unit_count
		),
		consumption_totals AS (
			SELECT user_id, sum(cost_unit) AS cost_unit
			FROM eligible_consumption
			GROUP BY user_id
		),
		consumed AS (
			UPDATE "UserQuotaTable" AS user_quota
			SET
				routine_task_cost_unit_used = user_quota.routine_task_cost_unit_used + consumption_totals.cost_unit,
				updated_at = NOW()
			FROM consumption_totals
			JOIN "UserTable" AS user_table ON user_table.id = consumption_totals.user_id
			JOIN "PlanLimitationTable" AS plan_limitation ON plan_limitation.key = user_table.plan
			WHERE user_quota.user_id = consumption_totals.user_id
				AND user_quota.routine_task_cost_unit_used + consumption_totals.cost_unit <= plan_limitation.max_routine_task_cost_unit_count
			RETURNING user_quota.user_id
		)
		SELECT eligible_consumption.routine_task_id
		FROM eligible_consumption
		JOIN consumed ON consumed.user_id = eligible_consumption.user_id
		`, strings.Join(valuePlaceholders, ",")), valueArgs...).
		Scan(&routineTaskIds)
	if result.Error != nil {
		return nil, exceptions.New(
			"FailedToConsume",
			"UserQuota",
			"ConsumeRoutineTaskCostUnits",
			"Failed to consume routine task quota",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	return routineTaskIds, nil
}
