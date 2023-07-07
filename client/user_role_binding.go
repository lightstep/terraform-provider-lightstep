package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type RoleBinding struct {
	RoleName    string   `json:"role-name"`
	ProjectName string   `json:"project-name"`
	Users       []string `json:"users"`
}

func (rb RoleBinding) ID() string {
	if rb.ProjectName == "" {
		return rb.RoleName
	}
	return fmt.Sprintf("%s/%s", rb.RoleName, rb.ProjectName)
}

type updateRoleBindingAPIResponse struct {
	Attributes RoleBinding `json:"attributes,omitempty"`
}

func (c *Client) ListRoleBinding(
	ctx context.Context,
	projectName string,
	roleName string,
) (RoleBinding, error) {
	var resp genericAPIResponse[updateRoleBindingAPIResponse]

	err := c.CallAPI(ctx, "GET", fmt.Sprintf("role-binding?role-name=%s&project=%s", url.QueryEscape(roleName), url.QueryEscape(projectName)), nil, &resp)
	if err != nil {
		return RoleBinding{}, err
	}

	return resp.Data.Attributes, nil
}

func (c *Client) UpdateRoleBinding(
	ctx context.Context,
	projectName string,
	roleName string,
	users ...string,
) (RoleBinding, error) {
	var resp Envelope
	var roleBinding updateRoleBindingAPIResponse

	bytes, err := json.Marshal(RoleBinding{
		ProjectName: projectName,
		RoleName:    roleName,
		Users:       users,
	})
	if err != nil {
		return roleBinding.Attributes, err
	}

	err = c.CallAPI(ctx, "POST", "role-binding", Envelope{Data: bytes}, &resp)
	if err != nil {
		return roleBinding.Attributes, err
	}

	err = json.Unmarshal(resp.Data, &roleBinding.Attributes)
	if err != nil {
		return roleBinding.Attributes, err
	}

	return roleBinding.Attributes, nil
}
