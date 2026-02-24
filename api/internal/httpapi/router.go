package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"clothes-shop/api/internal/auth"
	"clothes-shop/api/internal/config"
	"clothes-shop/api/internal/store"
)

type Server struct {
	cfg   config.Config
	store *store.Store
	auth  *auth.Service
}

func New(cfg config.Config, store *store.Store, authSvc *auth.Service) *Server {
	return &Server{cfg: cfg, store: store, auth: authSvc}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	corsOpts := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: false,
		MaxAge:           300,
	}
	if s.cfg.DevAllowAllCORS {
		corsOpts.AllowedOrigins = []string{"*"}
	} else if s.cfg.AllowedCORSOrigin != "" {
		corsOpts.AllowedOrigins = []string{s.cfg.AllowedCORSOrigin}
	}
	r.Use(cors.Handler(corsOpts))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Get("/products", s.handleListProducts)
		r.Get("/products/{slug}", s.handleGetProduct)

		r.Post("/cart", s.handleCreateCart)
		r.Get("/cart/{cartID}", s.handleGetCart)
		r.Post("/cart/{cartID}/items", s.handleUpsertCartItem)
		r.Delete("/cart/{cartID}/items/{variantID}", s.handleDeleteCartItem)

		r.Post("/checkout", s.handleCheckoutFromCart)
		r.Post("/payments/razorpay/verify", s.handleRazorpayVerify)
		r.Post("/webhooks/razorpay", s.handleRazorpayWebhook)

		r.Route("/admin", func(r chi.Router) {
			r.Post("/login", s.handleAdminLogin)

			r.Group(func(r chi.Router) {
				r.Use(s.auth.Middleware)
				r.Use(auth.RequireRole(auth.RoleAdmin))

				r.Get("/products", s.handleAdminListProducts)
				r.Post("/products", s.handleAdminCreateProduct)
				r.Put("/products/{productID}", s.handleAdminUpdateProduct)
				r.Post("/products/{productID}/variants", s.handleAdminCreateVariant)
				r.Put("/variants/{variantID}", s.handleAdminUpdateVariant)
				r.Post("/inventory/adjust", s.handleAdminAdjustInventory)

				r.Get("/orders", s.handleAdminListOrders)
				r.Get("/orders/{orderID}", s.handleAdminGetOrder)
			})
		})
	})

	return r
}

