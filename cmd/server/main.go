package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"blog-studio-web/internal/auth"
	"blog-studio-web/internal/config"
	"blog-studio-web/internal/httpapi"
)

func main() {
	if len(os.Args) == 3 && os.Args[1] == "hash-password" {
		hash, err := auth.HashPassword(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(hash)
		return
	}
	adminHash := os.Getenv("BLOG_STUDIO_ADMIN_PASSWORD_HASH")
	sessionSecret := os.Getenv("BLOG_STUDIO_SESSION_SECRET")
	if adminHash == "" || sessionSecret == "" || len(sessionSecret) < 32 {
		log.Fatal("BLOG_STUDIO_ADMIN_PASSWORD_HASH and BLOG_STUDIO_SESSION_SECRET (32+ chars) are required")
	}
	paths := config.DefaultPaths()
	store := config.NewStore(paths)
	cfg, cfgErr := store.Load()
	if cfgErr != nil {
		log.Fatal(cfgErr.Error())
	}
	if err := store.Save(cfg); err != nil {
		log.Fatal(err.Error())
	}
	sessions := auth.NewStore(sessionSecret, 12*time.Hour, cfg.BasePath)
	server := httpapi.New(paths, cfg, store, adminHash, sessions)
	addr := ":" + env("PORT", "8080")
	log.Printf("blog-studio-web listening on %s%s", addr, strings.TrimRight(cfg.BasePath, "/"))
	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
