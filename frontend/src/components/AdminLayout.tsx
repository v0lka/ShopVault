import React from "react";
import { Link, Outlet, useLocation } from "react-router-dom";

export default function AdminLayout() {
  const location = useLocation();

  const navItems = [
    { path: "/admin", label: "Dashboard" },
    { path: "/admin/products", label: "Products" },
    { path: "/admin/orders", label: "Orders" },
    { path: "/admin/users", label: "Users" },
    { path: "/admin/import", label: "Import" },
  ];

  return (
    <div className="container-fluid mt-3">
      <div className="row">
        <div className="col-md-2">
          <h5 className="px-3">Admin Panel</h5>
          <ul className="nav flex-column">
            {navItems.map((item) => (
              <li key={item.path} className="nav-item">
                <Link
                  to={item.path}
                  className={`nav-link ${location.pathname === item.path ? "active" : ""}`}
                >
                  {item.label}
                </Link>
              </li>
            ))}
          </ul>
        </div>
        <div className="col-md-10">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
