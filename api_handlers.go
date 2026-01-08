package main

import (
	"errors"
	"net/http"
	"slices"
)

func ProfileParamsValidator(r *http.Request) error {
	var data ProfileParams

	// required
	data.Login = r.URL.Query().Get("login")
	if data.Login == "" {
		return errors.New("login must me not empty")
	}

	return nil
}

func CreateParamsValidator(r *http.Request) error {
	var data CreateParams

	// required
	data.Login = r.URL.Query().Get("login")
	if data.Login == "" {
		return errors.New("login must me not empty")
	}

	// enum
	enums := []string{"user, moderator, admin"}
	data.Status = r.URL.Query().Get("status")
	if !slices.Contains(enums, data.Status) {
		return errors.New("status must be one of [user, moderator, admin]")
	}

	// default
	data.Status = r.URL.Query().Get("status")
	if data.Status == "" {
		data.Status = "user"
	}

	return nil
}

func MyApiUserProfile(w http.ResponseWriter, r *http.Request) {

	// some shit
}

func MyApiUserCreate(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("X-Auth")
	if auth != "100500" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	// some shit
}

func OtherCreateParamsValidator(r *http.Request) error {
	var data OtherCreateParams

	// required
	data.Username = r.URL.Query().Get("username")
	if data.Username == "" {
		return errors.New("username must me not empty")
	}

	// enum
	enums := []string{"warrior, sorcerer, rouge"}
	data.Class = r.URL.Query().Get("class")
	if !slices.Contains(enums, data.Class) {
		return errors.New("class must be one of [warrior, sorcerer, rouge]")
	}

	// default
	data.Class = r.URL.Query().Get("class")
	if data.Class == "" {
		data.Class = "warrior"
	}

	return nil
}

func OtherApiUserCreate(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("X-Auth")
	if auth != "100500" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	// some shit
}
