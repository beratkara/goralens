package goralens

import "time"

type EntryType string

const (
	TypeRequest       EntryType = "request"
	TypeCommand       EntryType = "command"
	TypeSchedule      EntryType = "schedule"
	TypeJob           EntryType = "job"
	TypeException     EntryType = "exception"
	TypeLog           EntryType = "log"
	TypeDump          EntryType = "dump"
	TypeQuery         EntryType = "query"
	TypeModel         EntryType = "model"
	TypeEvent         EntryType = "event"
	TypeMail          EntryType = "mail"
	TypeCache         EntryType = "cache"
	TypeRedis         EntryType = "redis"
	TypeGate          EntryType = "gate"
	TypeClientRequest EntryType = "client_request"
)

var EntryTypes = []EntryType{
	TypeRequest, TypeCommand, TypeSchedule, TypeJob, TypeException, TypeLog, TypeDump, TypeQuery,
	TypeModel, TypeEvent, TypeMail, TypeCache, TypeRedis, TypeGate, TypeClientRequest,
}

type Entry struct {
	Sequence   int64          `json:"sequence"`
	UUID       string         `json:"uuid"`
	BatchID    string         `json:"batch_id"`
	FamilyHash string         `json:"family_hash,omitempty"`
	Type       EntryType      `json:"type"`
	Content    map[string]any `json:"content"`
	Tags       []string       `json:"tags"`
	CreatedAt  time.Time      `json:"created_at"`

	DisplayOnIndex bool `json:"-"`
}

func (e *Entry) AddTags(tags ...string) {
	for _, tag := range tags {
		if tag == "" || e.HasTag(tag) {
			continue
		}
		e.Tags = append(e.Tags, tag)
	}
}

func (e *Entry) HasTag(tag string) bool {
	for _, t := range e.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func newEntry(t EntryType, content map[string]any) *Entry {
	return &Entry{Type: t, Content: content, DisplayOnIndex: true}
}
