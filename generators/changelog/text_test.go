package main

import "testing"

func TestRemoveEnum(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "removes enum values",
			text: "Enum: `io_uring`, `sync`, `worker`. EXPERIMENTAL: Controls the maximum number of I/O operations that one process can execute simultaneously. Version 18 and up only. Changing this parameter causes a service restart. Default: `worker`.",
			want: "EXPERIMENTAL: Controls the maximum number of I/O operations that one process can execute simultaneously. Version 18 and up only. Changing this parameter causes a service restart. Default: `worker`.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeEnum(tt.text); got != tt.want {
				t.Errorf("removeEnum() = %q, want %q", got, tt.want)
			}
		})
	}
}
