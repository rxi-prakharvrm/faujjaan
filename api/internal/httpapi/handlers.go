package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"clothes-shop/api/internal/auth"
	"clothes-shop/api/internal/razorpay"
	"clothes-shop/api/internal/store"
)

func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.store.ListProductsWithVariants(r.Context(), true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list products")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"products": products})
}

func (s *Server) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := s.store.GetProductBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load product")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleCreateCart(w http.ResponseWriter, r *http.Request) {
	id, err := s.store.CreateCart(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create cart")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"cart_id": id})
}

func (s *Server) handleGetCart(w http.ResponseWriter, r *http.Request) {
	cid, err := uuid.Parse(chi.URLParam(r, "cartID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cart_id")
		return
	}
	cart, err := s.store.GetCart(r.Context(), cid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "cart not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load cart")
		return
	}
	writeJSON(w, http.StatusOK, cart)
}

type upsertCartItemRequest struct {
	VariantID string `json:"variant_id"`
	Quantity  int    `json:"quantity"`
}

func (s *Server) handleUpsertCartItem(w http.ResponseWriter, r *http.Request) {
	cid, err := uuid.Parse(chi.URLParam(r, "cartID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cart_id")
		return
	}
	var req upsertCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Quantity <= 0 || req.Quantity > 20 {
		writeError(w, http.StatusBadRequest, "invalid quantity")
		return
	}
	vid, err := uuid.Parse(req.VariantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid variant_id")
		return
	}
	if err := s.store.UpsertCartItem(r.Context(), cid, vid, req.Quantity); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update cart item")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleDeleteCartItem(w http.ResponseWriter, r *http.Request) {
	cid, err := uuid.Parse(chi.URLParam(r, "cartID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cart_id")
		return
	}
	vid, err := uuid.Parse(chi.URLParam(r, "variantID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid variant_id")
		return
	}
	if err := s.store.DeleteCartItem(r.Context(), cid, vid); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete cart item")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type checkoutRequest struct {
	CartID string         `json:"cart_id"`
	Name   string         `json:"customer_name"`
	Phone  string         `json:"customer_phone"`
	Email  string         `json:"customer_email"`
	Addr   map[string]any `json:"shipping_address"`
}

func (s *Server) handleCheckoutFromCart(w http.ResponseWriter, r *http.Request) {
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	cid, err := uuid.Parse(req.CartID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cart_id")
		return
	}
	if req.Name == "" || req.Phone == "" {
		writeError(w, http.StatusBadRequest, "customer_name and customer_phone are required")
		return
	}
	res, err := s.store.CheckoutFromCart(
		r.Context(),
		cid,
		store.CheckoutCustomer{Name: req.Name, Phone: req.Phone, Email: req.Email, Address: req.Addr},
		s.cfg.ShippingFlatINR,
		s.cfg.TaxRateBps,
	)
	if err != nil {
		if errors.Is(err, store.ErrInsufficientStock) {
			writeError(w, http.StatusConflict, "insufficient stock")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var rz *struct {
		KeyID    string `json:"key_id"`
		OrderID  string `json:"order_id"`
		AmountINR int   `json:"amount_inr"`
		Currency string `json:"currency"`
	}
	if s.cfg.RazorpayKeyID != "" && s.cfg.RazorpayKeySecret != "" {
		rzc := razorpay.NewClient(s.cfg.RazorpayKeyID, s.cfg.RazorpayKeySecret)
		rzOrderID, err := rzc.CreateOrder(r.Context(), res.AmountINR, res.Currency, res.OrderID.String())
		if err != nil {
			writeError(w, http.StatusBadGateway, "failed to create razorpay order")
			return
		}
		if err := s.store.SetPaymentRazorpayOrderID(r.Context(), res.PaymentID, rzOrderID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to persist razorpay order")
			return
		}
		rz = &struct {
			KeyID    string `json:"key_id"`
			OrderID  string `json:"order_id"`
			AmountINR int   `json:"amount_inr"`
			Currency string `json:"currency"`
		}{
			KeyID:    s.cfg.RazorpayKeyID,
			OrderID:  rzOrderID,
			AmountINR: res.AmountINR,
			Currency: res.Currency,
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"order_id":   res.OrderID,
		"payment_id": res.PaymentID,
		"amount_inr": res.AmountINR,
		"currency":   res.Currency,
		"provider":   res.Provider,
		"razorpay":   rz,
	})
}

type adminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	var req adminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password required")
		return
	}

	u, err := s.store.GetAdminUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !auth.CheckPasswordHash(req.Password, u.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := s.auth.IssueAdminToken(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "role": "admin"})
}

