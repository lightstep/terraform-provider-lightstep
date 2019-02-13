package lightstep

type CreateUserAPIResponse struct {
	Data *CreateUserResponse `json:"data"`
}

type CreateUserResponse struct {
	Response
	LinksObj
	Attributes CreateUserResponseAttributes `json:"attributes,omitempty"`
}

type CreateUserResponseAttributes struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type CreateUserRequestAttributes struct {
	Username      string `json:"username"`
	Role          string `json:"role"`
	WithLoginLink bool   `json:"with-login-link"`
}

type CreateUserRequest struct {
	Attributes CreateUserRequestAttributes `json:"attributes"`
}

type CreateUserRequestBody struct {
	Data *CreateUserRequest `json:"data"`
}
