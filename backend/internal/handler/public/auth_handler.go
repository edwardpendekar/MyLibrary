// Package public exposes the read-mostly, RBAC-open endpoints: auth, books,
// chapters/verses, search, favorites/bookmarks/notes. Package admin (sibling)
// exposes the write-heavy, role-gated management endpoints.
package public

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"bookreader/backend/internal/config"
	"bookreader/backend/internal/dto"
	"bookreader/backend/internal/middleware"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/apperror"
	"bookreader/backend/pkg/csrf"
	"bookreader/backend/pkg/httpx"
	"bookreader/backend/pkg/response"
)

type AuthHandler struct {
	auth *service.AuthService
	cfg  *config.Config
}

func NewAuthHandler(auth *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, cfg: cfg}
}

// Register godoc
// @Summary      Register a new reader account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RegisterRequest true "Registration payload"
// @Success      201 {object} response.Envelope{data=dto.UserResponse}
// @Failure      422 {object} response.Envelope
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	user, err := h.auth.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, dto.ToUserResponse(user))
}

// Login godoc
// @Summary      Log in and receive httpOnly session cookies
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LoginRequest true "Credentials"
// @Success      200 {object} response.Envelope{data=dto.AuthResponse}
// @Failure      401 {object} response.Envelope
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !httpx.BindJSON(c, &req) {
		return
	}

	tokens, err := h.auth.Login(c.Request.Context(), req.Email, req.Password, requestMeta(c))
	if err != nil {
		response.Fail(c, err)
		return
	}

	h.setAuthCookies(c, tokens)
	response.OK(c, dto.AuthResponse{
		User:                 dto.ToUserResponse(tokens.User),
		AccessTokenExpiresAt: tokens.AccessTokenExpiresAt,
	})
}

// Refresh godoc
// @Summary      Rotate the refresh token and issue a new access token
// @Tags         auth
// @Produce      json
// @Success      200 {object} response.Envelope{data=dto.AuthResponse}
// @Failure      401 {object} response.Envelope
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	raw, err := c.Cookie("refresh_token")
	if err != nil || raw == "" {
		response.Fail(c, apperror.Unauthorized("no refresh token present"))
		return
	}

	tokens, err := h.auth.Refresh(c.Request.Context(), raw, requestMeta(c))
	if err != nil {
		h.clearAuthCookies(c)
		response.Fail(c, err)
		return
	}

	h.setAuthCookies(c, tokens)
	response.OK(c, dto.AuthResponse{
		User:                 dto.ToUserResponse(tokens.User),
		AccessTokenExpiresAt: tokens.AccessTokenExpiresAt,
	})
}

// ForgotPassword godoc
// @Summary      Request a password reset link
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.ForgotPasswordRequest true "Account email"
// @Success      204
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	if err := h.auth.RequestPasswordReset(c.Request.Context(), req.Email); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// ResetPassword godoc
// @Summary      Reset a password using a token from the forgot-password email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dto.ResetPasswordRequest true "Token and new password"
// @Success      204
// @Failure      401 {object} response.Envelope
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if !httpx.BindJSON(c, &req) {
		return
	}
	if err := h.auth.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

// Logout godoc
// @Summary      Revoke the current refresh token and clear session cookies
// @Tags         auth
// @Produce      json
// @Success      204
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	if raw, err := c.Cookie("refresh_token"); err == nil && raw != "" {
		_ = h.auth.Logout(c.Request.Context(), raw)
	}
	h.clearAuthCookies(c)
	response.NoContent(c)
}

func (h *AuthHandler) setAuthCookies(c *gin.Context, tokens *service.AuthTokens) {
	secure := h.cfg.JWT.CookieSecure
	sameSite := sameSiteFromString(h.cfg.JWT.CookieSameSite)
	c.SetSameSite(sameSite)

	c.SetCookie(middleware.AccessTokenCookie, tokens.AccessToken,
		int(time.Until(tokens.AccessTokenExpiresAt).Seconds()), "/", h.cfg.JWT.CookieDomain, secure, true)
	c.SetCookie("refresh_token", tokens.RefreshToken,
		int(time.Until(tokens.RefreshTokenExpiresAt).Seconds()), "/api/v1/auth", h.cfg.JWT.CookieDomain, secure, true)

	// Deliberately httpOnly=false: the frontend must be able to read this and
	// echo it back as the X-CSRF-Token header (see pkg/csrf and middleware.CSRF).
	if token, err := csrf.GenerateToken(); err == nil {
		c.SetCookie(csrf.CookieName, token,
			int(time.Until(tokens.AccessTokenExpiresAt).Seconds()), "/", h.cfg.JWT.CookieDomain, secure, false)
	}
}

func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	secure := h.cfg.JWT.CookieSecure
	c.SetCookie(middleware.AccessTokenCookie, "", -1, "/", h.cfg.JWT.CookieDomain, secure, true)
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth", h.cfg.JWT.CookieDomain, secure, true)
	c.SetCookie(csrf.CookieName, "", -1, "/", h.cfg.JWT.CookieDomain, secure, false)
}

func requestMeta(c *gin.Context) service.RequestMeta {
	return service.RequestMeta{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
}

func sameSiteFromString(v string) http.SameSite {
	switch v {
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}
