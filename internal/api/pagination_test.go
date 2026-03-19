package api

import (
	"net/url"
	"testing"
)

func TestApplyPaging_AllFields(t *testing.T) {
	vals := url.Values{}
	p := &PageParams{Limit: 25, Before: "cur_before", After: "cur_after"}

	ApplyPaging(vals, p)

	if got := vals.Get("limit"); got != "25" {
		t.Errorf("limit = %q, want %q", got, "25")
	}
	if got := vals.Get("before"); got != "cur_before" {
		t.Errorf("before = %q, want %q", got, "cur_before")
	}
	if got := vals.Get("after"); got != "cur_after" {
		t.Errorf("after = %q, want %q", got, "cur_after")
	}
}

func TestApplyPaging_ZeroLimitOmitted(t *testing.T) {
	vals := url.Values{}
	p := &PageParams{Limit: 0, Before: "b", After: "a"}

	ApplyPaging(vals, p)

	if vals.Has("limit") {
		t.Error("limit should be omitted when zero")
	}
	if got := vals.Get("before"); got != "b" {
		t.Errorf("before = %q, want %q", got, "b")
	}
	if got := vals.Get("after"); got != "a" {
		t.Errorf("after = %q, want %q", got, "a")
	}
}

func TestApplyPaging_EmptyBeforeAfterOmitted(t *testing.T) {
	vals := url.Values{}
	p := &PageParams{Limit: 10}

	ApplyPaging(vals, p)

	if got := vals.Get("limit"); got != "10" {
		t.Errorf("limit = %q, want %q", got, "10")
	}
	if vals.Has("before") {
		t.Error("before should be omitted when empty")
	}
	if vals.Has("after") {
		t.Error("after should be omitted when empty")
	}
}

func TestApplyPaging_NilIsNoOp(t *testing.T) {
	vals := url.Values{}

	ApplyPaging(vals, nil)

	if len(vals) != 0 {
		t.Errorf("expected no params, got %v", vals)
	}
}

func TestParsePaging_FullObject(t *testing.T) {
	raw := map[string]any{
		"paging": map[string]any{
			"cursors": map[string]any{
				"before": "abc",
				"after":  "xyz",
			},
			"next":     "https://graph.threads.net/next",
			"previous": "https://graph.threads.net/prev",
		},
	}

	info := ParsePaging(raw)

	if info.Before != "abc" {
		t.Errorf("Before = %q, want %q", info.Before, "abc")
	}
	if info.After != "xyz" {
		t.Errorf("After = %q, want %q", info.After, "xyz")
	}
	if info.Next != "https://graph.threads.net/next" {
		t.Errorf("Next = %q, want %q", info.Next, "https://graph.threads.net/next")
	}
	if info.Previous != "https://graph.threads.net/prev" {
		t.Errorf("Previous = %q, want %q", info.Previous, "https://graph.threads.net/prev")
	}
}

func TestParsePaging_MissingCursors(t *testing.T) {
	raw := map[string]any{
		"paging": map[string]any{
			"next": "https://graph.threads.net/next",
		},
	}

	info := ParsePaging(raw)

	if info.Before != "" {
		t.Errorf("Before = %q, want empty", info.Before)
	}
	if info.After != "" {
		t.Errorf("After = %q, want empty", info.After)
	}
	if info.Next != "https://graph.threads.net/next" {
		t.Errorf("Next = %q, want %q", info.Next, "https://graph.threads.net/next")
	}
}

func TestParsePaging_NoPagingKey(t *testing.T) {
	raw := map[string]any{
		"data": []any{"something"},
	}

	info := ParsePaging(raw)

	if info.Before != "" || info.After != "" || info.Next != "" || info.Previous != "" {
		t.Errorf("expected empty PageInfo, got %+v", info)
	}
}

func TestPageInfo_HasNext_WithNext(t *testing.T) {
	p := &PageInfo{Next: "https://graph.threads.net/next"}

	if !p.HasNext() {
		t.Error("HasNext() = false, want true when Next is set")
	}
}

func TestPageInfo_HasNext_WithOnlyAfter(t *testing.T) {
	p := &PageInfo{After: "cursor123"}

	if !p.HasNext() {
		t.Error("HasNext() = false, want true when After is set")
	}
}

func TestPageInfo_HasNext_Neither(t *testing.T) {
	p := &PageInfo{}

	if p.HasNext() {
		t.Error("HasNext() = true, want false when neither Next nor After is set")
	}
}
