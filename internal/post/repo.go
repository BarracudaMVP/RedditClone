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
	GetAllPosts() []Post
	PostExistsByID(id id.ID) bool
	GetPostByIDAndIncrementViews(id id.ID) (Post, error)
	GetPostsByCategory(category Category) []Post
	GetPostByAuthorID(author id.ID) []Post
	CreatePost(req CreatePostRequest, author user.User) (Post, error)
	CreateComment(req CreateCommentRequest, postId id.ID, user user.User) (Post, error)
}
type MemoryPostRepository struct {
	Posts []Post
	mu    sync.RWMutex
}

func (repo *MemoryPostRepository) PostExistsByID(id id.ID) bool {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	post := repo.getPostByID(id)
	return post != nil
}

func (repo *MemoryPostRepository) GetAllPosts() []Post {
	repo.mu.RLock()
	posts := make([]Post, len(repo.Posts))
	copy(posts, repo.Posts)
	repo.mu.RUnlock()
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score > posts[j].Score
	})
	return posts
}

func (repo *MemoryPostRepository) GetPostsByCategory(c Category) []Post {
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

func (repo *MemoryPostRepository) GetPostByAuthorID(author id.ID) []Post {
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
func (repo *MemoryPostRepository) getPostByID(id id.ID) *Post {
	for i := range repo.Posts {
		if repo.Posts[i].ID == id {
			return &repo.Posts[i]
		}
	}
	return nil
}

// repo.Posts[i].IncrementViews()
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
func (repo *MemoryPostRepository) CreatePost(req CreatePostRequest, author user.User) (Post, error) {
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
	var pid id.ID
	for {
		pid, err = id.NewID()
		if err != nil {
			return Post{}, err
		}
		if repo.getPostByID(pid) == nil {
			break
		}
	}
	post := Post{
		ID:       pid,
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
func (repo *MemoryPostRepository) GetPostByIDAndIncrementViews(id id.ID) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	p := repo.getPostByID(id)
	if p == nil {
		return Post{}, ErrPostNotFound
	}
	p.IncrementViews()
	return *p, nil
}
func (repo *MemoryPostRepository) CreateComment(req CreateCommentRequest, postID id.ID, user user.User) (Post, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	p := repo.getPostByID(postID)
	if p == nil {
		return Post{}, ErrPostNotFound
	}
	err := ValidateComment(req.Comment)
	if err != nil {
		return Post{}, err
	}
	var pid id.ID
	for {
		pid, err = id.NewID()
		if err != nil {
			return Post{}, err
		}
		if _, ok := p.GetCommentByID(pid); !ok {
			break
		}
	}
	comment := Comment{
		ID:      pid,
		Author:  user,
		Body:    req.Comment,
		Created: time.Now(),
	}
	p.Comments = append(p.Comments, comment)
	return *p, nil
}
