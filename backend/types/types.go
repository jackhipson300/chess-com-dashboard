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

type DBMap map[string]*LockedDB
type SetupRequests LockedResource[map[string]string]

type ServerState struct {
	DBMap         DBMap
	SetupRequests SetupRequests
}

func NewServerState() *ServerState {
	dbMap := make(map[string]*LockedDB)

	setupRequests := make(map[string]string)
	return &ServerState{
		DBMap: dbMap,
		SetupRequests: SetupRequests{
			Mu:       sync.Mutex{},
			Resource: &setupRequests,
		},
	}
}
