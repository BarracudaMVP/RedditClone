package post

import (
	"RedditClone/internal/id"
	"RedditClone/internal/user"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidCategoryLength = errors.New("incorrect category length")
var ErrInvalidCommentLength = errors.New("incorrect comment length")

var ErrInvalidTitleLength = errors.New("incorrect title length")

var ErrInvalidTitleFormat = errors.New("incorrect symbols in title")

var ErrInvalidTextPostLength = errors.New("incorrect text post length")

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

const (
	MinLengthCategory = 1
	MinLengthTitle    = 1
	MaxLengthTitle    = 100
	MinLengthComment  = 1
	MaxLengthComment  = 2000
	MinLengthTextPost = 4
)

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

type CreatePostRequest struct {
	Title    string   `json:"title"`
	Category Category `json:"category"`
	Type     PostType `json:"type"`
	URL      string   `json:"url,omitzero"`
	Text     string   `json:"text,omitzero"`
}
type CreateCommentRequest struct {
	Comment string `json:"comment"`
}

func ValidateTitle(value string) error {
	length := utf8.RuneCountInString(value)
	if length < MinLengthTitle || length > MaxLengthTitle {
		return ErrInvalidTitleLength
	}
	if strings.TrimSpace(value) != value {
		return ErrInvalidTitleFormat
	}
	return nil
}
func ValidateComment(value string) error {
	length := utf8.RuneCountInString(value)
	if length < MinLengthComment || length > MaxLengthComment {
		return ErrInvalidCommentLength
	}
	return nil
}

func ValidateTextPost(value string) error {
	length := utf8.RuneCountInString(value)
	if length < MinLengthTextPost {
		return ErrInvalidTextPostLength
	}
	return nil
}
func ParseCategory(value string) (Category, error) {
	length := utf8.RuneCountInString(value)
	if length < MinLengthCategory {
		return "", ErrInvalidCategoryLength
	}
	category := Category(strings.ToLower(value))
	return category, nil
}
func ValidatePostRequest(сpr CreatePostRequest) error {}
func (p *Post) IncrementViews() {
	p.Views++
}
