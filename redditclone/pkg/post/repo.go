package post

import (
	"errors"
	"fakereddit/redditclone/pkg/comment"
	"fakereddit/redditclone/pkg/user"
	"log"
	"sort"
	"sync"
	"time"
)

var (
	ErrNoPost = errors.New("no comment found")
)

type PostsDataRepo struct {
	mu     *sync.RWMutex
	LastID uint32
	Data   []*Post
}

func NewPostsRepo() *PostsDataRepo {
	log.Printf("NewPostsRepo: created PostsDataRepo")
	return &PostsDataRepo{
		Data: make([]*Post, 0),
		mu:   &sync.RWMutex{},
	}
}

func (pr *PostsDataRepo) Create(post *Post) (uint32, error) {
	pr.mu.Lock()
	pr.LastID++
	post.ID = pr.LastID
	post.Comments = make([]*comment.Comment, 0)
	post.Created = time.Now().Format(time.RFC3339)
	post.Votes = make([]*SingeVote, 0)
	pr.Data = append(pr.Data, post)
	pr.mu.Unlock()
	log.Printf("Created post: %v", post.ID)
	return pr.LastID, nil
}

func (pr *PostsDataRepo) ReadAll() ([]*Post, error) {
	pr.mu.RLock()
	data := pr.Data
	pr.mu.RUnlock()
	sort.Slice(data, func(i, j int) bool {
		return data[i].Score > data[j].Score
	})
	log.Printf("List posts")
	return pr.Data, nil
}

func (pr *PostsDataRepo) ReadCategory(category string) ([]*Post, error) {
	res := make([]*Post, 0)
	pr.mu.RLock()
	for _, elem := range pr.Data {
		if elem.Category == category {
			res = append(res, elem)
		}
	}
	log.Printf("ReadCategory: '%v'", category)
	pr.mu.RUnlock()
	sort.Slice(res, func(i, j int) bool {
		return res[i].Score > res[j].Score
	})
	return res, nil
}

func (pr *PostsDataRepo) Read(id uint32) (*Post, error) {
	detect := -1
	pr.mu.RLock()
	for idx, elem := range pr.Data {
		if elem.ID == id {
			detect = idx
			break
		}
	}
	if detect < 0 {
		log.Printf("ERROR: No post: '%v'", id)
		return nil, ErrNoPost
	}
	pr.mu.RUnlock()
	log.Printf("Read post: '%v'", id)
	return pr.Data[detect], nil
}

func (pr *PostsDataRepo) ReadUser(login string) ([]*Post, error) {
	res := make([]*Post, 0)
	pr.mu.RLock()
	for _, elem := range pr.Data {
		if elem.Author.Username == login {
			res = append(res, elem)
		}
	}
	log.Printf("ReadUser: '%v'", login)
	pr.mu.RUnlock()
	sort.Slice(res, func(i, j int) bool {
		return res[i].Score > res[j].Score
	})
	return res, nil
}

func (pr *PostsDataRepo) UpVote(id uint32, u *user.User) (*Post, error) {
	pr.mu.RLock()
	detect := -1
	for idx, elem := range pr.Data {
		if elem.ID == id {
			detect = idx
			break
		}
	}

	if detect < 0 {
		log.Printf("UpVote: no post '%v'", id)
		return nil, ErrNoPost
	}

	voteIdx := -1
	for idx, elem := range pr.Data[detect].Votes {
		if elem.UserID == u.ID {
			voteIdx = idx
			break
		}
	}

	if voteIdx < 0 {
		pr.Data[detect].Votes = append(pr.Data[detect].Votes,
			&SingeVote{
				Vote:   UpVote,
				UserID: u.ID,
				PostID: id,
			})
		pr.Data[detect].Score++
		pr.Data[detect].UpvotePercentage = UpVotePer(pr.Data[detect])
	} else {
		pr.Data[detect].Votes[voteIdx].Vote = UpVote
		pr.Data[detect].Score = SetScore(pr.Data[detect])
		pr.Data[detect].UpvotePercentage = UpVotePer(pr.Data[detect])
	}

	log.Printf("UpVoted: post_'%v'", id)

	pr.mu.RUnlock()
	return pr.Data[detect], nil
}

