package post

import (
	"RedditClone/internal/user"
	"sort"
	"sync"
	"time"
)

type PostRepository interface {
	GetAll() []Post
}
type MemoryPostRepository struct {
	Posts []Post
	mu    sync.RWMutex
}

func (repo *MemoryPostRepository) GetAll() []Post {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	posts := make([]Post, len(repo.Posts))
	copy(posts, repo.Posts)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score > posts[j].Score
	})
	return posts
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
			ID:       "507f1f77bcf86cd799439012",
			Title:    "Hello",
			Category: "programming",
			Type:     "text",
			Text:     "Hello, world!",
			Author: user.User{
				ID:       "507f1f77bcf86cd799439011",
				Username: "alice",
			},
			Score:            0,
			Votes:            []Vote{},
			Comments:         []Comment{},
			Created:          time.Date(2026, time.November, 1, 10, 0, 0, 0, time.UTC),
			Views:            0,
			UpvotePercentage: 100,
		},
	},
	}
}
