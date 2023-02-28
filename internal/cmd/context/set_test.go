package context

import (
	"testing"

	"github.com/mia-platform/miactl/internal/clioptions"
	"github.com/stretchr/testify/require"
)

func TestUpdateContextMap(t *testing.T) {
	// Test creating a new context
	opts := &clioptions.ContextOptions{APIBaseURL: "https://url", ProjectID: "project1", CompanyID: "company1"}
	newContext := map[string]string{"apibaseurl": "https://url", "projectid": "project1", "companyid": "company1"}
	expectedContexts := make(map[string]interface{})
	expectedContexts["context1"] = newContext
	actualContexts := updateContextMap(opts, "context1")
	require.Equal(t, expectedContexts, actualContexts)

	// Test updating the existing context
	opts = &clioptions.ContextOptions{APIBaseURL: "https://url2", ProjectID: "project2", CompanyID: "company2"}
	updatedContext := map[string]string{"apibaseurl": "https://url2", "projectid": "project2", "companyid": "company2"}
	expectedContexts["context1"] = updatedContext
	actualContexts = updateContextMap(opts, "context1")
	require.Equal(t, expectedContexts, actualContexts)
}