func (pr *PostsDataRepo) UnVote(id uint32, u *user.User) (*Post, error) {
	pr.mu.RLock()
	detect := -1
	for idx, elem := range pr.Data {
		if elem.ID == id {
			detect = idx
			break
		}
	}

	if detect < 0 {
		log.Printf("UnVote: no post '%v'", id)
		return nil, ErrNoPost
	}

	voteIdx := -1
	for idx, elem := range pr.Data[detect].Votes {
		if elem.UserID == u.ID {
			voteIdx = idx
			break
		}
	}

	if voteIdx < len(pr.Data[detect].Votes)-1 {
		copy(pr.Data[detect].Votes[voteIdx:], pr.Data[detect].Votes[voteIdx+1:])
	}
	pr.Data[detect].Votes[len(pr.Data[detect].Votes)-1] = nil
	pr.Data[detect].Votes = pr.Data[detect].Votes[:len(pr.Data[detect].Votes)-1]

	pr.Data[detect].Score = SetScore(pr.Data[detect])
	pr.Data[detect].UpvotePercentage = UpVotePer(pr.Data[detect])
	log.Printf("UnVoted: post_'%v'", id)

	pr.mu.RUnlock()
	return pr.Data[detect], nil
}

func (pr *PostsDataRepo) DownVote(id uint32, u *user.User) (*Post, error) {
	pr.mu.RLock()
	detect := -1
	for idx, elem := range pr.Data {
		if elem.ID == id {
			detect = idx
			break
		}
	}

	if detect < 0 {
		log.Printf("DownVote: no post '%v'", id)
		return nil, ErrNoPost
	}

	voteIdx := -1
	for idx, elem := range pr.Data[detect].Votes {
		if elem.UserID == u.ID {
			voteIdx = idx
			break
		}
	}

	if voteIdx < 0 {
		pr.Data[detect].Votes = append(pr.Data[detect].Votes,
			&SingeVote{
				Vote:   DownVote,
				UserID: u.ID,
				PostID: id,
			})
		pr.Data[detect].Score = SetScore(pr.Data[detect])
		pr.Data[detect].UpvotePercentage = UpVotePer(pr.Data[detect])
	} else {
		pr.Data[detect].Votes[voteIdx].Vote = DownVote
		pr.Data[detect].Score = SetScore(pr.Data[detect])
		pr.Data[detect].UpvotePercentage = UpVotePer(pr.Data[detect])
	}

	log.Printf("DownVoted: post_'%v'", id)
	pr.mu.RUnlock()
	return pr.Data[detect], nil
}

func (pr *PostsDataRepo) Delete(id uint32) (bool, error) {
	pr.mu.Lock()
	detect := -1
	for idx, elem := range pr.Data {
		if elem.ID == id {
			detect = idx
			continue
		}
	}

	if detect < 0 {
		log.Printf("Post Delete, can't find post %v", id)
		return false, ErrNoPost
	}

	if detect < len(pr.Data)-1 {
		copy(pr.Data[detect:], pr.Data[detect+1:])
	}
	pr.Data[len(pr.Data)-1] = nil
	pr.Data = pr.Data[:len(pr.Data)-1]
	pr.mu.Unlock()
	log.Printf("Deleted post: post_%v", id)
	return true, nil
}

func UpVotePer(p *Post) int {
	count := 0
	for _, elem := range p.Votes {
		if elem.Vote == UpVote {
			count++
		}
	}

	if len(p.Votes) == 0 {
		return 0
	}

	return int(float32(count) / float32(len(p.Votes)) * 100)
}

func SetScore(p *Post) int {
	count := 0
	for _, elem := range p.Votes {
		if elem.Vote == UpVote {
			count++
		} else {
			count--
		}
	}

	return count
}
