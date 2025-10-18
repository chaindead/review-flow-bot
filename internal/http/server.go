package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/do/v2"

	"github.com/chaindead/review-flow-bot/internal/config"
)

type Server struct {
	cfg config.HTTP `do:"cfg.http"`

	srv *http.Server
}

func New(i do.Injector) (*Server, error) {
	s, err := do.InvokeStruct[*Server](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke http server")
	}

	r := gin.Default()
	s.setupRoutes(r)

	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Port),
		Handler: r,
	}

	return s, nil
}

func (s *Server) setupRoutes(r *gin.Engine) {
	r.POST("gitlab/webhook", s.webhookHandler)
}

func (s *Server) Start() {
	log.Info().Str("port", s.srv.Addr).Msg("starting http server")

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()
}

func (s *Server) Shutdown() error {
	log.Info().Msg("shutting down http server")

	err := s.srv.Shutdown(context.Background())
	if err != nil {
		return errors.Wrap(err, "http server shutdown")
	}

	return nil
}
