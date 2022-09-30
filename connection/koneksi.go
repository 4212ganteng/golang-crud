package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)
var Konekdb *pgx.Conn //pointer line 16 konekdb
var err error
func Dbkonek() {
	// postgres://postgres:password@localhost:5432/database_name
	DatabaseUrl := "postgres://postgres:admin@localhost:5432/personal-web"

	Konekdb, err = pgx.Connect(context.Background(),DatabaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("succes conect database lurr")
}