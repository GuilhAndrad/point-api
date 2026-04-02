package main

import (
    "log"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/jmoiron/sqlx"
    "github.com/joho/godotenv"
    _ "github.com/lib/pq"

    "github.com/GuilhAndrad/point-api/internal/handler"
    "github.com/GuilhAndrad/point-api/internal/repository"
    "github.com/GuilhAndrad/point-api/internal/service"
)

func main() {
    // Carrega variáveis do .env
    godotenv.Load()

    // Conecta ao banco
    db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal("erro ao conectar no banco:", err)
    }
    defer db.Close()

    // Monta as camadas (de baixo pra cima)
    userRepo   := repository.NewUserRepository(db)
    recordRepo := repository.NewRecordRepository(db)

    authService   := service.NewAuthService(userRepo, os.Getenv("JWT_SECRET"))
    punchService  := service.NewPunchService(recordRepo)

    authHandler  := handler.NewAuthHandler(authService)
    punchHandler := handler.NewPunchHandler(punchService)

    // Rotas
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    r.Post("/auth/register", authHandler.Register)
    r.Post("/auth/login",    authHandler.Login)

    // Rotas autenticadas
    r.Group(func(r chi.Router) {
        r.Use(handler.AuthMiddleware(os.Getenv("JWT_SECRET")))

        r.Post("/punch",        punchHandler.Punch)
        r.Get("/my/records",   punchHandler.MyRecords)
        r.Get("/my/report",    punchHandler.MyReport)

        // Apenas admin
        r.Group(func(r chi.Router) {
            r.Use(handler.AdminOnly)
            r.Get("/admin/employees",  punchHandler.AllEmployees)
            r.Get("/admin/report",     punchHandler.AdminReport)
        })
    })

    log.Println("Servidor rodando em :8080")
    http.ListenAndServe(":8080", r)
}