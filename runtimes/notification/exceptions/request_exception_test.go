package exceptions

import "testing"

func TestRequestRecipientRequired(t *testing.T) {
	exception := NewRequestException("Notification").RecipientRequired()
	if exception.Reason != "RecipientUserPublicIdRequired" {
		t.Fatalf("unexpected request exception: %#v", exception)
	}
}
