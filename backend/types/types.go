package types

import (
	"database/sql"
	"sync"
)

type LockedResource[T any] struct {
	Mu       sync.Mutex
	Resource *T
}

type LockedDB LockedResource[sql.DB]

func NewLockedDB(db *sql.DB) *LockedDB {
	return &LockedDB{
		Mu:       sync.Mutex{},
		Resource: db,
	}
}

type SetupStatus string

const (
	SetupStatusPending  SetupStatus = "Pending"
	SetupStatusUpdating SetupStatus = "Updating"
	SetupStatusStarted  SetupStatus = "Started"
	SetupStatusComplete SetupStatus = "Complete"
	SetupStatusFailed   SetupStatus = "Failed"
)

type DBMap map[string]*LockedDB
type SetupStatuses LockedResource[map[string]SetupStatus]

type ServerState struct {
	DBMap         DBMap
	SetupStatuses SetupStatuses
}

func NewServerState() *ServerState {
	dbMap := make(map[string]*LockedDB)

	setupStatuses := make(map[string]SetupStatus)
	return &ServerState{
		DBMap: dbMap,
		SetupStatuses: SetupStatuses{
			Mu:       sync.Mutex{},
			Resource: &setupStatuses,
		},
	}
}
