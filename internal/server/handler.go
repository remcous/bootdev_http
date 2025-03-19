package server

import (
	"github.com/remcous/bootdev_http/internal/request"
	"github.com/remcous/bootdev_http/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)
