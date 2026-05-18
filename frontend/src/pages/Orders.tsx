import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import api from "../api/client";

interface OrderItem {
  id: number;
  product_id: number;
  quantity: number;
  price: number;
}

interface Order {
  id: number;
  total: number;
  status: string;
  shipping_address: string;
  cc_number: string;
  cc_expiry: string;
  cc_cvv: string;
  coupon_code: string;
  discount_percent: number;
  created_at: string;
  items: OrderItem[];
}

export default function Orders() {
  const [orders, setOrders] = useState<Order[]>([]);

  useEffect(() => {
    api.get("/orders").then((res) => setOrders(res.data.orders)).catch(() => {});
  }, []);

  return (
    <div className="container mt-4">
      <h3>My Orders</h3>
      {orders.length === 0 ? (
        <p>No orders yet. <Link to="/">Start shopping</Link></p>
      ) : (
        orders.map((order) => (
          <div key={order.id} className="card mb-3">
            <div className="card-header">
              <div className="row">
                <div className="col">
                  <strong>Order #{order.id}</strong>
                </div>
                <div className="col text-end">
                  <span className={`badge bg-${order.status === "delivered" ? "success" : "warning"}`}>
                    {order.status}
                  </span>
                </div>
              </div>
            </div>
            <div className="card-body">
              <p><strong>Date:</strong> {new Date(order.created_at).toLocaleString()}</p>
              <p><strong>Shipping:</strong> {order.shipping_address}</p>
              <p><strong>Payment:</strong> ****{order.cc_number.slice(-4)}</p>
              {order.coupon_code && (
                <p><strong>Coupon:</strong> {order.coupon_code} ({order.discount_percent}% off)</p>
              )}
              <p><strong>Total:</strong> ${order.total.toFixed(2)}</p>
              {order.items && order.items.length > 0 && (
                <div>
                  <strong>Items:</strong>
                  <ul>
                    {order.items.map((item) => (
                      <li key={item.id}>
                        Product #{item.product_id} x{item.quantity} - ${item.price.toFixed(2)}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </div>
        ))
      )}
    </div>
  );
}
