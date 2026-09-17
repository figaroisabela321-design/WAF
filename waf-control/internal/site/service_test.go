package site

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateRequest
		wantErr string
	}{
		{
			name: "valid http",
			req: CreateRequest{
				Name: "portal", Domain: "portal.gov.cn",
				UpstreamHost: "10.0.0.1", UpstreamPort: 80, Protocol: "http",
			},
		},
		{
			name: "valid https protect",
			req: CreateRequest{
				Name: "api", Domain: "api.gov.cn",
				UpstreamPort: 443, Protocol: "https", ProtectionMode: "protect",
			},
		},
		{
			name:    "missing name",
			req:     CreateRequest{Domain: "a.com", UpstreamPort: 80, Protocol: "http"},
			wantErr: "name is required",
		},
		{
			name:    "missing domain",
			req:     CreateRequest{Name: "x", UpstreamPort: 80, Protocol: "http"},
			wantErr: "domain is required",
		},
		{
			name:    "port zero",
			req:     CreateRequest{Name: "x", Domain: "a.com", UpstreamPort: 0, Protocol: "http"},
			wantErr: "upstream_port must be between 1 and 65535",
		},
		{
			name:    "port too high",
			req:     CreateRequest{Name: "x", Domain: "a.com", UpstreamPort: 70000, Protocol: "http"},
			wantErr: "upstream_port must be between 1 and 65535",
		},
		{
			name:    "bad protocol",
			req:     CreateRequest{Name: "x", Domain: "a.com", UpstreamPort: 80, Protocol: "ftp"},
			wantErr: "protocol must be http or https",
		},
		{
			name: "bad protection mode",
			req: CreateRequest{
				Name: "x", Domain: "a.com", UpstreamPort: 80, Protocol: "http",
				ProtectionMode: "block",
			},
			wantErr: "protection_mode must be observe or protect",
		},
		{
			name: "bad status",
			req: CreateRequest{
				Name: "x", Domain: "a.com", UpstreamPort: 80, Protocol: "http",
				Status: "active",
			},
			wantErr: "status must be enabled or disabled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreate(&tt.req)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
