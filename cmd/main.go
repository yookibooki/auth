package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yookibooki/auth/auth"
	"github.com/yookibooki/auth/config"
	"github.com/yookibooki/auth/db"
	"github.com/yookibooki/auth/email"
	"github.com/yookibooki/auth/handlers"
	"github.com/yookibooki/auth/middleware"
	"github.com/yookibooki/auth/repo"
	"github.com/yookibooki/auth/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(database)

	tmpls := web.Parse()

	userRepo := repo.NewUserRepo(database)
	authCodeRepo := repo.NewAuthCodeRepo(database)
	pwdResetRepo := repo.NewPwdResetTokenRepo(database)

	pwdHasher := auth.NewPasswordHasher()
	tokenGenerator := auth.NewSecureTokenGenerator(32)
	emailValidator := auth.NewEmailValidator()
	authCodeMgr := auth.NewAuthCodeManager(pwdHasher, tokenGenerator, 15*time.Minute)

	emailSender := email.NewSMTPSender(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.User,
		cfg.SMTP.Password,
		cfg.SMTP.From,
	)

	baseURL := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)

	authHandlers := handlers.NewAuthHandlers(
		tmpls,
		userRepo,
		authCodeRepo,
		pwdHasher,
		authCodeMgr,
		emailValidator,
		emailSender,
		baseURL,
	)

	pwdResetHandlers := handlers.NewPwdResetHandlers(
		tmpls,
		userRepo,
		pwdResetRepo,
		pwdHasher,
		authCodeMgr,
		emailSender,
		baseURL,
	)

	accountHandlers := handlers.NewAccountHandlers(
		tmpls,
		userRepo,
		pwdHasher,
		emailValidator,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/", authHandlers.ServeAuth)
	mux.HandleFunc("/auth", authHandlers.ServeAuth)
	mux.HandleFunc("/auth/email", authHandlers.HandleEmail)
	mux.HandleFunc("/auth/password", authHandlers.HandlePassword)
	mux.HandleFunc("/auth/confirm", authHandlers.HandleConfirm)

	mux.HandleFunc("/reset", pwdResetHandlers.ServeReset)
	mux.HandleFunc("/reset/request", pwdResetHandlers.HandleRequest)
	mux.HandleFunc("/reset/confirm", pwdResetHandlers.HandleConfirm)
	mux.HandleFunc("/reset/complete", pwdResetHandlers.HandleComplete)

	accountMux := http.NewServeMux()
	accountMux.HandleFunc("/account", accountHandlers.ServeAccount)
	accountMux.HandleFunc("/account/email", accountHandlers.HandleChangeEmail)
	accountMux.HandleFunc("/account/password", accountHandlers.HandleChangePassword)
	accountMux.HandleFunc("/account/delete", accountHandlers.HandleDeleteAccount)

	mux.Handle("/account/", middleware.Auth(userRepo)(accountMux))

	handler := middleware.Recovery(middleware.Logger(mux))

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
