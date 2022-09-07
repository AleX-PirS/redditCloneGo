package handlers

import (
	"bytes"
	"encoding/json"
	"fakereddit/redditclone/pkg/comment"
	"fakereddit/redditclone/pkg/post"
	"fakereddit/redditclone/pkg/session"
	"fakereddit/redditclone/pkg/user"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	ToReplace   = `"data":`
	TextReplace = `"text":`
	URLReplace  = `"url":`

	Success = "success"

	JSONContentType = "application/json"
)

const (
	UnexpectedErrorTXT = `something goes wrong`
	MarshalErrorTXT    = `invalid data`
	UnmarshalErrorTXT  = `invalid json`
)

type PostHandler struct {
	PostRepo    post.PostsRepo
	CommentRepo comment.CommentsRepo
	Sessions    *session.SessionsManager
}

type PostForm struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	Text     string `json:"text"`
	URL      string `json:"url"`
}

type CommForm struct {
	Comment string `json:"comment"`
}

type ChangeForm struct {
	Message string `json:"message"`
}

type ErrorMsg struct {
	Errors []*DetailError `json:"errors"`
}

type DetailError struct {
	Location string `json:"location"` // "body"
	Message  string `json:"msg"`      // "is required"
	Param    string `json:"param"`    // "comment"
}

func (h *PostHandler) GetAll(w http.ResponseWriter, _ *http.Request) {
	posts, err := h.PostRepo.ReadAll()
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}
	comments, err := h.CommentRepo.List()
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, elem := range posts {
		elem.Comments = comments[elem.ID]
	}

	res, err := json.Marshal(posts)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}
	res = Normalize(res, len(posts))
	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) NewPost(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != JSONContentType {
		JSONErrorBuilder(w, "unknown payload", http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}
	data := &PostForm{}
	err = json.Unmarshal(body, data)
	if err != nil {
		JSONErrorBuilder(w, UnmarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	sess, err := h.Sessions.Check(r)
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	newPost := &post.Post{
		Author:   &user.User{ID: sess.UserID, Username: sess.UserName},
		Type:     data.Type,
		Title:    data.Title,
		Category: data.Category,
	}

	if data.Type == "text" {
		newPost.Data = data.Text
	} else if data.Type == "link" {
		newPost.Data = data.URL
	}

	ID, err := h.PostRepo.Create(newPost)
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	postBD, err := h.PostRepo.UpVote(ID, newPost.Author)
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(postBD)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	res = Normalize(res, 1)

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/post/"))
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	ok, err := h.PostRepo.Delete(uint32(postID))
	if !ok || err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(ChangeForm{Message: Success})
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) DeleteComm(w http.ResponseWriter, r *http.Request) {
	data := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/post/"), "/")

	postID, err := strconv.Atoi(data[0])
	if err != nil {
		JSONErrorBuilder(w, "invalid post id", http.StatusBadRequest)
		return
	}
	commID, err := strconv.Atoi(data[1])
	if err != nil {
		JSONErrorBuilder(w, "invalid comment id", http.StatusBadRequest)
		return
	}

	ok, err := h.CommentRepo.Delete(uint32(postID), uint32(commID))
	if !ok || err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	postByID, err := h.PostRepo.Read(uint32(postID))
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	comments, err := h.CommentRepo.ReadAll(uint32(postID))
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}
	postByID.Comments = comments

	res, err := json.Marshal(postByID)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	res = Normalize(res, 1)

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) NewComm(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != JSONContentType {
		JSONErrorBuilder(w, "unknown payload", http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}
	data := &CommForm{}
	err = json.Unmarshal(body, data)
	if err != nil {
		JSONErrorBuilder(w, UnmarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	if data.Comment == "" {
		res, inErr := json.Marshal(&ErrorMsg{
			Errors: []*DetailError{
				{
					Location: "body",
					Param:    "comment",
					Message:  "is required",
				},
			},
		})
		if inErr != nil {
			JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
			return
		}
		http.Error(w, string(res), http.StatusUnprocessableEntity)
		return
	}

	defer r.Body.Close()

	postID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/post/"))
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusBadRequest)
		return
	}

	sess, err := h.Sessions.Check(r)
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	_, err = h.CommentRepo.Create(&comment.Comment{
		Body:   data.Comment,
		PostID: uint32(postID),
		Author: &user.User{
			ID:       sess.UserID,
			Username: sess.UserName,
		},
	})

	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	postByID, err := h.PostRepo.Read(uint32(postID))
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}

	comments, err := h.CommentRepo.ReadAll(uint32(postID))
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}
	postByID.Comments = comments

	res, err := json.Marshal(postByID)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	res = Normalize(res, 1)

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) Get(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/post/"))
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusBadRequest)
		return
	}
	postByID, err := h.PostRepo.Read(uint32(postID))
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusBadRequest)
		return
	}

	postByID.Views++

	res, err := json.Marshal(postByID)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	res = Normalize(res, 1)

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) GetByCategory(w http.ResponseWriter, r *http.Request) {
	categoryName := strings.TrimPrefix(r.URL.Path, "/api/posts/")

	categoryPosts, err := h.PostRepo.ReadCategory(categoryName)
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := json.Marshal(categoryPosts)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	res = Normalize(res, len(categoryPosts))

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	login := strings.TrimPrefix(r.URL.Path, "/api/user/")

	userPosts, err := h.PostRepo.ReadUser(login)
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := json.Marshal(userPosts)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	res = Normalize(res, len(userPosts))

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) Upvote(w http.ResponseWriter, r *http.Request) {
	data := strings.TrimPrefix(r.URL.Path, "/api/post/")
	data = strings.TrimRight(data, "/upvote")
	postID, err := strconv.Atoi(data)

	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess, err := h.Sessions.Check(r)
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusInternalServerError)
		return
	}
	upPost, err := h.PostRepo.UpVote(uint32(postID), &user.User{ID: sess.UserID, Username: sess.UserName})

	res, err := json.Marshal(upPost)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	res = Normalize(res, 1)

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) DownVote(w http.ResponseWriter, r *http.Request) {
	data := strings.TrimPrefix(r.URL.Path, "/api/post/")
	data = strings.TrimRight(data, "/downvote")
	postID, err := strconv.Atoi(data)

	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess, err := h.Sessions.Check(r)
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusInternalServerError)
		return
	}
	upPost, err := h.PostRepo.DownVote(uint32(postID), &user.User{ID: sess.UserID, Username: sess.UserName})

	res, err := json.Marshal(upPost)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	res = Normalize(res, 1)

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func (h *PostHandler) UnVote(w http.ResponseWriter, r *http.Request) {
	data := strings.TrimPrefix(r.URL.Path, "/api/post/")
	data = strings.TrimRight(data, "/unvote")
	postID, err := strconv.Atoi(data)

	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess, err := h.Sessions.Check(r)
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusInternalServerError)
		return
	}
	upPost, err := h.PostRepo.UnVote(uint32(postID), &user.User{ID: sess.UserID, Username: sess.UserName})

	res, err := json.Marshal(upPost)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	res = Normalize(res, 1)

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
}

func Normalize(data []byte, l int) []byte {
	indexFlag := 0
	for i := 0; i < l; i++ {
		indexFlag = bytes.Index(data[indexFlag:], []byte(`"type":"`)) + len([]byte(`"type":"`)) + indexFlag
		if data[indexFlag] == 'l' {
			data = bytes.Replace(data, []byte(ToReplace), []byte(URLReplace), 1)
			continue
		}
		data = bytes.Replace(data, []byte(ToReplace), []byte(TextReplace), 1)
	}
	return data
}
