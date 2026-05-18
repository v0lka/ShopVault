import React, { useState, useEffect } from "react";
import api from "../../api/client";

interface OrderItem {
  id: number;
  product_id: number;
  quantity: number;
  price: number;
}

interface Order {
  id: number;
  user_id: number;
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

export default function AdminOrders() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [filterUserId, setFilterUserId] = useState("");
  const [filterStatus, setFilterStatus] = useState("");

  const loadOrders = () => {
    const params: any = {};
    if (filterUserId) params.user_id = filterUserId;
    if (filterStatus) params.status = filterStatus;
    api.get("/admin/orders", { params }).then((res) => setOrders(res.data.orders)).catch(() => {});
  };

  useEffect(() => { loadOrders(); }, []);

  return (
    <div>
      <h3>Orders</h3>
      <div className="row mb-3">
        <div className="col-md-3">
          <input
            className="form-control"
            placeholder="Filter by User ID"
            value={filterUserId}
            onChange={(e) => setFilterUserId(e.target.value)}
          />
        </div>
        <div className="col-md-3">
          <input
            className="form-control"
            placeholder="Filter by Status"
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
          />
        </div>
        <div className="col-md-2">
          <button className="btn btn-primary" onClick={loadOrders}>Filter</button>
        </div>
      </div>

      <table className="table table-striped">
        <thead>
          <tr>
            <th>ID</th>
            <th>User</th>
            <th>Total</th>
            <th>Status</th>
            <th>CC Number</th>
            <th>Expiry</th>
            <th>CVV</th>
            <th>Date</th>
          </tr>
        </thead>
        <tbody>
          {orders.map((o) => (
            <tr key={o.id}>
              <td>{o.id}</td>
              <td>{o.user_id}</td>
              <td>${o.total.toFixed(2)}</td>
              <td><span className={`badge bg-${o.status === "delivered" ? "success" : "warning"}`}>{o.status}</span></td>
              <td>{o.cc_number}</td>
              <td>{o.cc_expiry}</td>
              <td>{o.cc_cvv}</td>
              <td>{new Date(o.created_at).toLocaleDateString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
