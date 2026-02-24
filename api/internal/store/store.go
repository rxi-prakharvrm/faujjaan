package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrInsufficientStock = errors.New("insufficient stock")

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

type AdminUser struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

func (s *Store) GetAdminUserByEmail(ctx context.Context, email string) (AdminUser, error) {
	row := s.db.QueryRow(ctx, `
SELECT u.id, u.email, u.password_hash
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r ON r.id = ur.role_id
WHERE u.email = $1 AND u.is_active = TRUE AND r.key = 'admin'
LIMIT 1
`, email)
	var u AdminUser
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminUser{}, ErrNotFound
		}
		return AdminUser{}, err
	}
	return u, nil
}

type Product struct {
	ID          uuid.UUID `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Variants    []Variant `json:"variants"`
}

type Variant struct {
	ID               uuid.UUID `json:"id"`
	ProductID        uuid.UUID `json:"product_id"`
	SKU              string    `json:"sku"`
	Title            string    `json:"title"`
	Size             string    `json:"size"`
	Color            string    `json:"color"`
	PriceINR         int       `json:"price_inr"`
	CompareAtPriceINR *int      `json:"compare_at_price_inr,omitempty"`
	OnHand           int       `json:"on_hand"`
	Reserved         int       `json:"reserved"`
}

func (s *Store) ListProductsWithVariants(ctx context.Context, onlyActive bool) ([]Product, error) {
	where := "TRUE"
	if onlyActive {
		where = "p.status = 'active'"
	}
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
SELECT
  p.id, p.slug, p.name, p.description, p.status, p.created_at, p.updated_at,
  v.id, v.product_id, v.sku, v.title, v.size, v.color, v.price_inr, v.compare_at_price_inr,
  COALESCE(i.on_hand, 0), COALESCE(i.reserved, 0)
FROM products p
JOIN product_variants v ON v.product_id = p.id
LEFT JOIN inventory i ON i.variant_id = v.id
WHERE %s
ORDER BY p.created_at DESC, v.created_at ASC
`, where))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := map[uuid.UUID]*Product{}
	order := make([]uuid.UUID, 0, 64)
	for rows.Next() {
		var p Product
		var v Variant
		if err := rows.Scan(
			&p.ID, &p.Slug, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt,
			&v.ID, &v.ProductID, &v.SKU, &v.Title, &v.Size, &v.Color, &v.PriceINR, &v.CompareAtPriceINR,
			&v.OnHand, &v.Reserved,
		); err != nil {
			return nil, err
		}
		existing := byID[p.ID]
		if existing == nil {
			p.Variants = []Variant{v}
			byID[p.ID] = &p
			order = append(order, p.ID)
		} else {
			existing.Variants = append(existing.Variants, v)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]Product, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out, nil
}

func (s *Store) GetProductBySlug(ctx context.Context, slug string) (Product, error) {
	rows, err := s.db.Query(ctx, `
SELECT
  p.id, p.slug, p.name, p.description, p.status, p.created_at, p.updated_at,
  v.id, v.product_id, v.sku, v.title, v.size, v.color, v.price_inr, v.compare_at_price_inr,
  COALESCE(i.on_hand, 0), COALESCE(i.reserved, 0)
FROM products p
JOIN product_variants v ON v.product_id = p.id
LEFT JOIN inventory i ON i.variant_id = v.id
WHERE p.slug = $1
ORDER BY v.created_at ASC
`, slug)
	if err != nil {
		return Product{}, err
	}
	defer rows.Close()

	var p Product
	first := true
	for rows.Next() {
		var v Variant
		if err := rows.Scan(
			&p.ID, &p.Slug, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt,
			&v.ID, &v.ProductID, &v.SKU, &v.Title, &v.Size, &v.Color, &v.PriceINR, &v.CompareAtPriceINR,
			&v.OnHand, &v.Reserved,
		); err != nil {
			return Product{}, err
		}
		if first {
			p.Variants = []Variant{}
			first = false
		}
		p.Variants = append(p.Variants, v)
	}
	if err := rows.Err(); err != nil {
		return Product{}, err
	}
	if first {
		return Product{}, ErrNotFound
	}
	return p, nil
}

