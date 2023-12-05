package client

import (
	"context"
	"encoding/json"
)

type (
	updateSAMLGroupMappingsRequest struct {
		Attributes SAMLGroupMappings `json:"attributes"`
	}

	SAMLGroupMappings struct {
		Mappings []SAMLGroupMapping `json:"mappings,omitempty"`
	}

	SAMLGroupMapping struct {
		SAMLAttributeKey   string            `json:"saml-attribute-key"`
		SAMLAttributeValue string            `json:"saml-attribute-value"`
		OrganizationRole   string            `json:"organization-role"`
		ProjectRoles       map[string]string `json:"project-roles"`
	}
)

func (c *Client) ListSAMLGroupMappings(ctx context.Context) (SAMLGroupMappings, error) {
	var resp genericAPIResponse[updateSAMLGroupMappingsRequest]

	err := c.CallAPI(ctx, "GET", "saml_group_mappings", nil, &resp)
	if err != nil {
		return SAMLGroupMappings{}, err
	}

	return resp.Data.Attributes, nil
}

func (c *Client) UpdateSAMLGroupMappings(ctx context.Context, mappings SAMLGroupMappings) error {
	bytes, err := json.Marshal(updateSAMLGroupMappingsRequest{Attributes: mappings})
	if err != nil {
		return err
	}

	return c.CallAPI(ctx, "POST", "saml_group_mappings", Envelope{Data: bytes}, nil)
}
