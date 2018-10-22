package testpgdb

import (
	"testing"
)

func TestMakeDB(t *testing.T) {
	tdb, cleanup := MakeDB(t)
	defer cleanup()

	db := tdb.SqlxDial(t, "user=postgres")
	defer db.Close()

	_, err := db.Exec(`CREATE TABLE foo (
		id    integer PRIMARY KEY,
		name   varchar(40) NOT NULL
   );`)
	if err != nil {
		t.Fatal(err)
	}
}
