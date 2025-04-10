package inMemoryTokenBlacklist

import (
	"errors"
	"geo/db/tokenBlacklist"
	"testing"
	"time"
)

func TestBlacklist_Add(t *testing.T) {
	bl := NewBlacklist(time.Second * 1)
	type args struct {
		jti string
		exp time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "valid",
			args: args{
				jti: "123",
				exp: time.Now().UTC(),
			},
			wantErr: nil,
		},
		{
			name: "expired",
			args: args{
				jti: "123",
				exp: time.Now().UTC(),
			},
			wantErr: tokenBlacklist.Expired,
		},
		{
			name: "valid long live",
			args: args{
				jti: "456",
				exp: time.Now().Add(10 * time.Second).UTC(),
			},
			wantErr: nil,
		},
		{
			name: "add previously expired",
			args: args{
				jti: "123",
				exp: time.Now().Add(time.Second * 10).UTC(),
			},
			wantErr: nil,
		},
		{
			name: "add existing",
			args: args{
				jti: "123",
				exp: time.Now().Add(time.Minute).UTC(),
			},
			wantErr: tokenBlacklist.JTIAlreadyExists,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if i == 1 {
				time.Sleep(time.Second)
			}
			err := bl.Add(tt.args.jti, tt.args.exp)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBlacklist_clean(t *testing.T) {
	bl := NewBlacklist(time.Millisecond * 500)
	err := bl.Add("123", time.Now().UTC())
	if err != nil {
		t.Errorf("error adding token: %v", err)
	}
	tests := []struct {
		name string
	}{
		{
			name: "success",
		},
		{
			name: "not expired",
		},
		{
			name: "expired",
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if i == 0 {
				time.Sleep(time.Millisecond * 500)
			}
			bl.clean()
			if i == 0 {
				if len(bl.list) != 0 {
					t.Errorf("blacklist not clean")
				}
				err = bl.Add("123", time.Now().Add(time.Millisecond*100).UTC())
				if err != nil {
					t.Errorf("cannot add into blacklist: %v", err.Error())
				}
				time.Sleep(time.Millisecond * 100)
			} else if i == 1 {
				if len(bl.list) == 0 {
					t.Errorf("blacklist is clean")
				}
				time.Sleep(time.Millisecond * 500)
			} else if i == 2 {
				if len(bl.list) != 0 {
					t.Errorf("blacklist is not clean")
				}
			}
		})
	}
}

func TestBlacklist_IsInBlacklist(t *testing.T) {
	bl := NewBlacklist(time.Millisecond * 500)
	err := bl.Add("123", time.Now().UTC())
	if err != nil {
		t.Errorf("cannot add into blacklist: %v", err.Error())
	}
	type args struct {
		jti string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		blLen int
	}{
		{
			name: "success",
			args: args{
				jti: "123",
			},
			want:  true,
			blLen: 1,
		},
		{
			name: "empty after finding",
			args: args{
				jti: "123",
			},
			want:  true,
			blLen: 0,
		},
		{
			name: "empty initially",
			args: args{
				jti: "123",
			},
			want:  false,
			blLen: 0,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bl.Contains(tt.args.jti); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
			if tt.blLen != len(bl.list) {
				t.Errorf("len(bl.list) = %v, want %v", len(bl.list), tt.blLen)
			}
			if i == 0 {
				time.Sleep(time.Millisecond * 500)
			}
		})
	}
}
