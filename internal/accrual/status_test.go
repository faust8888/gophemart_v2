package accrual

import (
	"testing"

	"github.com/faust8888/gophemart_v2/internal/model"
)

func TestLocalStatus(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{in: StatusRegistered, want: model.OrderStatusProcessing, wantOK: true},
		{in: StatusProcessing, want: model.OrderStatusProcessing, wantOK: true},
		{in: StatusInvalid, want: model.OrderStatusInvalid, wantOK: true},
		{in: StatusProcessed, want: model.OrderStatusProcessed, wantOK: true},
		{in: "UNKNOWN", wantOK: false},
		{in: "", wantOK: false},
	}
	for _, tc := range cases {
		got, ok := LocalStatus(tc.in)
		if ok != tc.wantOK || got != tc.want {
			t.Errorf("LocalStatus(%q) = %q, %v, want %q, %v", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}
