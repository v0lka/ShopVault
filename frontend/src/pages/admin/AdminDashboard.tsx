import React from "react";
import { Link } from "react-router-dom";

export default function AdminDashboard() {
  return (
    <div>
      <h3>Admin Dashboard</h3>
      <div className="row mt-4">
        <div className="col-md-4 mb-3">
          <div className="card">
            <div className="card-body">
              <h5>Products</h5>
              <p>Manage your product catalog</p>
              <Link to="/admin/products" className="btn btn-primary">Manage</Link>
            </div>
          </div>
        </div>
        <div className="col-md-4 mb-3">
          <div className="card">
            <div className="card-body">
              <h5>Orders</h5>
              <p>View and manage customer orders</p>
              <Link to="/admin/orders" className="btn btn-primary">View</Link>
            </div>
          </div>
        </div>
        <div className="col-md-4 mb-3">
          <div className="card">
            <div className="card-body">
              <h5>Users</h5>
              <p>Manage registered users</p>
              <Link to="/admin/users" className="btn btn-primary">Manage</Link>
            </div>
          </div>
        </div>
        <div className="col-md-4 mb-3">
          <div className="card">
            <div className="card-body">
              <h5>Import</h5>
              <p>Import products from external source</p>
              <Link to="/admin/import" className="btn btn-primary">Import</Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
