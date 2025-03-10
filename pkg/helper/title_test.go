package helper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/helper"
)

func TestTitle(t *testing.T) {
	assert.Equal(t, "Kyverno", helper.Title("kyverno"))
	assert.Equal(t, "Trivy Vulnerability", helper.Title("trivy vulnerability"))
}
