module backend

go 1.18

require (
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.22
	gopkg.in/freeeve/pgn.v1 v1.0.1
)

require github.com/google/uuid v1.6.0 // indirect

replace gopkg.in/freeeve/pgn.v1 => ../../../lib/pgn
