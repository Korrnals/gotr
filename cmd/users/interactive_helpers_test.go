package users

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/interactive"
	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
)

func TestResolveUserIDInteractive_Table(t *testing.T) {
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
				GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) {
					return data.GetUsersResponse{{ID: 10, Name: "A", Email: "a@test"}, {ID: 20, Name: "B", Email: "b@test"}}, nil
				},
			},
			wantID: 20,
		},
		{
			name: "get users error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) { return nil, assert.AnError }},
			wantErrPart: "failed to get users",
		},
		{
			name: "no users",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) { return data.GetUsersResponse{}, nil }},
			wantErrPart: "no users found",
		},
		{
			name: "select error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			cli: &client.MockClient{GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) { return data.GetUsersResponse{{ID: 1, Name: "N", Email: "e"}}, nil }},
			wantErrPart: "failed to select user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := resolveUserIDInteractive(tt.ctx, tt.cli)
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

func TestResolveEmailInteractive_Table(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		cli         client.ClientInterface
		wantEmail   string
		wantErrPart string
	}{
		{
			name: "success",
			ctx: interactive.WithPrompter(context.Background(),
				interactive.NewMockPrompter().WithSelectResponses(interactive.SelectResponse{Index: 0})),
			cli: &client.MockClient{
				GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) {
					return data.GetUsersResponse{{ID: 10, Name: "A", Email: "a@test"}, {ID: 20, Name: "B", Email: "b@test"}}, nil
				},
			},
			wantEmail: "a@test",
		},
		{
			name: "get users error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) { return nil, assert.AnError }},
			wantErrPart: "failed to get users",
		},
		{
			name: "no users",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			cli: &client.MockClient{GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) { return data.GetUsersResponse{}, nil }},
			wantErrPart: "no users found",
		},
		{
			name: "select error",
			ctx:  interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			cli: &client.MockClient{GetUsersFunc: func(ctx context.Context) (data.GetUsersResponse, error) { return data.GetUsersResponse{{ID: 1, Name: "N", Email: "e"}}, nil }},
			wantErrPart: "failed to select user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := resolveEmailInteractive(tt.ctx, tt.cli)
			if tt.wantErrPart != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrPart)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantEmail, email)
		})
	}
}

func TestRequireInteractiveUserArg_Table(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		usage       string
		wantErrPart string
	}{
		{
			name:        "no prompter in context",
			ctx:         context.Background(),
			usage:       "gotr users get [user_id]",
			wantErrPart: "required argument is missing in non-interactive mode",
		},
		{
			name:        "explicit non-interactive prompter",
			ctx:         interactive.WithPrompter(context.Background(), interactive.NewNonInteractivePrompter()),
			usage:       "gotr users get [user_id]",
			wantErrPart: "required argument is missing in non-interactive mode",
		},
		{
			name:  "interactive prompter",
			ctx:   interactive.WithPrompter(context.Background(), interactive.NewMockPrompter()),
			usage: "gotr users get [user_id]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireInteractiveUserArg(tt.ctx, tt.usage)
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
