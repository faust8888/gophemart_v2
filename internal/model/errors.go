package model

import "errors"

// ErrLoginTaken возвращается при попытке создать пользователя с уже занятым логином.
var ErrLoginTaken = errors.New("login already taken")
