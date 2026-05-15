package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPServer HTTP 服务器包装
type HTTPServer struct {
	Engine       *gin.Engine
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// ListenAndServe 启动 HTTP 服务器
func (s *HTTPServer) ListenAndServe() error {
	server := &http.Server{
		Addr:         s.Addr,
		Handler:      s.Engine,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		IdleTimeout:  s.IdleTimeout,
	}
	return server.ListenAndServe()
}
