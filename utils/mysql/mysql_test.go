package mysql

import (
	"database/sql"
	"testing"

	"github.com/thinkeridea/go-extend/exbytes"
)

type Person struct {
	Mode sql.NullInt32
}

type Man struct {
	Id   int32
	Name []byte
	*Person

	s int32

	a int32 `db:"a"`
}

func TestStructScan(t *testing.T) {
	if c, err := Register(&MysqlDriverBasic{
		Id:       "i1",
		UserName: "tars",
		Passwd:   "tars2015",
		Host:     "172.16.116.50",
		Port:     3307,
		DB:       "userroominfo",
	}); err != nil {
		panic(err)
	} else {

		var man Man
		row := c.QueryRowx("select Id id, RoomName name, ScaleMode mode from live_room where Id=1002311")
		if row.StructScan(&man); err != nil {
			panic(err)
		}

		t.Logf("man: %+v %s\n", man, exbytes.ToString(man.Name))
	}
}
