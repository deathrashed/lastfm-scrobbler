package sessionstore

import (
	"testing"
	"time"
)

func TestHistoryAndPendingRoundTrip(t *testing.T) {
	store := New(t.TempDir())
	record := Record{
		ID: "session-1", Mode: "manual", StartedAt: time.Unix(1700000000, 0),
		Status: "pending", Completed: 1,
		Queue: []Track{{Artist: "Slayer", Album: "Hell Awaits", Title: "Hell Awaits"}},
	}
	if err := store.SavePending(record); err != nil {
		t.Fatal(err)
	}
	pending, err := store.LoadPending()
	if err != nil || pending == nil || pending.Completed != 1 {
		t.Fatalf("pending = %#v, err = %v", pending, err)
	}
	if err := store.Append(record); err != nil {
		t.Fatal(err)
	}
	history, err := store.LoadHistory()
	if err != nil || len(history) != 1 || history[0].ID != record.ID {
		t.Fatalf("history = %#v, err = %v", history, err)
	}
	if err := store.ClearPending(); err != nil {
		t.Fatal(err)
	}
	pending, err = store.LoadPending()
	if err != nil || pending != nil {
		t.Fatalf("pending after clear = %#v, err = %v", pending, err)
	}
}
