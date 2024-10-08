package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/YugenDev/go-platzi-CQRS/models"
	elastic "github.com/elastic/go-elasticsearch/v8"
)

type ElasticSearchRepository struct {
	Client *elastic.Client
}

func NewElastic(url string) (*ElasticSearchRepository, error) {
	client, err := elastic.NewClient(elastic.Config{
		Addresses: []string{url},
	})
	if err != nil {
		return nil, err
	}

	return &ElasticSearchRepository{
		Client: client,
	}, nil
}

func (r *ElasticSearchRepository) Close() {
	// pass
}

func (r *ElasticSearchRepository) IndexFeed(ctx context.Context, feed models.Feed) error {
	body, _ := json.Marshal(feed)
	_, err := r.Client.Index(
		"feeds",
		bytes.NewReader(body),
		r.Client.Index.WithDocumentID(feed.ID),
		r.Client.Index.WithContext(ctx),
		r.Client.Index.WithRefresh("wait_for"),
	)
	return err
}

func (r *ElasticSearchRepository) SearchFeed(ctx context.Context, query string) (results []models.Feed, err error) {
	var buf bytes.Buffer

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":            query,
				"fields":           []string{"title", "description"},
				"fuzziness":        3,
				"cutoff_frequency": 0.0001,
			},
		},
	}

	if err = json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, err
	}

	res, err := r.Client.Search(
		r.Client.Search.WithContext(ctx),
		r.Client.Search.WithIndex("feeds"),
		r.Client.Search.WithBody(&buf),
		r.Client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			results = nil
		}
	}()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var eRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&eRes); err != nil {
		return nil, err
	}

	var feeds []models.Feed

	for _, hit := range eRes["hits"].(map[string]interface{})["hits"].([]interface{}) {
		
		feed := models.Feed{}

		source := hit.(map[string]interface{})["_source"]

		marshal, err := json.Marshal(source)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(marshal, &feed); err != nil {
			return nil, err
		}

		feeds = append(feeds, feed)
	}

	return feeds, nil
}
