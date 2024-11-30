package cdn

import (
	"context"
	"testing"
)

func init() {
	InitCdn()
}

func TestPushCdnCache(t *testing.T) {
	type args struct {
		ctx  context.Context
		path []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "t1", args: args{path: []string{"https://cdn-testing.finogeeks.club/api/v1/cloud/swagger/index.html"}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PushCdnCache(tt.args.ctx, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("PushCdnCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
