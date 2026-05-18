import React, { useState } from "react";
import { Link } from "react-router-dom";
import { useCart } from "../context/CartContext";
import { useAuth } from "../context/AuthContext";
import api from "../api/client";

export default function Cart() {
  const { items, removeItem, updateQuantity, total, clearCart } = useCart();
  const { user } = useAuth();
  const [couponCode, setCouponCode] = useState("");
  const [couponDiscount, setCouponDiscount] = useState(0);
  const [couponError, setCouponError] = useState("");

  const handleApplyCoupon = async () => {
    setCouponError("");
    try {
      const res = await api.post("/coupons/validate", { code: couponCode });
      setCouponDiscount(res.data.discount_percent);
    } catch (err: any) {
      setCouponError(err.response?.data?.error || "Invalid coupon");
      setCouponDiscount(0);
    }
  };

  const discountedTotal = total * (1 - couponDiscount / 100);

  return (
    <div className="container mt-4">
      <h3>Shopping Cart</h3>
      {items.length === 0 ? (
        <p>Your cart is empty. <Link to="/">Continue shopping</Link></p>
      ) : (
        <>
          <table className="table">
            <thead>
              <tr>
                <th>Product</th>
                <th>Price</th>
                <th>Quantity</th>
                <th>Subtotal</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => (
                <tr key={item.product_id}>
                  <td>
                    <Link to={`/product/${item.product_id}`}>{item.name}</Link>
                  </td>
                  <td>${item.price.toFixed(2)}</td>
                  <td>
                    <input
                      type="number"
                      className="form-control form-control-sm"
                      style={{ width: "80px" }}
                      value={item.quantity}
                      onChange={(e) =>
                        updateQuantity(item.product_id, parseInt(e.target.value) || 0)
                      }
                      min="0"
                    />
                  </td>
                  <td>${(item.price * item.quantity).toFixed(2)}</td>
                  <td>
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => removeItem(item.product_id)}
                    >
                      Remove
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <div className="row mb-3">
            <div className="col-md-4">
              <div className="input-group">
                <input
                  type="text"
                  className="form-control"
                  placeholder="Coupon code"
                  value={couponCode}
                  onChange={(e) => setCouponCode(e.target.value)}
                />
                <button className="btn btn-outline-secondary" onClick={handleApplyCoupon}>
                  Apply
                </button>
              </div>
              {couponError && <small className="text-danger">{couponError}</small>}
              {couponDiscount > 0 && (
                <small className="text-success">{couponDiscount}% discount applied</small>
              )}
            </div>
            <div className="col-md-4 offset-md-4 text-end">
              <h5>Total: ${discountedTotal.toFixed(2)}</h5>
              {couponDiscount > 0 && (
                <small className="text-muted">(was ${total.toFixed(2)})</small>
              )}
            </div>
          </div>

          <div className="text-end">
            {user ? (
              <Link to="/checkout" className="btn btn-primary btn-lg">
                Proceed to Checkout
              </Link>
            ) : (
              <Link to="/login" className="btn btn-primary btn-lg">
                Login to Checkout
              </Link>
            )}
          </div>
        </>
      )}
    </div>
  );
}
