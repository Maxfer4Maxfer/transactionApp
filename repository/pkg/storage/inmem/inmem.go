package inmem

import (
	repo "repository/pkg/model"
	"sync"

	"github.com/google/uuid"
)

// New create in memory repository for storing nodes
func New() NodeStorage {
	return NodeStorage{
		nodes: make(map[repo.NodeID]repo.Node),
	}
}

// NodeStorage implements in memory storage of nodes
type NodeStorage struct {
	mtx   sync.RWMutex
	nodes map[repo.NodeID]repo.Node
}

func (ns NodeStorage) NewNode(n repo.Node) (repo.NodeID, error) {
	ns.mtx.Lock()
	defer ns.mtx.Unlock()

	id := uuid.New()
	n.ID = repo.NodeID{id}
	n.JobsCount = 0
	ns.nodes[n.ID] = n

	return n.ID, nil
}
func (ns NodeStorage) SaveNode(n repo.Node) error {
	ns.mtx.Lock()
	defer ns.mtx.Unlock()

	ns.nodes[n.ID] = n

	return nil
}

func (ns NodeStorage) GetAllNodes() ([]repo.Node, error) {
	ns.mtx.RLock()
	defer ns.mtx.RUnlock()

	result := make([]repo.Node, 0)
	for id := range ns.nodes {
		result = append(result, ns.nodes[id])
	}

	return result, nil
}

func (ns NodeStorage) DeleteNode(id repo.NodeID) {
	ns.mtx.Lock()
	defer ns.mtx.Unlock()
	delete(ns.nodes, id)
}
