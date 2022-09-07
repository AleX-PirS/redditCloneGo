package main

import (
	"fakereddit/redditclone/pkg/comment"
	"fakereddit/redditclone/pkg/handlers"
	"fakereddit/redditclone/pkg/middleware"
	"fakereddit/redditclone/pkg/post"
	"fakereddit/redditclone/pkg/session"
	"fakereddit/redditclone/pkg/user"
	"github.com/gorilla/mux"
	"net/http"
)

var (
	postsRepo = post.NewPostsRepo()
	userRepo  = user.NewUsersRepo()
	commRepo  = comment.NewCommentsRepo()
)

func main() {
	sm := session.NewSessionsManager()

	userHandler := &handlers.UserHandler{
		UserRepo: userRepo,
		Sessions: sm,
	}

	handler := &handlers.PostHandler{
		Sessions:    sm,
		PostRepo:    postsRepo,
		CommentRepo: commRepo,
	}

	Handler := http.StripPrefix("/static/", http.FileServer(http.Dir("../../static/")))

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(Handler)
	r.HandleFunc("/", userHandler.Index)
	r.HandleFunc("/api/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/api/posts/", handler.GetAll).Methods("GET")
	r.HandleFunc("/api/posts", handler.NewPost).Methods("POST")
	r.HandleFunc("/api/posts/{CATEGORY_NAME}", handler.GetByCategory).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}", handler.Get).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}", handler.NewComm).Methods("POST")
	r.HandleFunc("/api/post/{POST_ID}/{COMMENT_ID}", handler.DeleteComm).Methods("DELETE")
	r.HandleFunc("/api/post/{POST_ID}/upvote", handler.Upvote).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/downvote", handler.DownVote).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/unvote", handler.UnVote).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}", handler.DeletePost).Methods("DELETE")
	r.HandleFunc("/api/user/{USER_LOGIN}", handler.GetByUser).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../../static/html/index.html")
	})

	muxer := middleware.Panic(r)
	muxer = middleware.AccessLog(muxer)

	addr := ":8081"
	err := http.ListenAndServe(addr, muxer)
	if err != nil {
		return
	}
}
