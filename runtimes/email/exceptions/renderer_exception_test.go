package exceptions

import "testing"

func TestRendererInvalidTemplate(t *testing.T) {
	exception := NewRendererException("Email").InvalidTemplate()
	if exception.Reason != "InvalidTemplate" || exception.Domain != "Email" {
		t.Fatalf("unexpected renderer exception: %#v", exception)
	}
}
