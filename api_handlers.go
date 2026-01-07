package main

import "net/http"

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

func OtherApiUserCreate(w http.ResponseWriter, r *http.Request) {

	auth := r.Header.Get("X-Auth")
	if auth != "100500" {
	    http.Error(w, "unauthorized", http.StatusUnauthorized)
	}

	// some shit
}
