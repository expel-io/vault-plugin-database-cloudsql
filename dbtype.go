package cloudsql

import (
	"strings"
)

type DBType string

const (
	Postgres DBType = "cloudsql-postgres"
	// add other DBTypes in the future as an example:
	// MySQL = "cloudsql-mysql"
)

var dbTypes = []DBType{
	Postgres,
}

func (d DBType) String() string {
	return string(d)
}

func FromString(str string) (DBType, bool) {
	for _, v := range dbTypes {
		if strings.ToLower(str) == v.String() {
			return v, true
		}
	}
	return "", false
}
