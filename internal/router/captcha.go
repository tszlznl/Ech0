package router

import (
	"net/http"
	"sync"

	"github.com/lin-snow/ech0/internal/captcha"
)

var (
	commentCaptchaHandlerOnce sync.Once
	commentCaptchaHandler     http.Handler
	commentCaptchaHandlerErr  error
)

func commentCaptchaHTTPHandler() (http.Handler, error) {
	commentCaptchaHandlerOnce.Do(func() {
		engine, err := captcha.NewEngine()
		if err != nil {
			commentCaptchaHandlerErr = err
			return
		}
		commentCaptchaHandler = http.StripPrefix("/api", engine.Handler())
	})
	return commentCaptchaHandler, commentCaptchaHandlerErr
}
