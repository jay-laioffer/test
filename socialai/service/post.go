package service

import (
	"mime/multipart"
	"reflect"
	"socialai/backend"
	"socialai/constants"
	"socialai/model"

	"github.com/olivere/elastic/v7"
)

func SearchPostsByUser(user string) ([]model.Post, error) {
	// 1. prepare request for backend
	query := elastic.NewTermQuery("user", user)

	// 2. call backend
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.POST_INDEX)
	if err != nil {
		return nil, err
	}

	// 3. response
	return getPostFromSearchResult(searchResult), nil
}

func SearchPostsByKeywords(keywords string) ([]model.Post, error) {
	// 1. prepare request for backend
	query := elastic.NewMatchQuery("message", keywords)
	query.Operator("AND")
	if keywords == "" {
		query.ZeroTermsQuery("all")
	}

	// 2. call backend
	searchResult, err := backend.ESBackend.ReadFromES(query, constants.POST_INDEX)
	if err != nil {
		return nil, err
	}

	// 3. response
	return getPostFromSearchResult(searchResult), nil
}

func getPostFromSearchResult(searchResult *elastic.SearchResult) []model.Post {
	var ptype model.Post
	var posts []model.Post

	for _, item := range searchResult.Each(reflect.TypeOf(ptype)) {
		p := item.(model.Post)
		posts = append(posts, p)
	}
	return posts
}

func SavePost(post *model.Post, file multipart.File) error {
	// 1. Save to GCS
	medialink, err := backend.GCSBackend.SaveToGCS(file, post.Id)
	if err != nil {
		return err
	}
	post.Url = medialink

	// 2. Save to ES + 3. response
	err = backend.ESBackend.SaveToES(post, constants.POST_INDEX, post.Id)
	if err != nil {
		// retry 3 times

		// rollback: delete from GCS
	}

	return err
}
