// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: accounts.proto

package accountspb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/mwitkow/go-proto-validators"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *OauthCodeExchangeData) Validate() error {
	if !(len(this.User) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("User", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.User))
	}
	if !(len(this.ClientId) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("ClientId", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.ClientId))
	}
	if !(len(this.Scope) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Scope", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Scope))
	}
	if !(len(this.State) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("State", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.State))
	}
	return nil
}
func (this *UserEmailValidationData) Validate() error {
	if !(len(this.User) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("User", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.User))
	}
	if !(len(this.Email) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Email", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Email))
	}
	return nil
}
func (this *ResetPasswordEmailData) Validate() error {
	if !(len(this.User) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("User", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.User))
	}
	return nil
}
func (this *Authorization) Validate() error {
	if !(len(this.User) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("User", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.User))
	}
	if !(len(this.ApplicationId) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("ApplicationId", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.ApplicationId))
	}
	for _, item := range this.Scope {
		if !(len(item) > 0) {
			return github_com_mwitkow_go_proto_validators.FieldError("Scope", fmt.Errorf(`value '%v' must have a length greater than '0'`, item))
		}
	}
	if !(this.Duration > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Duration", fmt.Errorf(`value '%v' must be greater than '0'`, this.Duration))
	}
	return nil
}
func (this *Session) Validate() error {
	if !(len(this.Id) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Id", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Id))
	}
	if !(len(this.UserAgent) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("UserAgent", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.UserAgent))
	}
	if !(len(this.Jwt) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Jwt", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Jwt))
	}
	return nil
}
func (this *AccountValidationData) Validate() error {
	if !(len(this.User) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("User", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.User))
	}
	if !(len(this.Email) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Email", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Email))
	}
	return nil
}
func (this *AccountInfo) Validate() error {
	return nil
}
func (this *FindAccountRequest) Validate() error {
	return nil
}
func (this *FindAccountResponse) Validate() error {
	return nil
}
func (this *CreateAccountRequest) Validate() error {
	if !(len(this.Username) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Username", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Username))
	}
	if !(len(this.Email) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Email", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Email))
	}
	if !(len(this.Password) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Password", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Password))
	}
	return nil
}
func (this *CreateAccountResponse) Validate() error {
	return nil
}
func (this *ValidateAccountRequest) Validate() error {
	return nil
}
func (this *ValidateAccountResponse) Validate() error {
	return nil
}
func (this *AccountInfoRequest) Validate() error {
	return nil
}
func (this *AccountInfoResponse) Validate() error {
	if this.Info != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Info); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Info", err)
		}
	}
	return nil
}
func (this *SelectAccountRequest) Validate() error {
	return nil
}
func (this *SelectAccountResponse) Validate() error {
	return nil
}
func (this *ConfirmAccountRequest) Validate() error {
	return nil
}
func (this *ConfirmAccountResponse) Validate() error {
	return nil
}
func (this *RequestPasswordResetRequest) Validate() error {
	return nil
}
func (this *RequestPasswordResetResponse) Validate() error {
	return nil
}
func (this *UpdatePasswordRequest) Validate() error {
	return nil
}
func (this *UpdatePasswordResponse) Validate() error {
	return nil
}
func (this *LoginRequest) Validate() error {
	if !(len(this.User) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("User", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.User))
	}
	if !(len(this.Password) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Password", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Password))
	}
	return nil
}
func (this *LoginResponse) Validate() error {
	return nil
}
func (this *LogoutRequest) Validate() error {
	return nil
}
func (this *LogoutResponse) Validate() error {
	return nil
}
func (this *ListSessionsRequest) Validate() error {
	return nil
}
func (this *ListSessionsResponse) Validate() error {
	for _, item := range this.Sessions {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Sessions", err)
			}
		}
	}
	return nil
}
func (this *CloseSessionRequest) Validate() error {
	if !(len(this.SessionId) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("SessionId", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.SessionId))
	}
	return nil
}
func (this *CloseSessionResponse) Validate() error {
	return nil
}
func (this *AuthorizeRequest) Validate() error {
	if !(len(this.ResponseType) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("ResponseType", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.ResponseType))
	}
	if !(len(this.ClientId) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("ClientId", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.ClientId))
	}
	if !(len(this.Scope) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Scope", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Scope))
	}
	if !(len(this.State) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("State", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.State))
	}
	return nil
}
func (this *AuthorizeResponse) Validate() error {
	return nil
}
func (this *GetTokenRequest) Validate() error {
	if !(len(this.GrantType) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("GrantType", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.GrantType))
	}
	if !(len(this.Code) > 0) {
		return github_com_mwitkow_go_proto_validators.FieldError("Code", fmt.Errorf(`value '%v' must have a length greater than '0'`, this.Code))
	}
	return nil
}
func (this *GetTokenResponse) Validate() error {
	return nil
}
