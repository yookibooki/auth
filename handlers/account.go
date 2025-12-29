package handlers

import (
	"context"
	"html/template"
	"net/http"

	"github.com/yookibooki/auth/auth"
	"github.com/yookibooki/auth/middleware"
	"github.com/yookibooki/auth/repo"
)

type AccountHandlers struct {
	tmpls          *template.Template
	userRepo       repo.UserRepo
	pwdHasher      auth.Hasher
	emailValidator auth.EmailValidator
}

func NewAccountHandlers(
	tmpls *template.Template,
	userRepo repo.UserRepo,
	pwdHasher auth.Hasher,
	emailValidator auth.EmailValidator,
) *AccountHandlers {
	return &AccountHandlers{
		tmpls:          tmpls,
		userRepo:       userRepo,
		pwdHasher:      pwdHasher,
		emailValidator: emailValidator,
	}
}

type AccountPageData struct {
	Message           string
	Error             string
	ChangeEmailURL    string
	ChangePasswordURL string
	DeleteAccountURL  string
}

func (h *AccountHandlers) ServeAccount(w http.ResponseWriter, r *http.Request) {
	data := AccountPageData{
		ChangeEmailURL:    "/account/email",
		ChangePasswordURL: "/account/password",
		DeleteAccountURL:  "/account/delete",
	}
	h.tmpls.ExecuteTemplate(w, "account.html", data)
}

func (h *AccountHandlers) HandleChangeEmail(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")

	if !h.emailValidator.Validate(email) {
		h.renderAccountError(w, "Invalid email address")
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(int)

	if err := h.userRepo.UpdateEmail(ctx, userID, email); err != nil {
		h.renderAccountError(w, "Failed to update email")
		return
	}

	h.tmpls.ExecuteTemplate(w, "link-sent.html", nil)
}

func (h *AccountHandlers) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")

	if len(password) < 8 {
		h.renderAccountError(w, "Password must be at least 8 characters")
		return
	}

	pwdHash, err := h.pwdHasher.Hash(password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(int)

	if err := h.userRepo.UpdatePassword(ctx, userID, pwdHash); err != nil {
		h.renderAccountError(w, "Failed to update password")
		return
	}

	data := AccountPageData{
		Message:           "Password updated successfully",
		ChangeEmailURL:    "/account/email",
		ChangePasswordURL: "/account/password",
		DeleteAccountURL:  "/account/delete",
	}
	h.tmpls.ExecuteTemplate(w, "account.html", data)
}

func (h *AccountHandlers) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	userID := r.Context().Value(middleware.UserIDKey).(int)

	if err := h.userRepo.Delete(ctx, userID); err != nil {
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}

	h.tmpls.ExecuteTemplate(w, "success.html", nil)
}

func (h *AccountHandlers) renderAccountError(w http.ResponseWriter, errMsg string) {
	data := AccountPageData{
		Error:             errMsg,
		ChangeEmailURL:    "/account/email",
		ChangePasswordURL: "/account/password",
		DeleteAccountURL:  "/account/delete",
	}
	h.tmpls.ExecuteTemplate(w, "account.html", data)
}
