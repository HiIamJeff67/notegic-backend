package enums

type AccessControlPermission string

const (
	AccessControlPermission_Read  AccessControlPermission = "Read"
	AccessControlPermission_Write AccessControlPermission = "Write"
	AccessControlPermission_Admin AccessControlPermission = "Admin"
	AccessControlPermission_Owner AccessControlPermission = "Owner"
)
