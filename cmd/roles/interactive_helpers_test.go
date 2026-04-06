package roles

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResolveRoleIDInteractive_Table(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantID      int64
		wantErrPart string
	}{
		{
			name: "success",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 1})),
			cli: &client.MockClient{
				GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) {
					return data.GetRolesResponse{{ID: 11, Name: "R1"}, {ID: 22, Name: "R2"}}, nil
				},
			},
			wantID: 22,
		},
		{
			name: "get roles error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) { return nil, assert.AnError }},
			wantErrPart: "failed to get roles list",
		},
		{
			name: "no roles",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) { return data.GetRolesResponse{}, nil }},
			wantErrPart: "no roles found",
		},
		{
			name: "select error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			cli: &client.MockClient{GetRolesFunc: func(ctx context.Context) (data.GetRolesResponse, error) { return data.GetRolesResponse{{ID: 9, Name: "R"}}, nil }},
			wantErrPart: "failed to select role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveRoleIDInteractive(tt.ctx, tt.cli)
			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

func TestRequireInteractiveRoleArg_Table(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		usage       string
		wantErrPart string
	}{
		{
			name:        "no prompter in context",
			ctx:         context.Background(),
			usage:       "gotr roles get [role_id]",
			wantErrPart: "required argument is missing in non-interactive mode",
		},
		{
			name:        "explicit non-interactive prompter",
			ctx:         interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			usage:       "gotr roles get [role_id]",
			wantErrPart: "required argument is missing in non-interactive mode",
		},
		{
			name:  "interactive prompter",
			ctx:   interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			usage: "gotr roles get [role_id]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireInteractiveRoleArg(tt.ctx, tt.usage)
			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				assert.Contains(t, err.Error(), tt.usage)
				return
			}
			assert.NoError(t, err)
		})
	}
}
