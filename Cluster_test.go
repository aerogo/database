package nano_test

import (
	"testing"
	"time"

	"github.com/aerogo/nano"
	"github.com/stretchr/testify/assert"
)

const nodeCount = 5

func TestClusterClose(t *testing.T) {
	nodes := make([]*nano.Node, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New()
	}

	time.Sleep(100 * time.Millisecond)

	for i := 0; i < nodeCount; i++ {
		nodes[i].Close()
	}
}

func TestClusterDataSharing(t *testing.T) {
	// Create cluster where the server has initial data
	nodes := make([]*nano.Node, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New()
		nodes[i].Namespace("test", types...)

		if i == 0 {
			assert.True(t, nodes[i].IsServer())
			nodes[i].Namespace("test").Set("User", "100", newUser(100))
		} else {
			assert.False(t, nodes[i].IsServer())
		}
	}

	time.Sleep(300 * time.Millisecond)

	// Check data on client nodes
	for i := 1; i < nodeCount; i++ {
		user, err := nodes[i].Namespace("test").Get("User", "100")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	}

	for i := 0; i < nodeCount; i++ {
		nodes[i].Clear()
		nodes[i].Close()
	}
}

func TestClusterSet(t *testing.T) {
	// Create cluster
	nodes := make([]*nano.Node, nodeCount, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodes[i] = nano.New()
		nodes[i].Namespace("test", types...)

		if i == 0 {
			assert.True(t, nodes[i].IsServer())
		} else {
			assert.False(t, nodes[i].IsServer())
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Make sure that node #0 does not have the record
	nodes[0].Namespace("test").Delete("User", "42")

	// Set record on node #2
	nodes[2].Namespace("test").Set("User", "42", newUser(42))
	time.Sleep(200 * time.Millisecond)

	// Confirm that all nodes have the record now
	for i := 0; i < nodeCount; i++ {
		user, err := nodes[i].Namespace("test").Get("User", "42")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	}

	for i := 0; i < nodeCount; i++ {
		nodes[i].Clear()
		nodes[i].Close()
	}
}
