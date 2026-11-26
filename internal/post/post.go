package post

import (
	"RedditClone/internal/id"
	"RedditClone/internal/user"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalidCategoryLength = errors.New("incorrect category length")
var ErrInvalidCommentLength = errors.New("incorrect comment length")
var ErrInvalidTitleLength = errors.New("incorrect title length")
var ErrInvalidTitleFormat = errors.New("incorrect symbols in title")
var ErrInvalidPostType = errors.New("invalid post type")
var ErrInvalidPostTextLength = errors.New("incorrect post text length")
var ErrInvalidPostURLFormat = errors.New("incorrect post url")
var ErrCommentNotFound = errors.New("comment not found")
var ErrPostNotFound = errors.New("post not found")
var ErrInvalidPostId = errors.New("invalid post id")

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
	MinLengthPostText = 4
	MinLengthPostURL  = 0
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

func ParseCategory(value string) (Category, error) {
	if err := ValidateCategory(value); err != nil {
		return "", err
	}
	category := Category(strings.ToLower(value))
	return category, nil
}
func ValidateCategory(value string) error {
	length := utf8.RuneCountInString(value)
	if length < MinLengthCategory {
		return ErrInvalidCategoryLength
	}
	return nil
}
func ValidatePostText(value string) error {
	length := utf8.RuneCountInString(value)
	if length < MinLengthPostText {
		return ErrInvalidPostTextLength
	}
	return nil
}

func ValidatePostURL(value string) error {
	u, err := url.Parse(value)
	if err != nil {
		return ErrInvalidPostURLFormat
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return ErrInvalidPostURLFormat
	}
	return nil
}
func (p *Post) IncrementViews() {
	p.Views++
}

func (p *Post) GetCommentByID(id id.ID) (Comment, bool) {
	for i := range p.Comments {
		if p.Comments[i].ID == id {
			return p.Comments[i], true
		}
	}
	return Comment{}, false
}

func (p *Post) CheckPermissonForComment(comment Comment, actor user.User) bool {
	return actor.Admin || p.Author.ID == actor.ID || comment.Author.ID == actor.ID
}
func (p *Post) CheckPermissonForPost(actor user.User) bool {
	return actor.Admin || p.Author.ID == actor.ID
}
func (p *Post) SetUserVote(user user.User, value VoteValue) {
	p.Votes = slices.Clone(p.Votes)
	for i := range p.Votes {
		if p.Votes[i].User == user.ID {
			p.Votes[i].Vote = value
			return
		}
	}
	vote := Vote{User: user.ID, Vote: value}
	votes := slices.Clone(p.Votes)
	votes = append(votes, vote)
	p.Votes = votes

}
func (p *Post) UserUnvote(user user.User) {
	votes := slices.Clone(p.Votes)
	p.Votes = slices.DeleteFunc(votes, func(v Vote) bool {
		return v.User == user.ID
	})
}
func (p *Post) CalculateUpvotePercentage() {
	if len(p.Votes) == 0 {
		p.UpvotePercentage = 0
		return
	}
	var upvotes int
	for i := range p.Votes {
		if p.Votes[i].Vote == Upvote {
			upvotes++
		}
	}
	p.UpvotePercentage = int64(100 * upvotes / len(p.Votes))
}
func (p *Post) UpdateScore() {
	p.Score = 0
	for i := range p.Votes {
		p.Score += int64(p.Votes[i].Vote)
	}
}
