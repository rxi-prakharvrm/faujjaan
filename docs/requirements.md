# MVP requirements (v1)

## Scope

- **Storefront (public)**: browse product catalog, product details, search/filter, cart, checkout initiation, checkout status (success/fail).
- **Admin dashboard**: secure login, CRUD products + variants, adjust inventory, view orders + order detail.
- **Backend**: stateless Go API, Postgres persistence, server-calculated pricing, Razorpay integration + verified webhooks.

Non-goals for MVP:

- Promotions/coupons, returns portal, exchanges, partial shipments, multi-warehouse, multi-currency, customer accounts/self-service.

## Roles & permissions

- **guest**: browse products; create/use a cart; start checkout.
- **admin**: full access to admin endpoints/UI.

Auth model:

- Admin login via email + password.
- API issues JWT access tokens; admin endpoints require `role=admin`.

## Product model & attributes

### Product

- `id` (UUID)
- `slug` (unique, URL-safe)
- `name`
- `description` (rich text later; MVP plain string)
- `status`: `draft` | `active` | `archived`
- `category` (optional for MVP; multiple categories supported in schema)
- `created_at`, `updated_at`

### Variant (size/color)

Each product has **one or more variants**.

- `id` (UUID)
- `product_id`
- `sku` (unique)
- `title` (optional display label; e.g. "Black / M")
- **Attributes**:
  - `size`: freeform string (e.g. `XS`, `S`, `M`, `L`, `XL`, `32`, `34`)
  - `color`: freeform string (e.g. `Black`, `Navy`, `Red`)
- **Pricing** (minor units):
  - `price_inr` (integer paise)
  - `compare_at_price_inr` (optional paise)

### Inventory

Inventory is per-variant:

- `on_hand`: physical stock
- `reserved`: stock held for `pending_payment` orders (to reduce oversell)

Available to sell:

\[
available = \max(0, on\_hand - reserved)
\]

### Images

- Stored as URLs (Supabase Storage now; S3/R2 later).
- Attached to product and/or variant, with `sort_order`.

## Cart behavior (server-side)

- Cart is stored server-side to prevent price tampering.
- Storefront uses a `cart_id` (UUID) persisted client-side (localStorage for MVP).
- Cart items reference `variant_id` and `quantity`.
- Totals are calculated on the server on read/update.

## Order & payment lifecycle

### Order statuses

- `draft`: created but not ready for payment (internal)
- `pending_payment`: created, inventory reserved, awaiting Razorpay payment result
- `paid`: payment captured
- `failed`: payment failed/expired
- `cancelled`: cancelled by admin (or timed out) before fulfillment
- `fulfilled`: shipped/delivered (MVP: admin mark fulfilled)
- `refunded`: refunded after capture (MVP: record-only; no automated refund call)

### Payment statuses

- `created`: Razorpay order created
- `authorized`: payment authorized (optional; depends on capture flow)
- `captured`: payment captured (success)
- `failed`: failed/expired
- `refunded`: refunded

### Allowed transitions (MVP)

- **Order**
  - `draft` → `pending_payment`
  - `pending_payment` → `paid` | `failed` | `cancelled`
  - `paid` → `fulfilled` | `refunded`
  - `failed` → `cancelled` (optional cleanup)
- **Payment**
  - `created` → `authorized` | `captured` | `failed`
  - `authorized` → `captured` | `failed`
  - `captured` → `refunded`

Inventory rules:

- On `pending_payment` creation: reserve inventory (increment `reserved`).
- On `paid`: decrement `on_hand` and decrement `reserved` (commit stock).
- On `failed/cancelled`: decrement `reserved` (release stock).

## Tax & shipping (MVP rules)

Currency: **INR** (paise in storage/calculations).

- **Shipping**: flat rate per order (configurable, default `₹0` in dev).
- **Tax**: percentage rate (configurable, default `0%` in dev).
- **Tax base**: \(subtotal + shipping\).

Totals:

- `subtotal` = sum(line_item_unit_price * qty)
- `shipping_amount` = flat rate
- `tax_amount` = round((subtotal + shipping_amount) * tax_rate)
- `total` = subtotal + shipping_amount + tax_amount

MVP checkout requires:

- `customer_name`, `phone`, `email` (optional), `shipping_address` (single address block)

