package snap

import (
	"context"
	"errors"
	"io"

	"github.com/Korrnals/gotr/internal/models/data"
)

// errAttachAPINotImplemented is returned by mock attachment methods that
// are not exercised in the existing rollback test suites.
var errAttachAPINotImplemented = errors.New("attachment API not implemented in this mock")

func (m *mockCasesAPI) DownloadAttachment(_ context.Context, _ int64) (io.ReadCloser, error) {
	return nil, errAttachAPINotImplemented
}

func (m *mockCasesAPI) AddAttachmentToCase(_ context.Context, _ int64, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}

func (m *mockCasesAPI) AddAttachmentToPlan(_ context.Context, _ int64, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}

func (m *mockCasesAPI) AddAttachmentToPlanEntry(_ context.Context, _ int64, _, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}

func (m *mockCasesAPI) AddAttachmentToResult(_ context.Context, _ int64, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}

func (m *mockCasesAPI) AddAttachmentToRun(_ context.Context, _ int64, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}

func (a *statefulCasesAPI) DownloadAttachment(_ context.Context, _ int64) (io.ReadCloser, error) {
	return nil, errAttachAPINotImplemented
}

func (a *statefulCasesAPI) AddAttachmentToCase(_ context.Context, _ int64, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}

func (a *statefulCasesAPI) AddAttachmentToPlan(_ context.Context, _ int64, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}

func (a *statefulCasesAPI) AddAttachmentToPlanEntry(_ context.Context, _ int64, _, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}

func (a *statefulCasesAPI) AddAttachmentToResult(_ context.Context, _ int64, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}

func (a *statefulCasesAPI) AddAttachmentToRun(_ context.Context, _ int64, _ string) (*data.AttachmentResponse, error) {
	return nil, errAttachAPINotImplemented
}
