import React from "react";
import { useAuth } from "../context/AuthContext";

export default function Profile() {
  const { user } = useAuth();

  if (!user) return null;

  return (
    <div className="container mt-4">
      <div className="row justify-content-center">
        <div className="col-md-6">
          <h3>My Profile</h3>
          <div className="card">
            <div className="card-body">
              <p><strong>Name:</strong> {user.full_name}</p>
              <p><strong>Email:</strong> {user.email}</p>
              <p><strong>Role:</strong> <span className="badge bg-secondary">{user.role}</span></p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
