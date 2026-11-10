package domain

import (
	"testing"
)

func TestDeliveryIdentity(t *testing.T) {
	raw := []byte("Message-ID: <reused@example.com>\r\n\r\noriginal")
	first, err := NewDeliveryKey("reader@EXAMPLE.com", raw)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := NewDeliveryKey("reader@example.com", raw)
	if err != nil {
		t.Fatal(err)
	}
	if first != retry {
		t.Fatal("domain casing changed retry identity")
	}
	if first != "mime-v1:a1a7b951cdf521275480e08041cb18469c2ed125825c85fd7e0d9faf43b05a9e" {
		t.Fatalf("identity algorithm changed: %s", first)
	}
	for _, tc := range []struct {
		recipient string
		raw       []byte
	}{
		{"other@example.com", raw},
		{"Reader@example.com", raw},
		{"reader+tag@example.com", raw},
		{"reader@example.com", []byte("Message-ID: <reused@example.com>\r\n\r\nchanged")},
	} {
		key, err := NewDeliveryKey(tc.recipient, tc.raw)
		if err != nil {
			t.Fatal(err)
		}
		if key == first {
			t.Fatal("different content or mailbox merged")
		}
	}
}

func TestDeliveryIdentityRejectsInvalidInput(t *testing.T) {
	for _, recipient := range []string{"", "not-a-mailbox", "Name <reader@example.com>", "reader@example.com\r\nX: injected", " reader@example.com"} {
		if _, err := NewDeliveryKey(recipient, []byte("message")); err == nil {
			t.Errorf("accepted invalid recipient %q", recipient)
		}
	}
	if _, err := NewDeliveryKey("reader@example.com", nil); err == nil {
		t.Fatal("accepted missing MIME")
	}
}
