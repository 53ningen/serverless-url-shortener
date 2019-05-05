package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"time"

	"github.com/guregu/dynamo"
)

var (
	// ErrURLID cannnot create URLID
	ErrURLID = errors.New("cannot create URLID")
)

// CreateURLRequestBody CreateURL request body
type CreateURLRequestBody struct {
	URL string `dynamo:"url"`
}

// URLMapping represent id:url mapping table record
type URLMapping struct {
	URLId     string `dynamo:"url_id"`
	URL       string `dynamo:"url"`
	ExpiredAt *int64 `dynamo:"expired_at"`
}

// URLMappingResult result of URLMapper.CreateMapping
type URLMappingResult struct {
	URLString       string  `json:"url"`
	MappedURLString string  `json:"mapped_url"`
	ExpiredAt       *string `json:"expired_at"`
}

// URLMappingStore url mapping store
type URLMappingStore interface {
	GetURLMapping(urlID string) (mapping *URLMapping, e error)
	PutURLMapping(mapping *URLMapping) error
}

// URLMapper create alias URL for given URL
type URLMapper struct {
	HostName     string
	MappingStore URLMappingStore
}

// CreateMapping create alias URL for given URL
func (m *URLMapper) CreateMapping(u *url.URL, now time.Time, ttl *int) (result *URLMappingResult, e error) {
	id, e := m.getAvailableURLID(u)
	if e != nil {
		return nil, e
	}

	var expireAt *time.Time
	if ttl != nil {
		t := now.Add(time.Duration(*ttl) * time.Second)
		expireAt = &t
	}

	mapping := &URLMapping{}
	mapping.URLId = id
	mapping.URL = (*u).String()
	timestamp := expireAt.Unix()
	mapping.ExpiredAt = &timestamp
	e = m.MappingStore.PutURLMapping(mapping)
	if e != nil {
		return nil, e
	}

	result = &URLMappingResult{}
	mappedURL := &url.URL{
		Scheme: "https",
		Host:   m.HostName,
		Path:   id,
	}
	result.URLString = u.String()
	result.MappedURLString = mappedURL.String()
	if expireAt != nil {
		expiredAt := expireAt.Format(time.RFC3339)
		result.ExpiredAt = &expiredAt
	}
	return result, nil
}

func (m *URLMapper) getAvailableURLID(originalURL *url.URL) (urlID string, e error) {
	md5 := getSHA256(originalURL.String())
	for i := 4; i < 10; i++ {
		next := md5[:i]
		if mapping, e := m.MappingStore.GetURLMapping(next); e != nil {
			return "", e
		} else if mapping == nil || mapping.URL == (*originalURL).String() {
			return next, nil
		}
		continue
	}
	return "", ErrURLID
}

func getSHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// DDBMappingStore mapping store with DynamoDB
type DDBMappingStore struct {
	Cache *(map[string]*URLMapping)
	Table *dynamo.Table
}

// GetURLMapping get url mapping
func (m DDBMappingStore) GetURLMapping(urlID string) (mapping *URLMapping, e error) {
	// check local caches first
	c := *m.Cache
	if val, ok := c[urlID]; ok {
		return val, nil
	}

	if err := m.Table.Get("url_id", urlID).One(&mapping); err != nil {
		if err == dynamo.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return mapping, nil
}

// PutURLMapping put url mapping
func (m DDBMappingStore) PutURLMapping(mapping *URLMapping) error {
	e := m.Table.Put(mapping).Run()
	if e != nil {
		return e
	}
	c := *m.Cache
	c[mapping.URLId] = mapping
	return nil
}
