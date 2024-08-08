package types

import "database/sql"

type ServerState struct {
	DBMap map[string]*sql.DB
}
