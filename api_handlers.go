package main


import (
	"net/http"
	"slices"
	"errors"
	"encoding/json"
)
	
func ProfileParamsValidator(r *http.Request) error {
    var data ProfileParams
	
	// required
	data.Login = r.URL.Query().Get("login")
	if data.Login == "" {
		return ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err: errors.New("login must me not empty"),
		}
	}
	
	return nil
}
	
func CreateParamsValidator(r *http.Request) error {
    var data CreateParams
	
	// required
	data.Login = r.URL.Query().Get("login")
	if data.Login == "" {
		return ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err: errors.New("login must me not empty"),
		}
	}
	
	// enum
	enums := []string{"user, moderator, admin"}
	data.Status = r.URL.Query().Get("status")
	if !slices.Contains(enums, data.Status) {
		return ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err: errors.New("status must be one of [user, moderator, admin]"),
		}
	}
	
	// default
	data.Status = r.URL.Query().Get("status")
	if data.Status == "" {
		data.Status = "user"
	}
	
	return nil
}
	
func (h *MyApi) UserProfile(w http.ResponseWriter, r *http.Request) {
	err := ProfileParamsValidator(r)

	if err != nil {
		apiError, ok := err.(ApiError)
		if ok {
			w.WriteHeader(apiError.HTTPStatus)
			jsonRaw, _ := json.Marshal(err)
			w.Write([]byte(jsonRaw))
			return
		}
	}
	
}

func (h *MyApi) UserCreate(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("X-Auth")
	if auth != "100500" {
	    http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
	err := CreateParamsValidator(r)

	if err != nil {
		apiError, ok := err.(ApiError)
		if ok {
			w.WriteHeader(apiError.HTTPStatus)
			jsonRaw, _ := json.Marshal(err)
			w.Write([]byte(jsonRaw))
			return
		}
	}
	
}

	func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    
	case "/user/profile":
		h.UserProfile(w, r)
	
	case "/user/create":
		h.UserCreate(w, r)
	
    default:
		apiError := ApiError{
			HTTPStatus: http.StatusNotFound,
			err: 		errors.New("unknown method")
		}
        w.WriteHeader(apiError.HTTPStatus)
		jsonRaw, _ := json.Marshal(apiError)
		w.Write([]byte(jsonRaw))
		return
    }
	
func OtherCreateParamsValidator(r *http.Request) error {
    var data OtherCreateParams
	
	// required
	data.Username = r.URL.Query().Get("username")
	if data.Username == "" {
		return ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err: errors.New("username must me not empty"),
		}
	}
	
	// enum
	enums := []string{"warrior, sorcerer, rouge"}
	data.Class = r.URL.Query().Get("class")
	if !slices.Contains(enums, data.Class) {
		return ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err: errors.New("class must be one of [warrior, sorcerer, rouge]"),
		}
	}
	
	// default
	data.Class = r.URL.Query().Get("class")
	if data.Class == "" {
		data.Class = "warrior"
	}
	
	return nil
}
	
func (h *OtherApi) UserCreate(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("X-Auth")
	if auth != "100500" {
	    http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
	err := CreateParamsValidator(r)

	if err != nil {
		apiError, ok := err.(ApiError)
		if ok {
			w.WriteHeader(apiError.HTTPStatus)
			jsonRaw, _ := json.Marshal(err)
			w.Write([]byte(jsonRaw))
			return
		}
	}
	
}

	func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    
	case "/user/profile":
		h.UserProfile(w, r)
	
	case "/user/create":
		h.UserCreate(w, r)
	
	case "/user/create":
		h.UserCreate(w, r)
	
    default:
		apiError := ApiError{
			HTTPStatus: http.StatusNotFound,
			err: 		errors.New("unknown method")
		}
        w.WriteHeader(apiError.HTTPStatus)
		jsonRaw, _ := json.Marshal(apiError)
		w.Write([]byte(jsonRaw))
		return
    }
	