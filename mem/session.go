package mem

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/google/uuid"
)

var SessionCache *ristretto.Cache[string, *Session]
var CurrentSessionKeyCache *ristretto.Cache[string, string]

type Session struct {
	key             string            `json:"k"`
	batch           int               `json:"b"`
	previousKey     string            `json:"pk"`
	nextKey         string            `json:"nk"`
	sessionID       string            `json:"sid"`
	messages        []*schema.Message `json:"msgs"`
	BusinessContext map[string]any    `json:"bc"`
}

func NewSession(key string, batch int, previousKey, nextKey, sessionID string, messages []*schema.Message) *Session {
	return &Session{
		key:             key,
		batch:           batch,
		previousKey:     previousKey,
		nextKey:         nextKey,
		sessionID:       sessionID,
		messages:        messages,
		BusinessContext: make(map[string]any),
	}
}

func WithNewSession() compose.NewGraphOption {
	return compose.WithGenLocalState(func(ctx context.Context) *Session {
		return NewSession("", 0, "", "", "", nil)
	})
}

func InitSessionStore() (func(), error) {
	sessionCache, err := ristretto.NewCache(&ristretto.Config[string, *Session]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}
	SessionCache = sessionCache

	sessionKeyCache, err := ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}
	CurrentSessionKeyCache = sessionKeyCache

	return func() {
		SessionCache.Close()
		CurrentSessionKeyCache.Close()
	}, nil
}

func GetCurrentSessionKey(sessionID string) (string, bool) {
	return CurrentSessionKeyCache.Get(sessionID)
}

func GetSession(key string) (*Session, bool) {
	return SessionCache.Get(key)
}

func GetCurrentSession(sessionID string) (*Session, bool) {
	key, ok := GetCurrentSessionKey(sessionID)
	if !ok {
		return nil, false
	}
	return GetSession(key)
}

func SetCurrentSessionKey(sessionID, key string) {
	CurrentSessionKeyCache.Set(sessionID, key, 1)
}

func SetSession(key string, session *Session) {
	SessionCache.Set(key, session, 1)
}

func SetCurrentSession(session *Session) {
	SetCurrentSessionKey(session.GetID(), session.GetKey())
	SetSession(session.GetKey(), session)
}

func AddMessagesToSession(sessionID string, messages ...*schema.Message) {
	sessions, ok := GetCurrentSession(sessionID)
	if !ok {
		return
	}

	sessions.messages = append(sessions.messages, messages...)

	SetCurrentSession(sessions)
}

func (s *Session) GetKey() string {
	return fmt.Sprintf("%s_%d", s.sessionID, s.batch)
}

func (s *Session) GetBatch() int {
	return s.batch
}

func (s *Session) GetPreviousKey() string {
	return s.previousKey
}

func (s *Session) GetNextKey() string {
	return s.nextKey
}

func (s *Session) GetID() string {
	return s.sessionID
}

func (s *Session) GetMessages() []*schema.Message {
	return s.messages
}

func (s *Session) SetMessages(messages []*schema.Message) {
	s.messages = messages
	SetCurrentSession(s)
}

func (s *Session) AppendMessages(messages ...*schema.Message) {
	s.messages = append(s.messages, messages...)
	SetCurrentSession(s)
}

func (s *Session) Load(sessionID string) bool {
	session, ok := GetCurrentSession(sessionID)

	if !ok {
		return false
	}

	s.key = session.GetKey()
	s.batch = session.GetBatch()
	s.previousKey = session.GetPreviousKey()
	s.nextKey = session.GetNextKey()
	s.sessionID = session.GetID()
	s.messages = session.GetMessages()

	return true
}

func (s *Session) InitNewAndSave(sessionID string) {
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	s.sessionID = sessionID
	s.batch = 1
	s.previousKey = ""
	s.nextKey = ""
	s.messages = make([]*schema.Message, 0)

	SetCurrentSession(s)
}

func (s *Session) Init(sessionID string) {
	ok := s.Load(sessionID)

	if !ok {
		s.InitNewAndSave(sessionID)
	}
}
