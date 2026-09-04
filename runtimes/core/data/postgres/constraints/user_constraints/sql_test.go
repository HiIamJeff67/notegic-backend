package userconstraints

import (
	"strings"
	"testing"
)

func TestUserForeignKeysCascadeDeletes(t *testing.T) {
	for _, fragment := range []string{
		"FOREIGN KEY (user_id) REFERENCES \"UserTable\" (id)",
		"FOREIGN KEY (owner_id) REFERENCES \"UserTable\" (id)",
		"FOREIGN KEY (author_id) REFERENCES \"UserTable\" (id)",
		"FOREIGN KEY (badge_id) REFERENCES \"BadgeTable\" (id)",
		"FOREIGN KEY (billing_plan_id) REFERENCES \"BillingPlanTable\" (id)",
		"FOREIGN KEY (plan) REFERENCES \"PlanLimitationTable\" (key)",
		"ON DELETE CASCADE",
	} {
		if !strings.Contains(UserForeignKeysSQL, fragment) {
			t.Fatalf("UserForeignKeysSQL must contain %q", fragment)
		}
	}
}
