package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_InferredServiceRule_GetInferredServiceRulesReturnsRule(t *testing.T) {
	var server *httptest.Server

	testInferredServiceRuleResponseAttributes := InferredServiceRuleResponseAttributes{
		Name:        "test-rule",
		Description: "test inferred service rule description",
		AttributeFilters: []AttributeFilter{
			{
				Key:    "db.type",
				Values: []string{"sql", "redis"},
			},
		},
		GroupByKeys: []string{"db.type"},
	}

	server = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/public/v0.2/blars/projects/terraform-provider-tests/inferred_service_rules/testRuleId", request.URL.Path)

		response, err := json.Marshal(map[string]any{
			"type": "inferred_service_rule",
			"id":   "testRuleId",
			"data": map[string]any{
				"attributes": testInferredServiceRuleResponseAttributes,
			},
		})
		require.NoError(t, err)

		_, err = responseWriter.Write(response)
		require.NoError(t, err)
		responseWriter.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiClient := NewClient("api", "blars", server.URL)
	inferredServiceRule, err := apiClient.GetInferredServiceRule(
		context.Background(),
		"terraform-provider-tests",
		"testRuleId",
	)

	assert.NoError(t, err)
	assert.Equal(t, testInferredServiceRuleResponseAttributes, inferredServiceRule.Attributes)

}

func Test_getInferredServiceRuleUrl_returnsInferredServiceRuleUrl(t *testing.T) {
	require.Equal(
		t,
		"projects/test_project/inferred_service_rules",
		getInferredServiceRuleUrl("test_project"),
	)
	require.Equal(
		t,
		"projects/TestProject/inferred_service_rules",
		getInferredServiceRuleUrl("TestProject"),
	)
}

func Test_getInferredServiceRuleUrlWithId_returnsInferredServiceRuleIdUrl(t *testing.T) {
	require.Equal(
		t,
		"projects/test_project/inferred_service_rules/1234",
		getInferredServiceRuleUrlWithId("test_project", "1234"),
	)
	require.Equal(
		t,
		"projects/TestProject/inferred_service_rules/fLx72349023",
		getInferredServiceRuleUrlWithId("TestProject", "fLx72349023"),
	)
}
