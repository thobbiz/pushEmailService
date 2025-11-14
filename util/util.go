package util

import (
	"log"

	"github.com/gin-gonic/gin"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func ErrorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
