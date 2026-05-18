import React, { useState, useEffect } from "react";
import api from "../../api/client";

interface Product {
  id: number;
  name: string;
  description: string;
  price: number;
  image_url: string;
  category: string;
  stock: number;
}

export default function AdminProducts() {
  const [products, setProducts] = useState<Product[]>([]);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState({ name: "", description: "", price: 0, category: "", stock: 0, image_url: "" });

  const loadProducts = () => {
    api.get("/admin/products").then((res) => setProducts(res.data.products)).catch(() => {});
  };

  useEffect(() => { loadProducts(); }, []);

  const handleCreate = async () => {
    try {
      await api.post("/admin/products", form);
      setForm({ name: "", description: "", price: 0, category: "", stock: 0, image_url: "" });
      loadProducts();
    } catch (err: any) {
      alert(err.response?.data?.error || "Failed to create product");
    }
  };

  const handleUpdate = async (id: number) => {
    try {
      await api.put(`/admin/products/${id}`, form);
      setEditingId(null);
      loadProducts();
    } catch (err: any) {
      alert(err.response?.data?.error || "Failed to update product");
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Are you sure?")) return;
    try {
      await api.delete(`/admin/products/${id}`);
      loadProducts();
    } catch (err: any) {
      alert(err.response?.data?.error || "Failed to delete product");
    }
  };

  const startEdit = (p: Product) => {
    setEditingId(p.id);
    setForm({ name: p.name, description: p.description, price: p.price, category: p.category, stock: p.stock, image_url: p.image_url });
  };

  return (
    <div>
      <h3>Products</h3>

      <div className="card mb-4">
        <div className="card-body">
          <h5>{editingId ? "Edit Product" : "Add Product"}</h5>
          <div className="row g-2">
            <div className="col-md-3">
              <input className="form-control" placeholder="Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
            </div>
            <div className="col-md-3">
              <input className="form-control" placeholder="Description" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
            </div>
            <div className="col-md-2">
              <input type="number" className="form-control" placeholder="Price" value={form.price || ""} onChange={(e) => setForm({ ...form, price: parseFloat(e.target.value) || 0 })} />
            </div>
            <div className="col-md-2">
              <input className="form-control" placeholder="Category" value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })} />
            </div>
            <div className="col-md-1">
              <input type="number" className="form-control" placeholder="Stock" value={form.stock || ""} onChange={(e) => setForm({ ...form, stock: parseInt(e.target.value) || 0 })} />
            </div>
            <div className="col-md-1">
              {editingId ? (
                <button className="btn btn-warning btn-sm" onClick={() => handleUpdate(editingId)}>Update</button>
              ) : (
                <button className="btn btn-success btn-sm" onClick={handleCreate}>Add</button>
              )}
              {editingId && (
                <button className="btn btn-secondary btn-sm ms-1" onClick={() => { setEditingId(null); setForm({ name: "", description: "", price: 0, category: "", stock: 0, image_url: "" }); }}>Cancel</button>
              )}
            </div>
          </div>
        </div>
      </div>

      <table className="table table-striped">
        <thead>
          <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Price</th>
            <th>Category</th>
            <th>Stock</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {products.map((p) => (
            <tr key={p.id}>
              <td>{p.id}</td>
              <td>{p.name}</td>
              <td>${p.price.toFixed(2)}</td>
              <td>{p.category}</td>
              <td>{p.stock}</td>
              <td>
                <button className="btn btn-sm btn-outline-primary me-1" onClick={() => startEdit(p)}>Edit</button>
                <button className="btn btn-sm btn-outline-danger" onClick={() => handleDelete(p.id)}>Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
