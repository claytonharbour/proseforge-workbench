package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotificationReadEndpointsUseGeneratedContract(t *testing.T) {
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/notifications":
			_, _ = w.Write([]byte(`{"notifications":[{"id":"notification-1","type":"contribution_ready","fromUserId":"actor-1","storyId":"story-1","actionUrl":"/story/story-1/room","isRead":false}],"unreadCount":1}`))
		case "/api/v1/notifications/unread/count":
			_, _ = w.Write([]byte(`{"count":1}`))
		default:
			t.Fatalf("unexpected request: %s", r.URL.RequestURI())
		}
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	list, err := client.ListNotifications(context.Background(), 25, 50)
	if err != nil {
		t.Fatal(err)
	}
	if list.UnreadCount == nil || *list.UnreadCount != 1 || list.Notifications == nil || len(*list.Notifications) != 1 {
		t.Fatalf("list = %#v", list)
	}
	notification := (*list.Notifications)[0]
	if notification.FromUserId == nil || *notification.FromUserId != "actor-1" ||
		notification.StoryId == nil || *notification.StoryId != "story-1" ||
		notification.ActionUrl == nil || *notification.ActionUrl != "/story/story-1/room" {
		t.Fatalf("notification attribution/link = %#v", notification)
	}
	count, err := client.GetNotificationUnreadCount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count.Count == nil || *count.Count != 1 {
		t.Fatalf("count = %#v", count)
	}

	want := []string{
		"GET /api/v1/notifications?limit=25&offset=50",
		"GET /api/v1/notifications/unread/count",
	}
	if len(requests) != len(want) {
		t.Fatalf("requests = %v", requests)
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Errorf("request[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestListNotificationsDecodesEmptyPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"notifications":[],"unreadCount":0}`))
	}))
	defer srv.Close()

	client, err := New(srv.URL, "test-token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ListNotifications(context.Background(), 25, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Notifications == nil || len(*result.Notifications) != 0 || result.UnreadCount == nil || *result.UnreadCount != 0 {
		t.Fatalf("result = %#v", result)
	}
}
