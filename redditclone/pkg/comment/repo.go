package comment

import (
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrNoComm = errors.New("no comment found")
)

type CommentsDataRepo struct {
	mu     *sync.RWMutex
	LastID map[uint32]uint32
	Data   map[uint32][]*Comment
}

func NewCommentsRepo() *CommentsDataRepo {
	log.Printf("NewCommentsRepo: created CommentsDataRepo")
	return &CommentsDataRepo{
		Data:   make(map[uint32][]*Comment),
		mu:     &sync.RWMutex{},
		LastID: make(map[uint32]uint32),
	}
}

func (cr *CommentsDataRepo) Create(comm *Comment) (uint32, error) {
	cr.mu.Lock()
	cr.LastID[comm.PostID]++
	comm.ID = cr.LastID[comm.PostID]
	comm.Created = time.Now().Format(time.RFC3339)
	cr.Data[comm.PostID] = append(cr.Data[comm.PostID], comm)
	cr.mu.Unlock()
	log.Printf("Created comment: %v", comm.ID)
	return cr.LastID[comm.PostID], nil
}

func (cr *CommentsDataRepo) ReadAll(postID uint32) ([]*Comment, error) {
	log.Printf("List comments, post %v", postID)
	return cr.Data[postID], nil
}

func (cr *CommentsDataRepo) List() (map[uint32][]*Comment, error) {
	log.Printf("List comments")
	return cr.Data, nil
}

func (cr *CommentsDataRepo) Delete(postID, commentID uint32) (bool, error) {
	cr.mu.Lock()
	detect := -1
	for idx, elem := range cr.Data[postID] {
		if elem.ID == commentID {
			detect = idx
			continue
		}
	}
	if detect < 0 {
		log.Printf("ERROR: Comment Delete, can't find post %v, id %v", postID, commentID)
		return false, ErrNoComm
	}

	if detect < len(cr.Data[postID])-1 {
		copy(cr.Data[postID][detect:], cr.Data[postID][detect+1:])
	}
	cr.Data[postID][len(cr.Data[postID])-1] = nil
	cr.Data[postID] = cr.Data[postID][:len(cr.Data[postID])-1]
	cr.mu.Unlock()
	log.Printf("Deleted post comment: postID %v, commID %v", postID, commentID)
	return true, nil
}
