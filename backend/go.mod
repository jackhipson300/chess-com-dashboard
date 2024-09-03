module backend

go 1.18

require (
	github.com/mattn/go-sqlite3 v1.14.22
	gopkg.in/freeeve/pgn.v1 v1.0.1
)

require github.com/rs/cors v1.11.1

replace gopkg.in/freeeve/pgn.v1 => ../../../lib/pgn
