package model

import (
	"time"

	"github.com/google/uuid"
)


// Node represents an executer instance machine
type Node struct {
	ID        NodeID `json:"id"`
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Port      string `json:"port"`
	JobsCount int    `json:"jobscount"`
	Jobs      []Job  `json:"jobs"`
}

// NodeID is a ID of particular node
type NodeID struct {
	uuid.UUID `json:"id"`
}

type JobID struct {
	uuid.UUID `json:"id"`
}

type Job struct {
	ID         JobID     `json:"id"`
	Per        float32   `json:"per"`
	Duration   float32   `json:"duration"`
	StartTime  time.Time `json:"startTime"`
	FinishTime time.Time `json:"finishTime"`
}
