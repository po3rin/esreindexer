package entity

import "time"

type Task struct {
	Index            string
	NumberOfReplicas int
	RefreshInterval  int
	ExpireDate       time.Time
	Status           Status
}

type Status int

const (
	Running Status = iota + 1 // <- 1から始める
	Done
)
