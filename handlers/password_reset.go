package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/yookibooki/auth/auth"
	"github.com/yookibooki/auth/email"
	"github.com/yookibooki/auth/repo"
)

type PwdResetHandlers struct {
	tmpls        *template.Template
	userRepo     repo.UserRepo
	pwdResetRepo repo.PwdResetTokenRepo
	pwdHasher    auth.Hasher
	authCodeMgr  *auth.AuthCodeManager
	emailSender  email.Sender
	baseURL      string
}

func NewPwdResetHandlers(
	tmpls *template.Template,
	userRepo repo.UserRepo,
	pwdResetRepo repo.PwdResetTokenRepo,
	pwdHasher auth.Hasher,
	authCodeMgr *auth.AuthCodeManager,
	emailSender email.Sender,
	baseURL string,
) *PwdResetHandlers {
	return &PwdResetHandlers{
		tmpls:        tmpls,
		userRepo:     userRepo,
		pwdResetRepo: pwdResetRepo,
		pwdHasher:    pwdHasher,
		authCodeMgr:  authCodeMgr,
		emailSender:  emailSender,
		baseURL:      baseURL,
	}
}

func (h *PwdResetHandlers) ServeReset(w http.ResponseWriter, r *http.Request) {
	h.tmpls.ExecuteTemplate(w, "reset.html", nil)
}

func (h *PwdResetHandlers) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")

	ctx := context.Background()
	user, err := h.userRepo.FindByEmail(ctx, email)
	if err != nil {
		h.tmpls.ExecuteTemplate(w, "link-sent.html", nil)
		return
	}

	tokenData, plainToken, err := h.authCodeMgr.CreatePwdResetToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	if err := h.pwdResetRepo.Create(ctx, tokenData); err != nil {
		http.Error(w, "Failed to save token", http.StatusInternalServerError)
		return
	}

	resetURL := fmt.Sprintf("%s/reset/confirm?token=%s", h.baseURL, plainToken)
	if err := h.emailSender.Send(email, "Reset your password", fmt.Sprintf("Click here to reset: %s", resetURL)); err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	h.tmpls.ExecuteTemplate(w, "link-sent.html", nil)
}

func (h *PwdResetHandlers) HandleConfirm(w http.ResponseWriter, r *http.Request) {
	tokenHash := r.URL.Query().Get("token")
	if tokenHash == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	tokenRecord, err := h.pwdResetRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	if tokenRecord.UsedAt.Valid {
		http.Error(w, "Token already used", http.StatusBadRequest)
		return
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		http.Error(w, "Token expired", http.StatusBadRequest)
		return
	}

	h.tmpls.ExecuteTemplate(w, "reset.html", struct {
		Token  string
		Action string
	}{
		Token:  tokenHash,
		Action: "complete",
	})
}

func (h *PwdResetHandlers) HandleComplete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	tokenHash := r.FormValue("token")
	password := r.FormValue("password")

	if len(password) < 8 {
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	tokenRecord, err := h.pwdResetRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	pwdHash, err := h.pwdHasher.Hash(password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	if err := h.userRepo.UpdatePassword(ctx, tokenRecord.UserID, pwdHash); err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	if err := h.pwdResetRepo.MarkUsed(ctx, tokenRecord.ID); err != nil {
		http.Error(w, "Failed to mark token as used", http.StatusInternalServerError)
		return
	}

	h.tmpls.ExecuteTemplate(w, "success.html", nil)
}
