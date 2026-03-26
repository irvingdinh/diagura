package entity

import "localhost/app/core/sqlite/orm"

type User struct {
	orm.BaseModel
	Email string `db:"email" json:"email"`
	Name  string `db:"name"  json:"name"`
}
