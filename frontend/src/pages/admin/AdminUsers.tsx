import React, { useState, useEffect } from "react";
import api from "../../api/client";

interface User {
  id: number;
  email: string;
  password_hash: string;
  full_name: string;
  role: string;
  reset_token: string;
  created_at: string;
}

export default function AdminUsers() {
  const [users, setUsers] = useState<User[]>([]);

  useEffect(() => {
    api.get("/admin/users").then((res) => setUsers(res.data.users)).catch(() => {});
  }, []);

  return (
    <div>
      <h3>Users</h3>
      <table className="table table-striped">
        <thead>
          <tr>
            <th>ID</th>
            <th>Email</th>
            <th>Name</th>
            <th>Role</th>
            <th>Password Hash</th>
            <th>Reset Token</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {users.map((u) => (
            <tr key={u.id}>
              <td>{u.id}</td>
              <td>{u.email}</td>
              <td>{u.full_name}</td>
              <td><span className="badge bg-secondary">{u.role}</span></td>
              <td><code style={{ fontSize: "0.75rem" }}>{u.password_hash}</code></td>
              <td><code style={{ fontSize: "0.7rem" }}>{u.reset_token || "-"}</code></td>
              <td>{new Date(u.created_at).toLocaleDateString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
