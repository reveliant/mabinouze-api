package api

import (
	"encoding/base64"
	"strings"
	"net/http"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

const AuthCredentials = "user_pair"
const AuthBearer = "user_bearer"

type Credentials struct {
	User		string
	Password	string
}

const expectedHost = "" //localhost:8080"

func BearerTokenRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ignore for OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		token, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !found {
			c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\"")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		decoded, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil {
			c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\", error=\"invalid_token\", error_description=\"Invalid base64 encoding\"")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(AuthBearer, string(decoded))
		c.Next()
	}
}

func TipplerAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ignore for OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		token, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !found {
			c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\"")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		b64username, b64password, found := strings.Cut(token, ".")
		if !found {
			// Try organizer password
			decoded, err := base64.RawURLEncoding.DecodeString(token)
			if err != nil {
				// That still a no, stop here with a failure
				c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\", error=\"invalid_token\", error_description=\"Invalid token format\"")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			// That might be an orgzanizer password
			c.Set(AuthBearer, string(decoded))
			return
		}

		username, err := base64.RawURLEncoding.DecodeString(b64username)
		if err != nil {
			c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\", error=\"invalid_token\", error_description=\"Invalid base64 encoding for username\"")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		
		password, err := base64.RawURLEncoding.DecodeString(b64password)
		if err != nil {
			c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\", error=\"invalid_token\", error_description=\"Invalid base64 encoding for password\"")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(AuthCredentials, Credentials{string(username), string(password)})
		c.Next()
	}
}

func AdminAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !found {
			c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\"")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		decoded, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil {
			c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\", error=\"invalid_token\", error_description=\"Invalid base64 encoding\"")
			c.Header("X-Debug", err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(AuthBearer, string(decoded))
		c.Next()
	}
}

func AccessTokenOptional() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !found {
			return
		}

		decoded, err := base64.RawURLEncoding.DecodeString(token)
		if err != nil {
			c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\", error=\"invalid_token\", error_description=\"Invalid base64 encoding\"")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(AuthBearer, string(decoded))
		c.Next()
	}
}

func ParseUUID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("uuid"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		}
		c.Set("uuid", id)
		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
  	return func(c *gin.Context) {
		if expectedHost != "" && c.Request.Host != expectedHost {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid host header"})
			return
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Content-Security-Policy", "default-src 'none'")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(), magnetometer=(),gyroscope=(),fullscreen=(),payment=()")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		//c.Header("X-Frame-Options", "DENY")
		//c.Header("X-XSS-Protection", "1; mode=block")
		//c.Header("X-Content-Type-Options", "nosniff")
		
		/*
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Length, Content-Type, Authorization")
			c.AbortWithStatus(http.StatusNoContent)
		}
		*/

		c.Next()
	}
}

func Preflight(allowedMethods string) gin.HandlerFunc {
  	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Methods", allowedMethods)
		c.Header("Access-Control-Allow-Headers", "Content-Length, Content-Type, Authorization")
		c.AbortWithStatus(http.StatusNoContent)
		c.Next()
	}
}