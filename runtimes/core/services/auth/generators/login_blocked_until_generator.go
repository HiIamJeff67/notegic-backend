package generators

import (
	"net/http"
	"time"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type LoginBlockedUntilGeneratorInterface interface {
	GenerateNextByLoginCount(loginCount int32) (*time.Time, *cexceptions.Exception)
}

type LoginBlockedUntilGenerator struct {
	blockDurationByLoginCount map[int32]time.Duration
}

func NewLoginBlockedUntilGenerator() LoginBlockedUntilGeneratorInterface {
	return &LoginBlockedUntilGenerator{
		blockDurationByLoginCount: map[int32]time.Duration{
			3:  5 * time.Minute,
			5:  15 * time.Minute,
			7:  30 * time.Minute,
			10: time.Hour,
			15: 6 * time.Hour,
			20: 24 * time.Hour,
			30: 7 * 24 * time.Hour,
		},
	}
}

func (g *LoginBlockedUntilGenerator) GenerateNextByLoginCount(
	loginCount int32,
) (*time.Time, *cexceptions.Exception) {
	if loginCount < 0 {
		return nil, cexceptions.New(
			"InvalidLoginCount",
			"Auth",
			"GetLoginBlockedUntil",
			"Login count is invalid",
			http.StatusInternalServerError,
			true,
		)
	}

	var blockDuration time.Duration
	var matchedLoginCount int32
	for count, duration := range g.blockDurationByLoginCount {
		if loginCount >= count && count > matchedLoginCount {
			matchedLoginCount = count
			blockDuration = duration
		}
	}
	if matchedLoginCount == 0 {
		return nil, nil
	}

	blockedUntil := time.Now().Add(blockDuration)
	return &blockedUntil, nil
}
