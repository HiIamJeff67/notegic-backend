package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type User struct {
	Id              uuid.UUID         `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	PublicId        uuid.UUID         `json:"publicId" gorm:"column:public_id; type:uuid; unique; not null; default:gen_random_uuid();"` // generated at BeforeCreate() trigger below
	Name            string            `json:"name" gorm:"column:name; unique; not null; size:32;"`                                       // validate:"required,min=6,max=16,alphaandnum"
	DisplayName     string            `json:"displayName" gorm:"column:display_name; not null; size:32;"`                                // validate:"required,min=6,max=32,alphaandnum"
	Email           string            `json:"email" gorm:"column:email; unique; not null;"`                                              // validate:"required,email"
	Password        string            `json:"password" gorm:"column:password; not null; size:1024;"`                                     // validate:"required,min=8,max=1024"      // since we store the hashed password which is quite long
	RefreshToken    string            `json:"refreshToken" gorm:"column:refresh_token; not null;"`                                       // validate:"omitnil"
	LoginCount      int32             `json:"loginCount" gorm:"column:login_count; type:integer; not null; default:0;"`
	BlockLoginUntil time.Time         `json:"blockLoginUntil" gorm:"column:block_login_until; type:timestamptz; not null;"`
	UserAgent       string            `json:"userAgent" gorm:"column:user_agent; not null;"`                          // validate:"required,isuseragent"
	Role            cenums.UserRole   `json:"role" gorm:"column:role; type:\"UserRole\"; not null; default:'Guest';"` // validate:"omitnil,isrole"
	Plan            cenums.UserPlan   `json:"plan" gorm:"column:plan; type:\"UserPlan\"; not null; default:'Free';"`  // validate:"omitnil,isplan" // only update the field of plan while upgrade logic happened in user account services
	PrevStatus      cenums.UserStatus `json:"prevStatus" gorm:"column:prev_status; type:\"UserStatus\"; not null; default:'Online';"`
	Status          cenums.UserStatus `json:"status" gorm:"column:status; type:\"UserStatus\"; not null; default:'Online';"` // validate:"omitnil,isstatus"
	UpdatedAt       time.Time         `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt       time.Time         `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	UserInfo            UserInfo              `json:"userInfo" gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	UserAccount         UserAccount           `json:"userAccount" gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	UserSetting         UserSetting           `json:"userSetting" gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	APIKeys             []APIKey              `json:"apiKeys" gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Themes              []Theme               `json:"themes" gorm:"foreignKey:AuthorId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	UsersToBadges       []UsersToBadges       `json:"usersToBadges" gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	UsersToShelves      []UsersToShelves      `json:"usersToShelves" gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	UsersToStations     []UsersToStations     `json:"usersToStations" gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	RoutineTags         []RoutineTag          `json:"routineTags" gorm:"foreignKey:OwnerId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Stations            []Station             `json:"stations" gorm:"foreignKey:OwnerId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	PlanLimitation      PlanLimitation        `json:"planLimitation" gorm:"foreignKey:Plan; references:Key; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	UsersToBillingPlans []UsersToBillingPlans `json:"usersToBillingPlans" gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// User Table Name
func (User) TableName() string {
	return postgres.TableName_UserTable.String()
}

// User Table Relations
type UserRelation postgres.RelationName

const (
	UserRelation_UserInfo            UserRelation = "UserInfo"
	UserRelation_UserAccount         UserRelation = "UserAccount"
	UserRelation_UserSetting         UserRelation = "UserSetting"
	UserRelation_APIKeys             UserRelation = "APIKeys"
	UserRelation_Themes              UserRelation = "Themes"
	UserRelation_UsersToBadges       UserRelation = "UsersToBadges"
	UserRelation_UsersToShelves      UserRelation = "UsersToShelves"
	UserRelation_UsersToStations     UserRelation = "UsersToStations"
	UserRelation_RoutineTags         UserRelation = "RoutineTags"
	UserRelation_Stations            UserRelation = "Stations"
	UserRelation_PlanLimitation      UserRelation = "PlanLimitation"
	UserRelation_UsersToBillingPlans UserRelation = "UsersToBillingPlans"
)

/* ============================== Relative Type Conversions ============================== */

func (u *User) ToPublicUser() *cgqlmodels.PublicUser {
	return &cgqlmodels.PublicUser{
		PublicID:    u.PublicId,
		Name:        u.Name,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		Plan:        u.Plan,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt,
	}
}

/* ============================== Trigger Hook ============================== */

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.BlockLoginUntil.IsZero() {
		// just to make the new user can login
		u.BlockLoginUntil = time.Now().Add(-10 * time.Minute)
	}
	if u.PublicId == uuid.Nil {
		u.PublicId = uuid.New()
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("Status") {
		u.PrevStatus = u.Status
	}
	return nil
}
