package commands

import (
	"testing"
	"time"
)

func TestGetTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		timeArg string
		want    int64
		wantErr bool
	}{
		{
			name:    "Parsing hour",
			timeArg: "1h",
			want:    time.Now().Add(time.Duration(1) * time.Hour).UnixMilli(),
			wantErr: false,
		},
		{
			name:    "Parsing minute",
			timeArg: "40m",
			want:    time.Now().Add(time.Duration(40) * time.Minute).UnixMilli(),
			wantErr: false,
		},
		{
			name:    "Parsing second",
			timeArg: "50s",
			want:    time.Now().Add(time.Duration(50) * time.Second).UnixMilli(),
			wantErr: false,
		},
		{
			name:    "Exact time",
			timeArg: "22:40",
			want: func() int64 {
				target := time.Date(now.Year(), now.Month(), now.Day(), 22, 40, 0, 0, now.Location())
				if target.Before(now) {
					target = target.Add(time.Duration(24) * time.Hour)
				}

				return target.UnixMilli()
			}(),
			wantErr: false,
		},
		{
			name:    "Wrong time",
			timeArg: "jas",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTime(tt.timeArg)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTime() = %v, want %v", got, tt.want)
			}
		})
	}


}
