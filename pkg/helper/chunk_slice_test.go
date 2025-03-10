package helper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/helper"
)

func TestChunkSize(t *testing.T) {
	chunks := helper.ChunkSlice([]int{1, 2, 3, 4, 5, 6, 7}, 3)

	assert.Len(t, chunks, 3)
	assert.Equal(t, []int{1, 2, 3}, chunks[0])
	assert.Equal(t, []int{4, 5, 6}, chunks[1])
	assert.Equal(t, []int{7}, chunks[2])
}
