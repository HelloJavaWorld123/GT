package myapp

import (
	"database/sql"
	"errors"
)

func (tx *Tx) CreatUser(info *UserInfo) (sql.Result, error) {
	if info == nil {
		return nil, errors.New("user must not bee null")
	}

	if info.UserName == "" {
		return nil, errors.New("userName must not bee null")
	}

	result, err := tx.Exec("insert into user_info values (....)", "")
	if err != nil {
		return nil, err
	}
	return result, err
}
