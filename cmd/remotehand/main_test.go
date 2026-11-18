package main

import "testing"

func TestPointerSensitivity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		want    float64
		wantErr bool
	}{
		{value: "", want: 1},
		{value: "0.5", want: 0.5},
		{value: "3", want: 3},
		{value: "0", wantErr: true},
		{value: "fast", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			t.Parallel()
			got, err := pointerSensitivity(test.value)
			if (err != nil) != test.wantErr {
				t.Fatalf("pointerSensitivity(%q) error = %v", test.value, err)
			}
			if got != test.want {
				t.Fatalf("pointerSensitivity(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}
