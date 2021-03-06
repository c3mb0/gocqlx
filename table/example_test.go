// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.

// +build all integration

package table_test

import (
	"testing"

	"github.com/c3mb0/gocqlx"
	. "github.com/c3mb0/gocqlx/gocqlxtest"
	"github.com/c3mb0/gocqlx/qb"
	"github.com/c3mb0/gocqlx/table"
)

func TestExample(t *testing.T) {
	session := CreateSession(t)
	defer session.Close()

	const personSchema = `
CREATE TABLE IF NOT EXISTS gocqlx_test.person (
    first_name text,
    last_name text,
    email list<text>,
    PRIMARY KEY(first_name, last_name)
)`
	if err := ExecStmt(session, personSchema); err != nil {
		t.Fatal("create table:", err)
	}

	// metadata specifies table name and columns it must be in sync with schema.
	var personMetadata = table.Metadata{
		Name:    "person",
		Columns: []string{"first_name", "last_name", "email"},
		PartKey: []string{"first_name"},
		SortKey: []string{"last_name"},
	}

	// personTable allows for simple CRUD operations based on personMetadata.
	var personTable = table.New(personMetadata)

	// Person represents a row in person table.
	// Field names are converted to camel case by default, no need to add special tags.
	// If you want to disable a field add `db:"-"` tag, it will not be persisted.
	type Person struct {
		FirstName string
		LastName  string
		Email     []string
	}

	// Insert, bind data from struct.
	{
		p := Person{
			"Patricia",
			"Citizen",
			[]string{"patricia.citzen@gocqlx_test.com"},
		}

		stmt, names := personTable.Insert()
		q := gocqlx.Query(session.Query(stmt), names).BindStruct(p)
		if err := q.ExecRelease(); err != nil {
			t.Fatal(err)
		}
	}

	// Get by primary key.
	{
		p := Person{
			"Patricia",
			"Citizen",
			nil, // no email
		}

		stmt, names := personTable.Get() // you can filter columns too
		q := gocqlx.Query(session.Query(stmt), names).BindStruct(p)
		if err := q.GetRelease(&p); err != nil {
			t.Fatal(err)
		}

		t.Log(p)
		// stdout: {Patricia Citizen [patricia.citzen@gocqlx_test.com]}
	}

	// Load all rows in a partition to a slice.
	{
		var people []Person

		stmt, names := personTable.Select() // you can filter columns too
		q := gocqlx.Query(session.Query(stmt), names).BindMap(qb.M{"first_name": "Patricia"})

		if err := q.SelectRelease(&people); err != nil {
			t.Fatal(err)
		}

		t.Log(people)
		// stdout: [{Patricia Citizen [patricia.citzen@gocqlx_test.com]}]
	}
}
