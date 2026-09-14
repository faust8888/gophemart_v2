package model

import "errors"

// ErrLoginTaken возвращается при попытке создать пользователя с уже занятым логином.
var ErrLoginTaken = errors.New("login already taken")

// ErrNotFound возвращается, если сущность не найдена в хранилище.
var ErrNotFound = errors.New("not found")
