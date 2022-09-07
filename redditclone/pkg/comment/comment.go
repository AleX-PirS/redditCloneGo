package comment

import (
	"fakereddit/redditclone/pkg/user"
)

type Comment struct {
	ID      uint32     `json:"id"`
	Author  *user.User `json:"author"`
	Created string     `json:"created"`
	Body    string     `json:"body"`
	PostID  uint32     `json:"-"`
}

type CommentsRepo interface {
	Create(comm *Comment) (uint32, error)
	ReadAll(postID uint32) ([]*Comment, error)
	Delete(postID, commentID uint32) (bool, error)
	List() (map[uint32][]*Comment, error)
}
