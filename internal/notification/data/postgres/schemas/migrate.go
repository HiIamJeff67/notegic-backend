package schemas

import cmodels "github.com/HiIamJeff67/notegic-backend/contracts/types/models"

var MigratingTables = []any{
	&Notification{},
	&cmodels.InboxEvent{},
	&cmodels.OutboxEvent{},
}
