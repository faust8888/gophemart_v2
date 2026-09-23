package luhn

import "testing"

func TestValid(t *testing.T) {
	cases := []struct {
		number string
		want   bool
	}{
		{number: "12345678903", want: true},
		{number: "79927398713", want: true},
		{number: "0", want: true},
		{number: "12345678900", want: false},
		{number: "123", want: false},
		{number: "", want: false},
		{number: "abcdef", want: false},
		{number: "1234567890a", want: false},
		{number: " 12345678903", want: false},
	}

	for _, tc := range cases {
		if got := Valid(tc.number); got != tc.want {
			t.Errorf("Valid(%q) = %v, want %v", tc.number, got, tc.want)
		}
	}
}
