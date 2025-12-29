package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/yookibooki/auth/auth"
	"github.com/yookibooki/auth/email"
	"github.com/yookibooki/auth/repo"
)

type AuthHandlers struct {
	tmpls           *template.Template
	userRepo        repo.UserRepo
	authCodeRepo    repo.AuthCodeRepo
	pwdHasher       auth.Hasher
	authCodeManager *auth.AuthCodeManager
	emailValidator  auth.EmailValidator
	emailSender     email.Sender
	baseURL         string
}

func NewAuthHandlers(
	tmpls *template.Template,
	userRepo repo.UserRepo,
	authCodeRepo repo.AuthCodeRepo,
	pwdHasher auth.Hasher,
	authCodeManager *auth.AuthCodeManager,
	emailValidator auth.EmailValidator,
	emailSender email.Sender,
	baseURL string,
) *AuthHandlers {
	return &AuthHandlers{
		tmpls:           tmpls,
		userRepo:        userRepo,
		authCodeRepo:    authCodeRepo,
		pwdHasher:       pwdHasher,
		authCodeManager: authCodeManager,
		emailValidator:  emailValidator,
		emailSender:     emailSender,
		baseURL:         baseURL,
	}
}

type AuthPageData struct {
	Step            string
	Email           string
	Error           string
	PostEmailURL    string
	PostPasswordURL string
}

func (h *AuthHandlers) ServeAuth(w http.ResponseWriter, r *http.Request) {
	redirectURI := r.URL.Query().Get("redirect_uri")
	clientID := r.URL.Query().Get("client_id")
	state := r.URL.Query().Get("state")

	if redirectURI == "" || clientID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	data := AuthPageData{
		Step:            "email",
		PostEmailURL:    "/auth/email?redirect_uri=" + url.QueryEscape(redirectURI) + "&client_id=" + url.QueryEscape(clientID) + "&state=" + url.QueryEscape(state),
		PostPasswordURL: "/auth/password",
	}

	h.tmpls.ExecuteTemplate(w, "auth.html", data)
}

func (h *AuthHandlers) HandleEmail(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	redirectURI := r.URL.Query().Get("redirect_uri")
	clientID := r.URL.Query().Get("client_id")
	state := r.URL.Query().Get("state")

	if !h.emailValidator.Validate(email) {
		h.renderAuthError(w, "Invalid email address")
		return
	}

	ctx := context.Background()
	_, err := h.userRepo.FindByEmail(ctx, email)

	data := AuthPageData{
		Email:           email,
		PostPasswordURL: fmt.Sprintf("/auth/password?redirect_uri=%s&client_id=%s&state=%s", url.QueryEscape(redirectURI), url.QueryEscape(clientID), url.QueryEscape(state)),
	}

	if err != nil {
		data.Step = "signup"
		h.tmpls.ExecuteTemplate(w, "auth.html", data)
		return
	}

	data.Step = "password"
	h.tmpls.ExecuteTemplate(w, "auth.html", data)
}

func (h *AuthHandlers) HandlePassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	redirectURI := r.URL.Query().Get("redirect_uri")
	clientID := r.URL.Query().Get("client_id")
	state := r.URL.Query().Get("state")

	if len(password) < 8 {
		h.renderAuthError(w, "Password must be at least 8 characters")
		return
	}

	ctx := context.Background()
	user, err := h.userRepo.FindByEmail(ctx, email)

	if err == nil {
		if !h.pwdHasher.Compare(user.PwdHash, password) {
			h.renderAuthError(w, "Invalid password")
			return
		}

		authCode, code, err := h.authCodeManager.CreateAuthCode(user.ID, clientID, redirectURI, state)
		if err != nil {
			http.Error(w, "Failed to create auth code", http.StatusInternalServerError)
			return
		}

		if err := h.authCodeRepo.Create(ctx, authCode); err != nil {
			http.Error(w, "Failed to save auth code", http.StatusInternalServerError)
			return
		}

		confirmURL := fmt.Sprintf("%s/auth/confirm?code=%s", h.baseURL, code)
		if err := h.sendEmail(email, "Login to your account", fmt.Sprintf("Click here to log in: %s", confirmURL)); err != nil {
			http.Error(w, "Failed to send email", http.StatusInternalServerError)
			return
		}

		h.tmpls.ExecuteTemplate(w, "link-sent.html", nil)
		return
	}

	pwdHash, err := h.pwdHasher.Hash(password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	newUser, err := h.userRepo.Create(ctx, email, pwdHash)
	if err != nil {
		h.renderAuthError(w, "Failed to create account")
		return
	}

	authCode, code, err := h.authCodeManager.CreateAuthCode(newUser.ID, clientID, redirectURI, state)
	if err != nil {
		http.Error(w, "Failed to create auth code", http.StatusInternalServerError)
		return
	}

	if err := h.authCodeRepo.Create(ctx, authCode); err != nil {
		http.Error(w, "Failed to save auth code", http.StatusInternalServerError)
		return
	}

	confirmURL := fmt.Sprintf("%s/auth/confirm?code=%s", h.baseURL, code)
	if err := h.sendEmail(email, "Confirm your email", fmt.Sprintf("Click here to confirm: %s", confirmURL)); err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	h.tmpls.ExecuteTemplate(w, "link-sent.html", nil)
}

func (h *AuthHandlers) HandleConfirm(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code", http.StatusBadRequest)
		return
	}

	codeBytes, err := base64.URLEncoding.DecodeString(code)
	if err != nil {
		http.Error(w, "Invalid code", http.StatusBadRequest)
		return
	}

	codeHash, err := h.pwdHasher.Hash(string(codeBytes))
	if err != nil {
		http.Error(w, "Invalid code", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	codeRecord, err := h.authCodeRepo.FindByCodeHash(ctx, codeHash)
	if err != nil {
		http.Error(w, "Invalid or expired code", http.StatusBadRequest)
		return
	}

	if codeRecord.UsedAt.Valid {
		http.Error(w, "Code already used", http.StatusBadRequest)
		return
	}

	if time.Now().After(codeRecord.ExpiresAt) {
		http.Error(w, "Code expired", http.StatusBadRequest)
		return
	}

	if err := h.authCodeRepo.MarkUsed(ctx, codeRecord.ID); err != nil {
		http.Error(w, "Failed to mark code as used", http.StatusInternalServerError)
		return
	}

	h.tmpls.ExecuteTemplate(w, "success.html", nil)
}

func (h *AuthHandlers) renderAuthError(w http.ResponseWriter, errMsg string) {
	data := AuthPageData{
		Step:  "email",
		Error: errMsg,
	}
	h.tmpls.ExecuteTemplate(w, "auth.html", data)
}

func (h *AuthHandlers) sendEmail(to, subject, body string) error {
	return h.emailSender.Send(to, subject, body)
}
