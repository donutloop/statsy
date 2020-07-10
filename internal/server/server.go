package server

import (
	"github.com/donutloop/statsy/internal/dao"
	"github.com/donutloop/statsy/internal/handler"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httptest"
)

// Server is the HTTP server.
type Server struct {
	mux        *chi.Mux
	logger     *logrus.Logger
	dao        *dao.DAO
	testserver *httptest.Server
	TestURL    string
}

// NewServer creates a new Server.
func New(dao *dao.DAO) *Server {
	if dao == nil {
		panic("dao is not set")
	}
	s := &Server{
		logger: logrus.New(),
		mux:    chi.NewRouter(),
		dao:    dao,
	}
	return s
}

func (s *Server) AddPostHandlerWithStats(pattern string, domainFunc handler.HandlerDomainFunc) {
	ctx := handler.Context{
		DAO:      s.dao,
		LogError: s.logger.Error,
		LogInfo:  s.logger.Info,
	}
	domainHandler := handler.StatsHandler(domainFunc, ctx)
	s.mux.Method(http.MethodPost, pattern, domainHandler)
}

func (s *Server) AddGetHandler(pattern string, domainFunc handler.HandlerDomainFunc) {
	ctx := handler.Context{
		DAO:      s.dao,
		LogError: s.logger.Error,
		LogInfo:  s.logger.Info,
	}
	domainHandler := handler.WrapHandler(domainFunc, ctx)
	s.mux.Method(http.MethodGet, pattern, domainHandler)
}

// Start starts the server.
func (s *Server) Start(addr string, test bool) error {

	if test {
		s.testserver = httptest.NewServer(s.mux)
		s.TestURL = s.testserver.URL
		return nil
	}

	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) Stop(test bool) error {
	if test {
		s.testserver.Close()
		return nil
	}
	return nil
}
