package dataloaders

import (
	"time"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

const (
	loaderDelayOfUser     = 50 * time.Microsecond
	loaderDelayOfUserInfo = 50 * time.Millisecond
	loaderDelayOfBadge    = 50 * time.Microsecond
)

type Dataloaders struct {
	UserDataLoader     UserDataloaderInterface
	UserInfoDataLoader UserInfoDataloaderInterface
	BadgeDataLoader    BadgeDataloaderInterface
}

func NewDataloaders(coreAdapter *coreadapters.CoreAdapter) Dataloaders {
	return Dataloaders{
		UserDataLoader:     NewUserDataloader(coreAdapter),
		UserInfoDataLoader: NewUserInfoDataloader(coreAdapter),
		BadgeDataLoader:    NewBadgeDataloader(coreAdapter),
	}
}
