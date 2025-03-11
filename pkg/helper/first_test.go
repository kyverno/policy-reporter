package helper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/helper"
)

type item struct {
	val int
}

func TestFirst(t *testing.T) {
	t.Run("return nil for empty list", func(t *testing.T) {
		assert.Nil(t, helper.First([]*item{}))
	})

	t.Run("return first item", func(t *testing.T) {
		assert.Equal(t, 0, helper.First([]*item{{val: 0}, {val: 1}}).val)
		assert.Equal(t, 3, helper.First([]*item{{val: 3}, {val: 1}, {val: 2}}).val)
	})
}
