package post

import (
	"RedditClone/internal/id"
	"RedditClone/internal/user"
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrPostNotFound = errors.New("post not found")

type PostRepository interface {
	GetAll() []Post
	GetByID(id id.ID) (Post, error)
	GetCategory(category Category) []Post
	GetByAuthorID(author id.ID) []Post
	Create(req CreatePostRequest, author user.User) (Post, error)
}
type MemoryPostRepository struct {
	Posts []Post
	mu    sync.RWMutex
}

func (repo *MemoryPostRepository) existsByID(id id.ID) bool {
	for i := range repo.Posts {
		if id == repo.Posts[i].ID {
			return true
		}
	}
	return false
}

func (repo *MemoryPostRepository) GetAll() []Post {
	repo.mu.RLock()
	posts := make([]Post, len(repo.Posts))
	copy(posts, repo.Posts)
	repo.mu.RUnlock()
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score > posts[j].Score
	})
	return posts
}

func (repo *MemoryPostRepository) GetCategory(c Category) []Post {
	repo.mu.RLock()
	posts := make([]Post, len(repo.Posts))
	copy(posts, repo.Posts)
	repo.mu.RUnlock()
	result := make([]Post, 0)
	for i := range posts {
		if c == posts[i].Category {
			result = append(result, posts[i])
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	return result
}

func (repo *MemoryPostRepository) GetByAuthorID(author id.ID) []Post {
	repo.mu.RLock()
	posts := make([]Post, len(repo.Posts))
	copy(posts, repo.Posts)
	repo.mu.RUnlock()
	result := make([]Post, 0)
	for i := range posts {
		if author == posts[i].Author.ID {
			result = append(result, posts[i])
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Created.After(result[j].Created)
	})
	return result
}
func (repo *MemoryPostRepository) GetByID(id id.ID) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for i := range repo.Posts {
		if repo.Posts[i].ID == id {
			repo.Posts[i].IncrementViews()
			return repo.Posts[i], nil
		}
	}
	return Post{}, ErrPostNotFound
}

func NewMemoryPostRepository() *MemoryPostRepository {
	return &MemoryPostRepository{Posts: []Post{
		{
			ID:       "507f1f77bcf86cd799439012",
			Title:    "The Go programming language",
			Category: "programming",
			Type:     "link",
			URL:      "https://go.dev/",
			Author: user.User{
				ID:       "507f1f77bcf86cd799439011",
				Username: "alice",
			},
			Score: 1,
			Votes: []Vote{
				{
					User: "507f1f77bcf86cd799439011",
					Vote: 1,
				},
			},
			Comments:         []Comment{},
			Created:          time.Date(2026, time.November, 2, 10, 0, 0, 0, time.UTC),
			Views:            0,
			UpvotePercentage: 100,
		},
		{
			ID:       "507f1f77bcf86cd799439013",
			Title:    "Hello",
			Category: "funny",
			Type:     "text",
			Text:     "Hello, world!",
			Author: user.User{
				ID:       "507f1f77bcf86cd799439011",
				Username: "alice",
			},
			Score: 0,
			Votes: []Vote{
				{
					User: "507f1f77bcf86cd799439011",
					Vote: 1,
				},
			},
			Comments: []Comment{
				{
					ID: "507f1f77bcf86cd799439016",
					Author: user.User{
						ID:       "507f1f77bcf86cd799439011",
						Username: "alice",
					},
					Body:    "LOL",
					Created: time.Date(2026, time.November, 1, 11, 0, 0, 0, time.UTC),
				},
			},
			Created:          time.Date(2026, time.November, 1, 10, 0, 0, 0, time.UTC),
			Views:            0,
			UpvotePercentage: 100,
		},
	},
	}
}
func (repo *MemoryPostRepository) Create(req CreatePostRequest, author user.User) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	err := ValidateTitle(req.Title)
	if err != nil {
		return Post{}, err
	}
	category, err := ParseCategory(string(req.Category))
	if err != nil {
		return Post{}, err
	}
	switch req.Type {
	case LinkPostType:
		if err := ValidatePostURL(req.URL); err != nil {
			return Post{}, err
		}
	case TextPostType:
		if err := ValidatePostText(req.Text); err != nil {
			return Post{}, err
		}
	default:
		return Post{}, ErrInvalidPostType

	}
	var uid id.ID
	for {
		uid, err = id.NewID()
		if err != nil {
			return Post{}, err
		}
		if !repo.existsByID(uid) {
			break
		}
	}
	post := Post{
		ID:       uid,
		Title:    req.Title,
		Category: category,
		Type:     req.Type,
		URL:      req.URL,
		Text:     req.Text,
		Author:   author,
		Score:    1,
		Votes: []Vote{{
			User: author.ID,
			Vote: 1,
		},
		},
		Comments:         []Comment{},
		Created:          time.Now(),
		Views:            0,
		UpvotePercentage: 100,
	}
	repo.Posts = append(repo.Posts, post)
	return post, nil
}
