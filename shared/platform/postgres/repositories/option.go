package repositories

import (
	"gorm.io/gorm"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"
)

const (
	LockingStrengthUpdate      = "UPDATE"
	LockingStrengthNoKeyUpdate = "NO KEY UPDATE"
	LockingStrengthShare       = "SHARE"
)

var defaultDB *gorm.DB

// SetDefaultDB configures the current runtime's default connection pool.
// Each runtime is a separate process, so this package-level default remains
// isolated per runtime while repository code stays platform-shared.
func SetDefaultDB(db *gorm.DB) {
	defaultDB = db
}

type RepositoryOptionFields struct {
	DB                   *gorm.DB
	IsTransactionStarted bool
	BatchSize            int
	AllowedPermissions   []enums.AccessControlPermission
	OnlyDeleted          types.Ternary
	LockingStrength      *string
}

type RepositoryOptions func(*RepositoryOptionFields)

func WithDB(db *gorm.DB) RepositoryOptions {
	return func(options *RepositoryOptionFields) {
		options.DB = db
	}
}

func WithIsTransactionStarted(isTransactionStarted bool) RepositoryOptions {
	return func(options *RepositoryOptionFields) {
		options.IsTransactionStarted = isTransactionStarted
	}
}

func WithTransactionDB(db *gorm.DB) RepositoryOptions {
	return func(options *RepositoryOptionFields) {
		options.DB = db
		options.IsTransactionStarted = true
	}
}

func WithBatchSize(batchSize int) RepositoryOptions {
	return func(options *RepositoryOptionFields) {
		options.BatchSize = batchSize
	}
}

func WithAllowedPermissions(allowedPermissions []enums.AccessControlPermission) RepositoryOptions {
	return func(options *RepositoryOptionFields) {
		options.AllowedPermissions = append([]enums.AccessControlPermission{}, allowedPermissions...)
	}
}

func WithOnlyDeleted(onlyDeleted types.Ternary) RepositoryOptions {
	return func(options *RepositoryOptionFields) {
		options.OnlyDeleted = onlyDeleted
	}
}

func WithLockingStrength(lockingStrength string) RepositoryOptions {
	return func(options *RepositoryOptionFields) {
		options.LockingStrength = &lockingStrength
	}
}

func (options RepositoryOptionFields) HasAllowedPermissions() bool {
	return options.AllowedPermissions != nil
}

func GetDefaultOptions() RepositoryOptionFields {
	return RepositoryOptionFields{
		DB:                   defaultDB,
		IsTransactionStarted: false,
		BatchSize:            1000,
		OnlyDeleted:          types.Ternary_Neutral,
	}
}

func ParseRepositoryOptions(opts ...RepositoryOptions) RepositoryOptionFields {
	options := GetDefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
