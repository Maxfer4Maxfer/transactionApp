package model

import (
	"time"

	"github.com/google/uuid"
)

// JobID is a ID of a particular job
type JobID struct {
	uuid.UUID `json:"id"`
}

type Job struct {
	ID         JobID         `json:"id"`
	Per        float32       `json:"per"`
	StartTime  time.Time     `json:"startTime"`
	Duration   time.Duration `json:"Duration"`
	FinishTime time.Time     `json:"finishTime"`
}

func NewJob() *Job {
	return &Job{
		ID:        JobID{uuid.New()},
		Per:       0,
		StartTime: time.Now(),
	}
}

func (j *Job) Finish() {
	j.Per = 100
	j.FinishTime = time.Now()
}
