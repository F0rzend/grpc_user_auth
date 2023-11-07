package proto

import (
	"net/mail"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	_ Validatable = (*CreateUserRequest)(nil)
	_ Validatable = (*GetUserRequest)(nil)
	_ Validatable = (*UpdateUserRequest)(nil)
	_ Validatable = (*DeleteUserRequest)(nil)
)

type Validatable interface {
	Validate() error
}

func (m *CreateUserRequest) Validate() error {
	if _, err := mail.ParseAddress(m.Email); err != nil {
		return status.Errorf(codes.InvalidArgument, "E-mail is invalid")
	}

	return nil
}

func (m *GetUserRequest) Validate() error {
	if _, err := uuid.Parse(m.Id); err != nil {
		return status.Errorf(codes.InvalidArgument, "ID is invalid")
	}

	return nil
}

func (m *UpdateUserRequest) Validate() error {
	if _, err := uuid.Parse(m.Id); err != nil {
		return status.Errorf(codes.InvalidArgument, "ID is invalid")
	}

	if _, err := mail.ParseAddress(m.Email); err != nil {
		return status.Errorf(codes.InvalidArgument, "E-mail is invalid")
	}

	return nil
}

func (m *DeleteUserRequest) Validate() error {
	if _, err := uuid.Parse(m.Id); err != nil {
		return status.Errorf(codes.InvalidArgument, "ID is invalid")
	}

	return nil
}
