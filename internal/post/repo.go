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

func (repo *MemoryPostRepository) GetByID(id id.ID) (Post, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
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
			Votes: []Vote{},
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
