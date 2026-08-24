package enums

type UserStatus string

const (
	UserStatus_Online       UserStatus = "Online"
	UserStatus_AFK          UserStatus = "AFK"
	UserStatus_DoNotDisturb UserStatus = "DoNotDisturb"
	UserStatus_Offline      UserStatus = "Offline"
)
