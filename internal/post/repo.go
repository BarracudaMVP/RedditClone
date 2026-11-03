package post

import "sync"

type PostRepository interface {
}
type MemoryPostRepository struct {
	Posts []Post
	mu    sync.RWMutex
}

func NewMemoryPostRepository() PostRepository {
	return &MemoryPostRepository{Posts: make([]Post, 0)}
}
