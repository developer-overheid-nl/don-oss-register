package oss_client

import (
	"sync"

	problem "github.com/developer-overheid-nl/don-oss-register/pkg/oss_client/helpers/problem"
	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
)

var installProblemErrorHookOnce sync.Once

func installProblemErrorHook() {
	installProblemErrorHookOnce.Do(func() {
		fallback := tonic.GetErrorHook()
		tonic.SetErrorHook(func(ctx *gin.Context, err error) (int, interface{}) {
			if apiErr, ok := err.(problem.ProblemJSON); ok {
				ctx.Header("Content-Type", "application/problem+json")
				return apiErr.Status, apiErr
			}
			return fallback(ctx, err)
		})
	})
}
