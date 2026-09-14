// Package model содержит доменные сущности накопительной системы лояльности.
package model

import "time"

// User представляет зарегистрированного пользователя системы лояльности.
type User struct {
	// ID — уникальный идентификатор пользователя.
	ID string
	// Login — уникальный логин пользователя.
	Login string
	// PasswordHash — bcrypt-хеш пароля, не предназначен для сериализации во внешние ответы.
	PasswordHash string
	// CreatedAt — время регистрации пользователя.
	CreatedAt time.Time
}
