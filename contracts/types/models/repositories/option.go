package repositories

import "gorm.io/gorm"

type RepositoryOptionFields struct {
	DB                   *gorm.DB
	IsTransactionStarted bool
	BatchSize            int
}

type RepositoryOptions func(*RepositoryOptionFields)

func WithDB(db *gorm.DB) RepositoryOptions {
	return func(options *RepositoryOptionFields) {
		options.DB = db
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

func ParseRepositoryOptions(opts ...RepositoryOptions) RepositoryOptionFields {
	options := RepositoryOptionFields{BatchSize: 1000}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
