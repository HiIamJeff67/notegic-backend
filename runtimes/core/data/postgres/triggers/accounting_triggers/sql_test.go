package accountingtriggersql

import (
	"strings"
	"testing"
)

func TestBlockAccountingTriggerFunctionsUseSecurityDefiner(t *testing.T) {
	for name, sql := range map[string]string{
		AccountingInsertedBlockTriggerFunctionName: AccountingInsertedBlockTriggerSQL,
		AccountingDeletedBlockTriggerFunctionName:  AccountingDeletedBlockTriggerSQL,
	} {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(
				sql,
				"$$ LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog, public;",
			) {
				t.Fatalf("%s must use SECURITY DEFINER with a fixed search_path", name)
			}
		})
	}
}
