package draftkings

import "testing"

func TestRetryableStatus(t *testing.T) {
	tests := []struct {
		status int
		want   bool
	}{
		{status: 200, want: false},
		{status: 403, want: false},
		{status: 429, want: true},
		{status: 500, want: true},
		{status: 503, want: true},
	}

	for _, test := range tests {
		if got := retryableStatus(test.status); got != test.want {
			t.Errorf("retryableStatus(%d) = %v, want %v", test.status, got, test.want)
		}
	}
}