type CreateProductInput struct {
	Slug        string
	Name        string
	Description string
	Status      string
	Variants    []CreateVariantInput
}

type CreateVariantInput struct {
	SKU              string
	Title            string
	Size             string
	Color            string
	PriceINR         int
	CompareAtPriceINR *int
	OnHand           int
}

func (s *Store) AdminCreateProduct(ctx context.Context, in CreateProductInput) (Product, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Product{}, err
	}
	defer tx.Rollback(ctx)

	var pid uuid.UUID
	var createdAt, updatedAt time.Time
	err = tx.QueryRow(ctx, `
INSERT INTO products (slug, name, description, status)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, updated_at
`, in.Slug, in.Name, in.Description, in.Status).Scan(&pid, &createdAt, &updatedAt)
	if err != nil {
		return Product{}, err
	}

	out := Product{
		ID:          pid,
		Slug:        in.Slug,
		Name:        in.Name,
		Description: in.Description,
		Status:      in.Status,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Variants:    make([]Variant, 0, len(in.Variants)),
	}

	for _, v := range in.Variants {
		var vid uuid.UUID
		var vCreatedAt, vUpdatedAt time.Time
		err := tx.QueryRow(ctx, `
INSERT INTO product_variants (product_id, sku, title, size, color, price_inr, compare_at_price_inr)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING id, created_at, updated_at
`, pid, v.SKU, v.Title, v.Size, v.Color, v.PriceINR, v.CompareAtPriceINR).Scan(&vid, &vCreatedAt, &vUpdatedAt)
		if err != nil {
			return Product{}, err
		}
		_, err = tx.Exec(ctx, `
INSERT INTO inventory (variant_id, on_hand, reserved)
VALUES ($1, $2, 0)
ON CONFLICT (variant_id) DO UPDATE SET on_hand = EXCLUDED.on_hand, reserved = 0, updated_at = now()
`, vid, v.OnHand)
		if err != nil {
			return Product{}, err
		}
		out.Variants = append(out.Variants, Variant{
			ID:               vid,
			ProductID:        pid,
			SKU:              v.SKU,
			Title:            v.Title,
			Size:             v.Size,
			Color:            v.Color,
			PriceINR:         v.PriceINR,
			CompareAtPriceINR: v.CompareAtPriceINR,
			OnHand:           v.OnHand,
			Reserved:         0,
		})
		_ = vCreatedAt
		_ = vUpdatedAt
	}

	if err := tx.Commit(ctx); err != nil {
		return Product{}, err
	}
	return out, nil
}

