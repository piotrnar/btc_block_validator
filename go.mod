module main

go 1.16

require (
	github.com/piotrnar/gocoin v0.0.0
)

replace (
   github.com/piotrnar/gocoin => ../gocoin // place your gocoin one level up
)
