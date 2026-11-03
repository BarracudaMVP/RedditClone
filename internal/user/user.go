package user

import "RedditClone/internal/id"

type User struct {
	ID       id.ID    `json:"id"`
	Username Username `json:"username"`
	Admin    bool     `json:"admin,omitempty"`
}
type Username string
