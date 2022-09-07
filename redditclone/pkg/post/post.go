package post

import (
	"fakereddit/redditclone/pkg/comment"
	"fakereddit/redditclone/pkg/user"
)

const (
	UpVote   = 1
	NoVote   = 0
	DownVote = -1
)

type Post struct {
	ID               uint32             `json:"id"`
	Score            int                `json:"score"`
	Views            uint32             `json:"views"`
	Type             string             `json:"type"`
	Title            string             `json:"title"`
	Author           *user.User         `json:"author"`
	Category         string             `json:"category"`
	Created          string             `json:"created"`
	Comments         []*comment.Comment `json:"comments"`
	Data             string             `json:"data"` // text/url
	UpvotePercentage int                `json:"upvotePercentage"`
	Votes            []*SingeVote       `json:"votes"`
}

type SingeVote struct {
	PostID uint32 `json:"-"`
	UserID uint32 `json:"user"`
	Vote   int    `json:"vote"`
}

type PostsRepo interface {
	Create(post *Post) (uint32, error)
	ReadAll() ([]*Post, error)
	ReadCategory(category string) ([]*Post, error)
	Read(id uint32) (*Post, error)
	ReadUser(login string) ([]*Post, error)
	UpVote(id uint32, u *user.User) (*Post, error)
	DownVote(id uint32, u *user.User) (*Post, error)
	UnVote(id uint32, u *user.User) (*Post, error)
	Delete(id uint32) (bool, error)
}
