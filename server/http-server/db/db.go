package db

import (
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

var Session *gocql.Session

func InitDb() {
	var err error
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "channels_db"
	Session, err = cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database initialized")
}
