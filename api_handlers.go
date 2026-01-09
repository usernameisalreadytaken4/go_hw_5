package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
)

func ProfileParamsValidator(r *http.Request) (ProfileParams, error) {
	var data ProfileParams

	// required
	data.Login = r.URL.Query().Get("login")
	if data.Login == "" {
		return data, &ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err:        errors.New("login must me not empty"),
		}
	}

	return data, nil
}

func CreateParamsValidator(r *http.Request) (CreateParams, error) {
	var data CreateParams

	// required
	data.Login = r.URL.Query().Get("login")
	if data.Login == "" {
		return data, &ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err:        errors.New("login must me not empty"),
		}
	}

	// default
	data.Status = r.URL.Query().Get("status")
	if data.Status == "" {
		data.Status = "user"
	}

	// enum
	enums := []string{"user, moderator, admin"}
	data.Status = r.URL.Query().Get("status")
	if !slices.Contains(enums, data.Status) {
		return data, &ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err:        errors.New("status must be one of [user, moderator, admin]"),
		}
	}

	return data, nil
}

func (h *MyApi) UserProfile(w http.ResponseWriter, r *http.Request) {

	resp := map[string]interface{}{
		"error": "unknown method",
	}
	in, err := ProfileParamsValidator(r)
	if err != nil {
		resp["error"] = err.Error()
		if apiErr, ok := err.(ApiError); ok {
			w.WriteHeader(apiErr.HTTPStatus)
		} else {
			resp["error"] = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}
		jsonRaw, _ := json.Marshal(resp)
		w.Write([]byte(jsonRaw))
		return
	}

	ctx := r.Context()
	data, err := h.Profile(ctx, in)
	if err != nil {
		resp["error"] = err.Error()
		if apiErr, ok := err.(ApiError); ok {
			w.WriteHeader(apiErr.HTTPStatus)
		} else {
			resp["error"] = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}
		jsonRaw, _ := json.Marshal(resp)
		w.Write([]byte(jsonRaw))
		return
	}
	resp["response"] = data

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return

}

func (h *MyApi) UserCreate(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("X-Auth")
	if auth != "100500" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	resp := map[string]interface{}{
		"error": "unknown method",
	}
	in, err := CreateParamsValidator(r)
	if err != nil {
		resp["error"] = err.Error()
		if apiErr, ok := err.(ApiError); ok {
			w.WriteHeader(apiErr.HTTPStatus)
		} else {
			resp["error"] = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}
		jsonRaw, _ := json.Marshal(resp)
		w.Write([]byte(jsonRaw))
		return
	}

	ctx := r.Context()
	data, err := h.Create(ctx, in)
	if err != nil {
		resp["error"] = err.Error()
		if apiErr, ok := err.(ApiError); ok {
			w.WriteHeader(apiErr.HTTPStatus)
		} else {
			resp["error"] = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}
		jsonRaw, _ := json.Marshal(resp)
		w.Write([]byte(jsonRaw))
		return
	}
	resp["response"] = data

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return

}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/user/profile":
		h.UserProfile(w, r)

	case "/user/create":
		h.UserCreate(w, r)

	default:
		resp := map[string]interface{}{
			"error": "unknown method",
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
}

func OtherCreateParamsValidator(r *http.Request) (OtherCreateParams, error) {
	var data OtherCreateParams

	// required
	data.Username = r.URL.Query().Get("username")
	if data.Username == "" {
		return data, &ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err:        errors.New("username must me not empty"),
		}
	}

	// default
	data.Class = r.URL.Query().Get("class")
	if data.Class == "" {
		data.Class = "warrior"
	}

	// enum
	enums := []string{"warrior, sorcerer, rouge"}
	data.Class = r.URL.Query().Get("class")
	if !slices.Contains(enums, data.Class) {
		return data, &ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err:        errors.New("class must be one of [warrior, sorcerer, rouge]"),
		}
	}

	return data, nil
}

func (h *OtherApi) UserCreate(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("X-Auth")
	if auth != "100500" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	resp := map[string]interface{}{
		"error": "unknown method",
	}
	in, err := OtherCreateParamsValidator(r)
	if err != nil {
		resp["error"] = err.Error()
		if apiErr, ok := err.(ApiError); ok {
			w.WriteHeader(apiErr.HTTPStatus)
		} else {
			resp["error"] = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}
		jsonRaw, _ := json.Marshal(resp)
		w.Write([]byte(jsonRaw))
		return
	}

	ctx := r.Context()
	data, err := h.Create(ctx, in)
	if err != nil {
		resp["error"] = err.Error()
		if apiErr, ok := err.(ApiError); ok {
			w.WriteHeader(apiErr.HTTPStatus)
		} else {
			resp["error"] = err.Error()
			w.WriteHeader(http.StatusBadRequest)
		}
		jsonRaw, _ := json.Marshal(resp)
		w.Write([]byte(jsonRaw))
		return
	}
	resp["response"] = data

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return

}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/user/create":
		h.UserCreate(w, r)

	default:
		resp := map[string]interface{}{
			"error": "unknown method",
		}
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
}
