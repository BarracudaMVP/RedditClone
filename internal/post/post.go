package post

import (
	"RedditClone/internal/id"
	"RedditClone/internal/user"
	"time"
)

type Post struct {
	ID               id.ID     `json:"id"`
	Title            string    `json:"title"`
	Category         Category  `json:"category"`
	Type             PostType  `json:"type"`
	URL              string    `json:"url,omitzero"`
	Text             string    `json:"text,omitzero"`
	Author           user.User `json:"author"`
	Score            int64     `json:"score"`
	Votes            []Vote    `json:"votes"`
	Comments         []Comment `json:"comments"`
	Created          time.Time `json:"created,omitzero"`
	Views            int64     `json:"views"`
	UpvotePercentage int64     `json:"upvotePercentage"`
}

type Category string

type PostType string

const (
	LinkPostType PostType = "link"
	TextPostType PostType = "text"
)

type Vote struct {
	User id.ID     `json:"user"`
	Vote VoteValue `json:"vote"`
}
type VoteValue int64

const (
	Downvote VoteValue = -1
	Upvote   VoteValue = 1
)

type Comment struct {
	ID      id.ID     `json:"id"`
	Author  user.User `json:"author"`
	Body    string    `json:"body"`
	Created time.Time `json:"created,omitzero"`
}
