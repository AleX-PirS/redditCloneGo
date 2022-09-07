package handlers

import (
	"encoding/json"
	"fakereddit/redditclone/pkg/session"
	"fakereddit/redditclone/pkg/user"
	"io/ioutil"
	"log"
	"net/http"
)

type UserHandler struct {
	UserRepo user.UsersRepo
	Sessions *session.SessionsManager
}

type JSONError struct {
	Message string `json:"message"`
}

type LogIn struct {
	Token string `json:"token"`
}

type LoginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../../static/html/index.html")
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != JSONContentType {
		JSONErrorBuilder(w, "unknown payload", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}
	data := &LoginForm{}
	err = json.Unmarshal(body, data)
	if err != nil {
		JSONErrorBuilder(w, UnmarshalErrorTXT, http.StatusInternalServerError)
	}

	defer r.Body.Close()

	u, err := h.UserRepo.CreateUser(data.Username, data.Password)

	if err == user.ErrAlreadyExist {
		JSONErrorBuilder(w, err.Error(), 409)
		return
	}
	sess, err := h.Sessions.Create(w, u.ID, data.Username)
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(LogIn{Token: sess.AccessToken})
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}

	log.Printf("created session for %v", sess.UserID)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != JSONContentType {
		JSONErrorBuilder(w, "unknown payload", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		JSONErrorBuilder(w, UnexpectedErrorTXT, http.StatusInternalServerError)
		return
	}
	data := &LoginForm{}
	err = json.Unmarshal(body, data)
	if err != nil {
		JSONErrorBuilder(w, UnmarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	u, err := h.UserRepo.Authorize(data.Username, data.Password)
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusBadRequest)
		return
	}

	sess, err := h.Sessions.Create(w, u.ID, data.Username)
	if err != nil {
		JSONErrorBuilder(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(LogIn{Token: sess.AccessToken})
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}

	_, err = w.Write(res)
	if err != nil {
		log.Printf("critical error, %v", err.Error())
		return
	}
	log.Printf("created session for %v", sess.UserID)
}

func JSONErrorBuilder(w http.ResponseWriter, inErr string, statusCode int) {
	data := JSONError{Message: inErr}
	res, err := json.Marshal(data)
	if err != nil {
		JSONErrorBuilder(w, MarshalErrorTXT, http.StatusInternalServerError)
		return
	}
	http.Error(w, string(res), statusCode)
}