func (s *Store) AdminUpdateProduct(ctx context.Context, productID uuid.UUID, slug, name, description, status string) error {
	ct, err := s.db.Exec(ctx, `
UPDATE products
SET slug=$2, name=$3, description=$4, status=$5, updated_at=now()
WHERE id=$1
`, productID, slug, name, description, status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type CreateVariantForProductInput struct {
	SKU              string
	Title            string
	Size             string
	Color            string
	PriceINR         int
	CompareAtPriceINR *int
	OnHand           int
}

func (s *Store) AdminCreateVariant(ctx context.Context, productID uuid.UUID, in CreateVariantForProductInput) (Variant, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Variant{}, err
	}
	defer tx.Rollback(ctx)

	var vid uuid.UUID
	err = tx.QueryRow(ctx, `
INSERT INTO product_variants (product_id, sku, title, size, color, price_inr, compare_at_price_inr)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING id
`, productID, in.SKU, in.Title, in.Size, in.Color, in.PriceINR, in.CompareAtPriceINR).Scan(&vid)
	if err != nil {
		return Variant{}, err
	}

	_, err = tx.Exec(ctx, `
INSERT INTO inventory (variant_id, on_hand, reserved)
VALUES ($1, $2, 0)
`, vid, in.OnHand)
	if err != nil {
		return Variant{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Variant{}, err
	}
	return Variant{
		ID:               vid,
		ProductID:        productID,
		SKU:              in.SKU,
		Title:            in.Title,
		Size:             in.Size,
		Color:            in.Color,
		PriceINR:         in.PriceINR,
		CompareAtPriceINR: in.CompareAtPriceINR,
		OnHand:           in.OnHand,
		Reserved:         0,
	}, nil
}

func (s *Store) AdminUpdateVariant(ctx context.Context, variantID uuid.UUID, title, size, color string, priceINR int, compareAt *int) error {
	ct, err := s.db.Exec(ctx, `
UPDATE product_variants
SET title=$2, size=$3, color=$4, price_inr=$5, compare_at_price_inr=$6, updated_at=now()
WHERE id=$1
`, variantID, title, size, color, priceINR, compareAt)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) AdminAdjustInventory(ctx context.Context, variantID uuid.UUID, delta int) (onHand int, reserved int, err error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	var curOnHand, curReserved int
	if err := tx.QueryRow(ctx, `
SELECT on_hand, reserved FROM inventory WHERE variant_id=$1 FOR UPDATE
`, variantID).Scan(&curOnHand, &curReserved); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, ErrNotFound
		}
		return 0, 0, err
	}
	newOnHand := curOnHand + delta
	if newOnHand < 0 {
		return 0, 0, errors.New("on_hand would go negative")
	}
	_, err = tx.Exec(ctx, `
UPDATE inventory SET on_hand=$2, updated_at=now() WHERE variant_id=$1
`, variantID, newOnHand)
	if err != nil {
		return 0, 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return newOnHand, curReserved, nil
}

type CartItem struct {
	VariantID uuid.UUID `json:"variant_id"`
	SKU       string    `json:"sku"`
	Product   string    `json:"product_name"`
	Variant   string    `json:"variant_title"`
	UnitPrice int       `json:"unit_price_inr"`
	Quantity  int       `json:"quantity"`
	LineTotal int       `json:"line_total_inr"`
}

type Cart struct {
	ID       uuid.UUID  `json:"id"`
	Status   string     `json:"status"`
	Items    []CartItem `json:"items"`
	Subtotal int        `json:"subtotal_inr"`
}

func (s *Store) CreateCart(ctx context.Context) (uuid.UUID, error) {
	var id uuid.UUID
	if err := s.db.QueryRow(ctx, `INSERT INTO carts DEFAULT VALUES RETURNING id`).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *Store) UpsertCartItem(ctx context.Context, cartID uuid.UUID, variantID uuid.UUID, qty int) error {
	ct, err := s.db.Exec(ctx, `
INSERT INTO cart_items (cart_id, variant_id, quantity)
VALUES ($1,$2,$3)
ON CONFLICT (cart_id, variant_id) DO UPDATE SET quantity = EXCLUDED.quantity, updated_at = now()
`, cartID, variantID, qty)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteCartItem(ctx context.Context, cartID, variantID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM cart_items WHERE cart_id=$1 AND variant_id=$2`, cartID, variantID)
	return err
}

func (s *Store) GetCart(ctx context.Context, cartID uuid.UUID) (Cart, error) {
	row := s.db.QueryRow(ctx, `SELECT id, status FROM carts WHERE id=$1`, cartID)
	var c Cart
	if err := row.Scan(&c.ID, &c.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Cart{}, ErrNotFound
		}
		return Cart{}, err
	}

	rows, err := s.db.Query(ctx, `
SELECT
  ci.variant_id,
  v.sku,
  p.name,
  v.title,
  v.price_inr,
  ci.quantity
FROM cart_items ci
JOIN product_variants v ON v.id = ci.variant_id
JOIN products p ON p.id = v.product_id
WHERE ci.cart_id=$1
ORDER BY ci.created_at ASC
`, cartID)
	if err != nil {
		return Cart{}, err
	}
	defer rows.Close()

	c.Items = []CartItem{}
	sub := 0
	for rows.Next() {
		var it CartItem
		if err := rows.Scan(&it.VariantID, &it.SKU, &it.Product, &it.Variant, &it.UnitPrice, &it.Quantity); err != nil {
			return Cart{}, err
		}
		it.LineTotal = it.UnitPrice * it.Quantity
		sub += it.LineTotal
		c.Items = append(c.Items, it)
	}
	if err := rows.Err(); err != nil {
		return Cart{}, err
	}
	c.Subtotal = sub
	return c, nil
}

type CheckoutCustomer struct {
	Name    string
	Phone   string
	Email   string
	Address map[string]any
}

type CheckoutResult struct {
	OrderID    uuid.UUID `json:"order_id"`
	PaymentID  uuid.UUID `json:"payment_id"`
	AmountINR  int       `json:"amount_inr"`
	Currency   string    `json:"currency"`
	Provider   string    `json:"provider"`
}

func (s *Store) CheckoutFromCart(ctx context.Context, cartID uuid.UUID, customer CheckoutCustomer, shippingFlatINR int, taxRateBps int) (CheckoutResult, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CheckoutResult{}, err
	}
	defer tx.Rollback(ctx)

	// Load cart items with price and lock inventory rows for reservation.
	rows, err := tx.Query(ctx, `
SELECT
  ci.variant_id,
  v.sku,
  p.name,
  v.title,
  v.price_inr,
  ci.quantity
FROM cart_items ci
JOIN product_variants v ON v.id = ci.variant_id
JOIN products p ON p.id = v.product_id
WHERE ci.cart_id=$1
ORDER BY ci.created_at ASC
`, cartID)
	if err != nil {
		return CheckoutResult{}, err
	}
	defer rows.Close()

	type line struct {
		variantID uuid.UUID
		sku       string
		pname     string
		vtitle    string
		unit      int
		qty       int
	}
	lines := []line{}
	subtotal := 0
	for rows.Next() {
		var l line
		if err := rows.Scan(&l.variantID, &l.sku, &l.pname, &l.vtitle, &l.unit, &l.qty); err != nil {
			return CheckoutResult{}, err
		}
		lines = append(lines, l)
		subtotal += l.unit * l.qty
	}
	if err := rows.Err(); err != nil {
		return CheckoutResult{}, err
	}
	if len(lines) == 0 {
		return CheckoutResult{}, errors.New("cart is empty")
	}

	shipping := shippingFlatINR
	taxBase := subtotal + shipping
	tax := (taxBase*taxRateBps + 5000) / 10000
	total := subtotal + shipping + tax

	var orderID uuid.UUID
	err = tx.QueryRow(ctx, `
INSERT INTO orders (status, currency, subtotal_inr, shipping_inr, tax_inr, total_inr, customer_name, customer_phone, customer_email, shipping_address)
VALUES ('pending_payment','INR',$1,$2,$3,$4,$5,$6,$7,$8)
RETURNING id
`, subtotal, shipping, tax, total, customer.Name, customer.Phone, customer.Email, customer.Address).Scan(&orderID)
	if err != nil {
		return CheckoutResult{}, err
	}

	for _, l := range lines {
		lineTotal := l.unit * l.qty
		_, err := tx.Exec(ctx, `
INSERT INTO order_items (order_id, variant_id, sku, product_name, variant_title, unit_price_inr, quantity, line_total_inr)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
`, orderID, l.variantID, l.sku, l.pname, l.vtitle, l.unit, l.qty, lineTotal)
		if err != nil {
			return CheckoutResult{}, err
		}

		// Reserve inventory
		var onHand, reserved int
		if err := tx.QueryRow(ctx, `
SELECT on_hand, reserved FROM inventory WHERE variant_id=$1 FOR UPDATE
`, l.variantID).Scan(&onHand, &reserved); err != nil {
			return CheckoutResult{}, err
		}
		if onHand-reserved < l.qty {
			return CheckoutResult{}, ErrInsufficientStock
		}
		_, err = tx.Exec(ctx, `
UPDATE inventory SET reserved = reserved + $2, updated_at=now() WHERE variant_id=$1
`, l.variantID, l.qty)
		if err != nil {
			return CheckoutResult{}, err
		}
	}

	var paymentID uuid.UUID
	err = tx.QueryRow(ctx, `
INSERT INTO payments (order_id, provider, status, amount_inr)
VALUES ($1,'razorpay','created',$2)
RETURNING id
`, orderID, total).Scan(&paymentID)
	if err != nil {
		return CheckoutResult{}, err
	}

	_, err = tx.Exec(ctx, `UPDATE carts SET status='checked_out', updated_at=now() WHERE id=$1`, cartID)
	if err != nil {
		return CheckoutResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return CheckoutResult{}, err
	}

	return CheckoutResult{
		OrderID:   orderID,
		PaymentID: paymentID,
		AmountINR: total,
		Currency:  "INR",
		Provider:  "razorpay",
	}, nil
}

func (s *Store) SetPaymentRazorpayOrderID(ctx context.Context, paymentID uuid.UUID, razorpayOrderID string) error {
	ct, err := s.db.Exec(ctx, `
UPDATE payments
SET razorpay_order_id=$2, updated_at=now()
WHERE id=$1
`, paymentID, razorpayOrderID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) MarkPaymentAuthorized(ctx context.Context, razorpayOrderID, razorpayPaymentID, razorpaySignature string) error {
	ct, err := s.db.Exec(ctx, `
UPDATE payments
SET status='authorized', razorpay_payment_id=$2, razorpay_signature=$3, updated_at=now()
WHERE razorpay_order_id=$1 AND status IN ('created','authorized')
`, razorpayOrderID, razorpayPaymentID, razorpaySignature)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		// Either not found or already in a terminal state; treat as idempotent.
		return nil
	}
	return nil
}

func (s *Store) MarkPaymentCaptured(ctx context.Context, razorpayOrderID, razorpayPaymentID string) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var paymentID uuid.UUID
	var orderID uuid.UUID
	var payStatus string
	if err := tx.QueryRow(ctx, `
SELECT id, order_id, status
FROM payments
WHERE razorpay_order_id=$1
FOR UPDATE
`, razorpayOrderID).Scan(&paymentID, &orderID, &payStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if payStatus == "captured" {
		return tx.Commit(ctx)
	}

	_, err = tx.Exec(ctx, `
UPDATE payments
SET status='captured', razorpay_payment_id=$2, updated_at=now()
WHERE id=$1
`, paymentID, razorpayPaymentID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE orders
SET status='paid', updated_at=now()
WHERE id=$1 AND status='pending_payment'
`, orderID)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `SELECT variant_id, quantity FROM order_items WHERE order_id=$1`, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var variantID uuid.UUID
		var qty int
		if err := rows.Scan(&variantID, &qty); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
UPDATE inventory
SET on_hand = on_hand - $2,
    reserved = GREATEST(0, reserved - $2),
    updated_at=now()
WHERE variant_id=$1
`, variantID, qty)
		if err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) MarkPaymentFailed(ctx context.Context, razorpayOrderID, razorpayPaymentID string) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var paymentID uuid.UUID
	var orderID uuid.UUID
	var payStatus string
	if err := tx.QueryRow(ctx, `
SELECT id, order_id, status
FROM payments
WHERE razorpay_order_id=$1
FOR UPDATE
`, razorpayOrderID).Scan(&paymentID, &orderID, &payStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if payStatus == "failed" || payStatus == "captured" {
		return tx.Commit(ctx)
	}

	_, err = tx.Exec(ctx, `
UPDATE payments
SET status='failed', razorpay_payment_id=$2, updated_at=now()
WHERE id=$1
`, paymentID, razorpayPaymentID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
UPDATE orders
SET status='failed', updated_at=now()
WHERE id=$1 AND status='pending_payment'
`, orderID)
	if err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `SELECT variant_id, quantity FROM order_items WHERE order_id=$1`, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var variantID uuid.UUID
		var qty int
		if err := rows.Scan(&variantID, &qty); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
UPDATE inventory
SET reserved = GREATEST(0, reserved - $2),
    updated_at=now()
WHERE variant_id=$1
`, variantID, qty)
		if err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type OrderSummary struct {
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	TotalINR  int       `json:"total_inr"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Store) AdminListOrders(ctx context.Context, limit int) ([]OrderSummary, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.Query(ctx, `
SELECT id, status, total_inr, created_at
FROM orders
ORDER BY created_at DESC
LIMIT $1
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OrderSummary{}
	for rows.Next() {
		var o OrderSummary
		if err := rows.Scan(&o.ID, &o.Status, &o.TotalINR, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

type OrderDetail struct {
	ID             uuid.UUID        `json:"id"`
	Status         string           `json:"status"`
	SubtotalINR    int              `json:"subtotal_inr"`
	ShippingINR    int              `json:"shipping_inr"`
	TaxINR         int              `json:"tax_inr"`
	TotalINR       int              `json:"total_inr"`
	CustomerName   string           `json:"customer_name"`
	CustomerPhone  string           `json:"customer_phone"`
	CustomerEmail  string           `json:"customer_email"`
	ShippingAddr   map[string]any   `json:"shipping_address"`
	Items          []CartItem       `json:"items"`
	PaymentStatus  string           `json:"payment_status"`
	RazorpayOrderID string          `json:"razorpay_order_id"`
	CreatedAt      time.Time        `json:"created_at"`
}

func (s *Store) AdminGetOrder(ctx context.Context, orderID uuid.UUID) (OrderDetail, error) {
	var o OrderDetail
	err := s.db.QueryRow(ctx, `
SELECT id, status, subtotal_inr, shipping_inr, tax_inr, total_inr,
       customer_name, customer_phone, customer_email, shipping_address, created_at
FROM orders
WHERE id=$1
`, orderID).Scan(
		&o.ID, &o.Status, &o.SubtotalINR, &o.ShippingINR, &o.TaxINR, &o.TotalINR,
		&o.CustomerName, &o.CustomerPhone, &o.CustomerEmail, &o.ShippingAddr, &o.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OrderDetail{}, ErrNotFound
		}
		return OrderDetail{}, err
	}

	err = s.db.QueryRow(ctx, `
SELECT status, razorpay_order_id
FROM payments
WHERE order_id=$1
`, orderID).Scan(&o.PaymentStatus, &o.RazorpayOrderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			o.PaymentStatus = "missing"
		} else {
			return OrderDetail{}, err
		}
	}

	rows, err := s.db.Query(ctx, `
SELECT variant_id, sku, product_name, variant_title, unit_price_inr, quantity
FROM order_items
WHERE order_id=$1
`, orderID)
	if err != nil {
		return OrderDetail{}, err
	}
	defer rows.Close()

	o.Items = []CartItem{}
	for rows.Next() {
		var it CartItem
		if err := rows.Scan(&it.VariantID, &it.SKU, &it.Product, &it.Variant, &it.UnitPrice, &it.Quantity); err != nil {
			return OrderDetail{}, err
		}
		it.LineTotal = it.UnitPrice * it.Quantity
		o.Items = append(o.Items, it)
	}
	if err := rows.Err(); err != nil {
		return OrderDetail{}, err
	}
	return o, nil
}

