module github.com/r4wm/bible_api

go 1.16

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gorilla/mux v1.8.0
	github.com/mattn/go-sqlite3 v1.14.7
	github.com/r4wm/mintz5 v0.0.0-20200913071705-f9eb5b929605
	github.com/r4wm/sqlite3_kjv v0.0.0-20201005151805-2fa1bb49fb45
	github.com/sirupsen/logrus v1.8.1
)

exclude github.com/mattn/go-sqlite3 v1.10.0
