package models

import (
	"database/sql"

)

type Student struct {
    ID       int            `db:"id"`
    UserID   int            `db:"user_id"`
    NIM      sql.NullString `db:"nim"`
    Address  sql.NullString `db:"address"`
    Phone    sql.NullString `db:"phone_number"`
}