package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ForumEngine manages on-chain discussion threads and comments.
type ForumEngine struct {
	led StateRW
	mu  sync.RWMutex
}

var (
	forum     *ForumEngine
	forumOnce sync.Once
)

// InitForum initialises the global forum engine with a ledger backend.
func InitForum(led StateRW) { forumOnce.Do(func() { forum = &ForumEngine{led: led} }) }

// Forum returns the singleton forum engine.
func Forum() *ForumEngine { return forum }

// Thread represents a forum discussion thread.
type Thread struct {
	ID        Hash    `json:"id"`
	Creator   Address `json:"creator"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	CreatedAt int64   `json:"created_at"`
}

// Comment represents a reply to a thread.
type Comment struct {
	ID        Hash    `json:"id"`
	ThreadID  Hash    `json:"thread_id"`
	Author    Address `json:"author"`
	Body      string  `json:"body"`
	CreatedAt int64   `json:"created_at"`
}

func (f *ForumEngine) keyThread(id Hash) []byte {
	return append([]byte("forum:thread:"), id[:]...)
}

func (f *ForumEngine) keyComment(tid, cid Hash) []byte {
	hexTid := hex.EncodeToString(tid[:])
	return []byte("forum:comment:" + hexTid + ":" + hex.EncodeToString(cid[:]))
}

// CreateThread stores a new discussion thread and returns its id.
func (f *ForumEngine) CreateThread(author Address, title, body string) (Hash, error) {
	if len(title) == 0 || len(body) == 0 {
		return Hash{}, errors.New("title and body required")
	}
	t := Thread{Creator: author, Title: title, Body: body, CreatedAt: time.Now().Unix()}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%x-%d-%s", author, t.CreatedAt, title)))
	t.ID = sum
	b, err := json.Marshal(t)
	if err != nil {
		return Hash{}, err
	}
	if err := f.led.SetState(f.keyThread(t.ID), b); err != nil {
		return Hash{}, err
	}
	return t.ID, nil
}

// GetThread retrieves a thread by id.
func (f *ForumEngine) GetThread(id Hash) (*Thread, error) {
	data, err := f.led.GetState(f.keyThread(id))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("thread %x not found", id)
	}
	var t Thread
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// ListThreads returns all existing threads.
func (f *ForumEngine) ListThreads() ([]Thread, error) {
	it := f.led.PrefixIterator([]byte("forum:thread:"))
	var out []Thread
	for it.Next() {
		var t Thread
		if err := json.Unmarshal(it.Value(), &t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, it.Error()
}

// AddComment appends a comment to the given thread.
func (f *ForumEngine) AddComment(tid Hash, author Address, body string) (Hash, error) {
	if len(body) == 0 {
		return Hash{}, errors.New("comment body required")
	}
	if _, err := f.GetThread(tid); err != nil {
		return Hash{}, err
	}
	c := Comment{ThreadID: tid, Author: author, Body: body, CreatedAt: time.Now().Unix()}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%x-%d-%s", author, c.CreatedAt, body)))
	c.ID = sum
	b, err := json.Marshal(c)
	if err != nil {
		return Hash{}, err
	}
	if err := f.led.SetState(f.keyComment(tid, c.ID), b); err != nil {
		return Hash{}, err
	}
	return c.ID, nil
}

// ListComments returns all comments for a thread.
func (f *ForumEngine) ListComments(tid Hash) ([]Comment, error) {
	prefix := []byte("forum:comment:" + hex.EncodeToString(tid[:]) + ":")
	it := f.led.PrefixIterator(prefix)
	var out []Comment
	for it.Next() {
		var c Comment
		if err := json.Unmarshal(it.Value(), &c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, it.Error()
}

// Forum_CreateThread is exposed as a VM opcode.
func Forum_CreateThread(author Address, title, body string) (Hash, error) {
	if forum == nil {
		return Hash{}, errors.New("forum not initialised")
	}
	return forum.CreateThread(author, title, body)
}

// Forum_GetThread retrieves a thread by id via opcode.
func Forum_GetThread(id Hash) (*Thread, error) {
	if forum == nil {
		return nil, errors.New("forum not initialised")
	}
	return forum.GetThread(id)
}

// Forum_ListThreads lists all threads via opcode.
func Forum_ListThreads() ([]Thread, error) {
	if forum == nil {
		return nil, errors.New("forum not initialised")
	}
	return forum.ListThreads()
}

// Forum_AddComment adds a comment via opcode.
func Forum_AddComment(tid Hash, author Address, body string) (Hash, error) {
	if forum == nil {
		return Hash{}, errors.New("forum not initialised")
	}
	return forum.AddComment(tid, author, body)
}

// Forum_ListComments lists comments for a thread via opcode.
func Forum_ListComments(tid Hash) ([]Comment, error) {
	if forum == nil {
		return nil, errors.New("forum not initialised")
	}
	return forum.ListComments(tid)
}
