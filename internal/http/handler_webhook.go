package http

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"

	_ "github.com/chaindead/review-flow-bot/internal/logger"
)

func (s *Server) webhookHandler(c *gin.Context) {
	secret := c.GetHeader("X-Gitlab-Token")
	if secret != s.cfg.Secret {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	bodyBytes, _ := io.ReadAll(c.Request.Body)

	var err error
	switch t := gjson.GetBytes(bodyBytes, "event_type").String(); t {
	case "note":
		err = s.onWebhookNote(bodyBytes)
	case "merge_request":
		err = s.onWebhookMergeRequest(bodyBytes)
	default:
		log.Warn().RawJSON("payload", bodyBytes).Msg("unknown event")
	}
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) onWebhookMergeRequest(data []byte) error {
	etype := gjson.GetBytes(data, "event_type").String()
	kind := gjson.GetBytes(data, "object_kind").String()
	action := gjson.GetBytes(data, "object_attributes.action").String()
	pid := gjson.GetBytes(data, "object_attributes.source_project_id").String()
	mid := gjson.GetBytes(data, "object_attributes.iid").String()
	who := gjson.GetBytes(data, "user.username").String()
	title := gjson.GetBytes(data, "object_attributes.title").String()
	url := gjson.GetBytes(data, "object_attributes.url").String()
	resolved := gjson.GetBytes(data, "object_attributes.blocking_discussions_resolved").String()

	log.Info().
		Str("pid", pid).
		Str("mid", mid).
		Str("who", who).
		Str("type", etype).
		Str("kind", kind).
		Str("action", action).
		Str("url", url).
		Str("title", title).
		Str("resolved", resolved).
		Msg("new webhook event")

	//actions: unapproved,approved,

	return nil
}

func (s *Server) onWebhookNote(data []byte) error {
	path := gjson.GetBytes(data, "object_attributes.position.new_path").String()
	if path == "" {
		path = gjson.GetBytes(data, "object_attributes.position.old_path").String()
	}
	resolvedAt := gjson.GetBytes(data, "object_attributes.resolved_at").String()
	noteType := gjson.GetBytes(data, "object_attributes.type").String()

	log.Info().
		Str("pid", gjson.GetBytes(data, "project_id").String()).
		Str("mid", gjson.GetBytes(data, "merge_request.iid").String()).
		Str("who", gjson.GetBytes(data, "user.username").String()).
		Str("type", gjson.GetBytes(data, "event_type").String()).
		Str("action", gjson.GetBytes(data, "object_attributes.action").String()).
		Str("kind", gjson.GetBytes(data, "object_kind").String()).
		Str("path", path).
		Str("resolved", resolvedAt).
		Str("note_type", noteType).
		Str("note", gjson.GetBytes(data, "object_attributes.note").String()).
		Str("url", gjson.GetBytes(data, "object_attributes.url").String()).
		Msg("new note event")

	return nil
}
