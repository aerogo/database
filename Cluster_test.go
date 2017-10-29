package nano_test

import (
	"testing"
	"time"

	"github.com/aerogo/nano"
	"github.com/stretchr/testify/assert"
)

const nodeCount = 5

func TestClusterDataSharing(t *testing.T) {
	nodes := make([]*nano.Database, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New("test", types)

		if i == 0 {
			nodes[i].Set("User", "1", newUser(1))
		}
	}

	assert.True(t, nodes[0].IsMaster())
	time.Sleep(100 * time.Millisecond)

	for i := 1; i < nodeCount; i++ {
		user, err := nodes[i].Get("User", "1")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	}

	for i := 0; i < nodeCount; i++ {
		nodes[i].ClearAll()
		nodes[i].Close()
	}
}

func TestClusterBroadcast(t *testing.T) {
	nodes := make([]*nano.Database, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New("test", types)
	}

	assert.True(t, nodes[0].IsMaster())
	time.Sleep(100 * time.Millisecond)

	// Make sure that node #0 does not have the record
	nodes[0].Delete("User", "42")

	// Set record on node #2
	nodes[2].Set("User", "42", newUser(42))
	time.Sleep(100 * time.Millisecond)

	// Confirm that both nodes have the record now
	user, err := nodes[0].Get("User", "42")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	user, err = nodes[2].Get("User", "42")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	for i := 0; i < nodeCount; i++ {
		nodes[i].ClearAll()
		nodes[i].Close()
	}
}