func (s *Server) handleAdminListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.store.ListProductsWithVariants(r.Context(), false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list products")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"products": products})
}

type adminCreateProductRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Variants    []struct {
		SKU       string `json:"sku"`
		Title     string `json:"title"`
		Size      string `json:"size"`
		Color     string `json:"color"`
		PriceINR  int    `json:"price_inr"`
		OnHand    int    `json:"on_hand"`
	} `json:"variants"`
}

func (s *Server) handleAdminCreateProduct(w http.ResponseWriter, r *http.Request) {
	var req adminCreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Slug == "" || req.Name == "" || len(req.Variants) == 0 {
		writeError(w, http.StatusBadRequest, "slug, name, and at least one variant are required")
		return
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	var variants []store.CreateVariantInput
	for _, v := range req.Variants {
		if v.SKU == "" || v.PriceINR < 0 || v.OnHand < 0 {
			writeError(w, http.StatusBadRequest, "invalid variant")
			return
		}
		variants = append(variants, store.CreateVariantInput{
			SKU:      v.SKU,
			Title:    v.Title,
			Size:     v.Size,
			Color:    v.Color,
			PriceINR: v.PriceINR,
			OnHand:   v.OnHand,
		})
	}
	p, err := s.store.AdminCreateProduct(r.Context(), store.CreateProductInput{
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Variants:    variants,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

type adminUpdateProductRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (s *Server) handleAdminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	pid, err := uuid.Parse(chi.URLParam(r, "productID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product_id")
		return
	}
	var req adminUpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Slug == "" || req.Name == "" || req.Status == "" {
		writeError(w, http.StatusBadRequest, "slug, name, status required")
		return
	}
	if err := s.store.AdminUpdateProduct(r.Context(), pid, req.Slug, req.Name, req.Description, req.Status); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type adminCreateVariantRequest struct {
	SKU      string `json:"sku"`
	Title    string `json:"title"`
	Size     string `json:"size"`
	Color    string `json:"color"`
	PriceINR int    `json:"price_inr"`
	OnHand   int    `json:"on_hand"`
}

func (s *Server) handleAdminCreateVariant(w http.ResponseWriter, r *http.Request) {
	pid, err := uuid.Parse(chi.URLParam(r, "productID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product_id")
		return
	}
	var req adminCreateVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.SKU == "" || req.PriceINR < 0 || req.OnHand < 0 {
		writeError(w, http.StatusBadRequest, "invalid variant")
		return
	}
	v, err := s.store.AdminCreateVariant(r.Context(), pid, store.CreateVariantForProductInput{
		SKU:      req.SKU,
		Title:    req.Title,
		Size:     req.Size,
		Color:    req.Color,
		PriceINR: req.PriceINR,
		OnHand:   req.OnHand,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

type adminUpdateVariantRequest struct {
	Title    string `json:"title"`
	Size     string `json:"size"`
	Color    string `json:"color"`
	PriceINR int    `json:"price_inr"`
}

func (s *Server) handleAdminUpdateVariant(w http.ResponseWriter, r *http.Request) {
	vid, err := uuid.Parse(chi.URLParam(r, "variantID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid variant_id")
		return
	}
	var req adminUpdateVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.PriceINR < 0 {
		writeError(w, http.StatusBadRequest, "invalid price_inr")
		return
	}
	if err := s.store.AdminUpdateVariant(r.Context(), vid, req.Title, req.Size, req.Color, req.PriceINR, nil); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "variant not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type adminAdjustInventoryRequest struct {
	VariantID string `json:"variant_id"`
	Delta     int    `json:"delta"`
}

func (s *Server) handleAdminAdjustInventory(w http.ResponseWriter, r *http.Request) {
	var req adminAdjustInventoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	vid, err := uuid.Parse(req.VariantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid variant_id")
		return
	}
	onHand, reserved, err := s.store.AdminAdjustInventory(r.Context(), vid, req.Delta)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"on_hand": onHand, "reserved": reserved})
}

func (s *Server) handleAdminListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := s.store.AdminListOrders(r.Context(), 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list orders")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": orders})
}

func (s *Server) handleAdminGetOrder(w http.ResponseWriter, r *http.Request) {
	oid, err := uuid.Parse(chi.URLParam(r, "orderID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid order_id")
		return
	}
	o, err := s.store.AdminGetOrder(r.Context(), oid)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load order")
		return
	}
	writeJSON(w, http.StatusOK, o)
}

