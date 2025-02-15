package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"socialai/model"
	"socialai/service"

	"github.com/pborman/uuid"

	jwt "github.com/form3tech-oss/jwt-go"
)

var (
	mediaTypes = map[string]string{ // key: extension(.jpg, .avi): value: image / video
		".jpeg": "image",
		".jpg":  "image",
		".gif":  "image",
		".png":  "image",
		".mov":  "video",
		".mp4":  "video",
		".avi":  "video",
		".flv":  "video",
		".wmv":  "video",
	}
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse from body of request to get a json object.
	fmt.Println("Received one upload request")

	// 1. Process request: input(form data) -> post struct
	token := r.Context().Value("user")
	claims := token.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"]

	p := model.Post{
		Id:      uuid.New(),
		User:    username.(string),
		Message: r.FormValue("message"),
	}
	file, header, err := r.FormFile("media_file")
	// fmt.Println(header.Filename)
	// files := r.MultipartForm.File["media_file"]
	// fmt.Println(len(files))
	// for _, fileHeader := range files {
	// 	fmt.Println(fileHeader.Filename)
	// }

	if err != nil {
		http.Error(w, "Media file is not available", http.StatusBadRequest)
		fmt.Printf("Media file is not available %v\n", err)
		return
	}
	suffix := filepath.Ext(header.Filename)
	if t, ok := mediaTypes[suffix]; ok {
		p.Type = t
	} else {
		// return error
		p.Type = "unknown"
	}

	// 2. business logic -> service
	err = service.SavePost(&p, file)
	if err != nil {
		http.Error(w, "Failed to save post to backend", http.StatusInternalServerError)
		fmt.Printf("Failed to save post to backend %v\n", err)
		return
	}

	// 3. response
	fmt.Println("Post is saved successfully.")
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one request for search")
	w.Header().Set("Content-Type", "application/json")

	// 1. Process request: URL -> string
	user := r.URL.Query().Get("user")
	keywords := r.URL.Query().Get("keywords")

	// 2. business logic -> service
	var posts []model.Post
	var err error
	if user != "" {
		posts, err = service.SearchPostsByUser(user)
	} else {
		posts, err = service.SearchPostsByKeywords(keywords)
	}

	if err != nil {
		http.Error(w, "Failed to read post from backend", http.StatusInternalServerError)
		fmt.Printf("Failed to read post from backend %v.\n", err)
		return
	}

	// 3. response: post struct => JSON string
	js, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, "Failed to parse posts into JSON format", http.StatusInternalServerError)
		fmt.Printf("Failed to parse posts into JSON format %v.\n", err)
		return
	}
	w.Write(js)
}
