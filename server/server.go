package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

const html = "<!DOCTYPE html>\n<html lang=\"ru\">\n<head>\n    <meta charset=\"UTF-8\">\n    <title>Knopka</title>\n    <script src=\"https://ajax.googleapis.com/ajax/libs/jquery/2.2.4/jquery.min.js\"></script>\n    <script src=\"/telegramform.js\"></script>\n    <style>\n        body{\n            background-color: black;\n        }\n        .button {\n            width: 100%; display: inline-block;\n            border-radius: 4px;\n            background-color: #f4511e;\n            border: none;\n            color: #FFFFFF;\n            text-align: center;\n            font-size: 28px;\n            padding: 20px;\n          transition: all 0.5s;\n       margin: 5px 0;     cursor: pointer;\n        }\n\n        .button span {\n            cursor: pointer;\n            display: inline-block;\n            position: relative;\n            transition: 0.5s;\n        }\n\n        .button span:after {\n            content: '\\00bb';\n            position: absolute;\n            opacity: 0;\n            top: 0;\n            right: -20px;\n            transition: 0.5s;\n        }\n\n        .button:hover span {\n            padding-right: 25px;\n        }\n\n        .button:hover span:after {\n            opacity: 1;\n            right: 0;\n        }\n        .token{\n            position: center;\n       }\n     .tokenform {    width: 50vw;\n    margin: 40vh auto; }  .tokenform div { max-height: 20vh; } .tokeninp{\n            position: center;\n            width: -webkit-fill-available;\n        }\n    </style>\n    <script>\n\n    </script>\n</head>\n<body>\n<div class=\"token\">\n    <form class=\"tokenform\" method=\"GET\" action=\"http://localhost:8080/analyze\">\n        <div>\n            <input class=\"tokeninp\" type=\"text\" placeholder=\"Enter ethereum address\" name=\"address\">\n            <button class=\"button\" style=\"vertical-align:middle\" type=\"submit\"><span>Analyze address</span></button>\n        </div>\n    </form>\n</div>\n</body>\n</html>"

type Server struct {
	Router *gin.Engine

	addr string
}

func New(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (s *Server) Run(ctx context.Context) error {
	s.Router = gin.Default()

	clusterController := NewController()
	clusterController.GroupWithCtx(ctx)(s.Router)

	s.Router.GET("/", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	err := s.Router.Run(s.addr)
	if err != nil {
		return err
	}

	return nil
}
