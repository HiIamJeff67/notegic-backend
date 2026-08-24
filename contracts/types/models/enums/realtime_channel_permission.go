package enums

type ChannelPermission string

const (
	ChannelPermission_Read  ChannelPermission = "read"
	ChannelPermission_Write ChannelPermission = "write"
)

func (p ChannelPermission) AllowedAccessControlPermissions() []AccessControlPermission {
	switch p {
	case ChannelPermission_Read:
		return []AccessControlPermission{
			AccessControlPermission_Owner,
			AccessControlPermission_Admin,
			AccessControlPermission_Write,
			AccessControlPermission_Read,
		}
	case ChannelPermission_Write:
		return []AccessControlPermission{
			AccessControlPermission_Owner,
			AccessControlPermission_Admin,
			AccessControlPermission_Write,
		}
	default:
		return nil
	}
}
