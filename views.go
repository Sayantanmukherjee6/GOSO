package main

import (
	"html/template"
	"net/http"
)

func renderStart(w http.ResponseWriter, r *http.Request) {
	// Parse the template
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the template
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
	}
}

func initProcess(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the "username" field from the form data
	userName := r.FormValue("uname")

	if len(userName) != 0 {
		roomId := generateRandomID()
		http.Redirect(w, r, "/room?rid="+roomId+"&uname="+userName, http.StatusSeeOther)

	} else {
		http.Error(w, "Username can't be empty", http.StatusBadRequest)
	}
}

func renderJoin(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	roomId := r.URL.Query().Get("rid")
	// Parse the HTML template from the "templates" folder
	tmpl, err := template.ParseFiles("templates/join.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	//Pass query parameters to the template
	data := struct {
		RoomId string
	}{
		RoomId: roomId,
	}

	// Render the template with the data
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func initJoin(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the "username" field from the form data
	userName := r.FormValue("uname")
	roomId := r.FormValue("rid")

	if len(userName) != 0 {
		http.Redirect(w, r, "/room?rid="+roomId+"&uname="+userName, http.StatusSeeOther)

	} else {
		http.Error(w, "Username can't be empty", http.StatusBadRequest)
	}
}

func renderRoom(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	roomId := r.URL.Query().Get("rid")
	userName := r.URL.Query().Get("uname")

	if len(userName) == 0 {
		http.Redirect(w, r, "/join?rid="+roomId, http.StatusSeeOther)
	}

	if roomId == "" {
		http.Error(w, "No Room", http.StatusBadRequest)
		return
	}
	// Parse the HTML template from the "templates" folder
	tmpl, err := template.ParseFiles("templates/room.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	// Pass query parameters to the template
	data := struct {
		Roomlink string
		RoomId   string
		Username string
	}{
		Roomlink: r.Host + r.URL.Path + "?rid=" + roomId,
		RoomId:   roomId,
		Username: userName,
	}

	// Render the template with the data
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
