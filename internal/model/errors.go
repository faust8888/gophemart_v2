package model

import "errors"

// ErrLoginTaken возвращается при попытке создать пользователя с уже занятым логином.
var ErrLoginTaken = errors.New("login already taken")

// ErrNotFound возвращается, если сущность не найдена в хранилище.
var ErrNotFound = errors.New("not found")

// ErrOrderAlreadyUploaded возвращается, если номер заказа уже загружен этим пользователем.
var ErrOrderAlreadyUploaded = errors.New("order already uploaded")

// ErrOrderConflict возвращается, если номер заказа уже загружен другим пользователем.
var ErrOrderConflict = errors.New("order uploaded by another user")

// ErrInsufficientFunds возвращается, если на счёте недостаточно баллов для списания.
var ErrInsufficientFunds = errors.New("insufficient funds")
