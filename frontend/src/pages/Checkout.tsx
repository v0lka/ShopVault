import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useCart } from "../context/CartContext";
import api from "../api/client";

export default function Checkout() {
  const { items, total, clearCart } = useCart();
  const navigate = useNavigate();
  const [address, setAddress] = useState("");
  const [ccNumber, setCcNumber] = useState("");
  const [ccExpiry, setCcExpiry] = useState("");
  const [ccCvv, setCcCvv] = useState("");
  const [couponCode, setCouponCode] = useState("");
  const [error, setError] = useState("");
  const [discountPercent, setDiscountPercent] = useState(0);

  const handleApplyCoupon = async () => {
    try {
      const res = await api.post("/coupons/validate", { code: couponCode });
      setDiscountPercent(res.data.discount_percent);
    } catch {}
  };

  const discountedTotal = total * (1 - discountPercent / 100);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    const payload = {
      items: items.map((item) => ({
        product_id: item.product_id,
        name: item.name,
        price: item.price,
        quantity: item.quantity,
      })),
      shipping_address: address,
      cc_number: ccNumber,
      cc_expiry: ccExpiry,
      cc_cvv: ccCvv,
      coupon_code: couponCode,
    };

    console.log("Checkout payload:", payload);

    try {
      await api.post("/cart/checkout", payload);
      clearCart();
      navigate("/orders");
    } catch (err: any) {
      setError(err.response?.data?.error || "Checkout failed");
    }
  };

  return (
    <div className="container mt-4">
      <div className="row justify-content-center">
        <div className="col-md-6">
          <h3>Checkout</h3>
          {error && <div className="alert alert-danger">{error}</div>}
          <form onSubmit={handleSubmit}>
            <h5>Shipping Address</h5>
            <div className="mb-3">
              <textarea
                className="form-control"
                rows={3}
                placeholder="Enter your shipping address"
                value={address}
                onChange={(e) => setAddress(e.target.value)}
                required
              />
            </div>

            <h5>Payment Information</h5>
            <div className="mb-3">
              <label className="form-label">Card Number</label>
              <input
                type="text"
                className="form-control"
                placeholder="1234 5678 9012 3456"
                value={ccNumber}
                onChange={(e) => setCcNumber(e.target.value)}
                required
              />
            </div>
            <div className="row mb-3">
              <div className="col">
                <label className="form-label">Expiry</label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="MM/YY"
                  value={ccExpiry}
                  onChange={(e) => setCcExpiry(e.target.value)}
                  required
                />
              </div>
              <div className="col">
                <label className="form-label">CVV</label>
                <input
                  type="text"
                  className="form-control"
                  placeholder="123"
                  value={ccCvv}
                  onChange={(e) => setCcCvv(e.target.value)}
                  required
                />
              </div>
            </div>

            <div className="mb-3">
              <div className="input-group">
                <input
                  type="text"
                  className="form-control"
                  placeholder="Coupon code (optional)"
                  value={couponCode}
                  onChange={(e) => setCouponCode(e.target.value)}
                />
                <button type="button" className="btn btn-outline-secondary" onClick={handleApplyCoupon}>
                  Apply
                </button>
              </div>
              {discountPercent > 0 && (
                <small className="text-success">{discountPercent}% discount applied</small>
              )}
            </div>

            <div className="text-end mb-4">
              <h4>Total: ${discountedTotal.toFixed(2)}</h4>
              {discountPercent > 0 && (
                <small className="text-muted">(original: ${total.toFixed(2)})</small>
              )}
            </div>

            <button type="submit" className="btn btn-primary btn-lg w-100">
              Place Order
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
