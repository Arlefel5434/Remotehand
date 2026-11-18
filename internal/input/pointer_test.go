package input

import "testing"

func TestMockPointer(t *testing.T) {
	t.Parallel()
	adapter, err := New("mock")
	if err != nil {
		t.Fatalf("New(mock): %v", err)
	}
	if err := adapter.Move(10, -5); err != nil {
		t.Fatalf("Move(): %v", err)
	}
	if err := adapter.Click(ButtonLeft); err != nil {
		t.Fatalf("Click(): %v", err)
	}
	if err := adapter.SendText("Привет 👋"); err != nil {
		t.Fatalf("SendText(): %v", err)
	}
	if err := adapter.PressKey(KeyEnter); err != nil {
		t.Fatalf("PressKey(): %v", err)
	}
}
