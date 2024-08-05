module backend

go 1.18

require (
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	gopkg.in/freeeve/pgn.v1 v1.0.1 // indirect
)

replace gopkg.in/freeeve/pgn.v1 => ../../../lib/pgn
