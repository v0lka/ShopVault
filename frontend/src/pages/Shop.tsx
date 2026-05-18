import React, { useState, useEffect } from "react";
import { Link, useSearchParams } from "react-router-dom";
import api from "../api/client";
import { useCart } from "../context/CartContext";

interface Product {
  id: number;
  name: string;
  description: string;
  price: number;
  image_url: string;
  category: string;
  stock: number;
}

export default function Shop() {
  const [products, setProducts] = useState<Product[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<Product[] | null>(null);
  const { addItem } = useCart();
  const [searchParams] = useSearchParams();

  useEffect(() => {
    api.get("/products").then((res) => setProducts(res.data.products)).catch(() => {});
  }, []);

  useEffect(() => {
    const hash = window.location.hash.slice(1);
    if (hash) {
      const el = document.getElementById("preview-area");
      if (el) {
        el.innerHTML = hash;
      }
    }
  }, []);

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      setSearchResults(null);
      return;
    }
    try {
      const res = await api.get("/products/search", { params: { q: searchQuery } });
      setSearchResults(res.data.products);
    } catch {
      setSearchResults([]);
    }
  };

  const displayedProducts = searchResults !== null ? searchResults : products;

  const handleAddToCart = (p: Product) => {
    addItem({
      product_id: p.id,
      name: p.name,
      price: p.price,
      quantity: 1,
    });
  };

  return (
    <div className="container mt-4">
      <div className="row mb-4">
        <div className="col-md-6">
          <div className="input-group">
            <input
              type="text"
              className="form-control"
              placeholder="Search products..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSearch()}
            />
            <button className="btn btn-primary" onClick={handleSearch}>
              Search
            </button>
          </div>
        </div>
      </div>

      {searchParams.get("q") && (
        <div className="alert alert-info">
          Search results for: <span dangerouslySetInnerHTML={{ __html: searchParams.get("q") || "" }} />
        </div>
      )}

      <div id="preview-area"></div>

      <div className="row">
        {displayedProducts.map((p) => (
          <div key={p.id} className="col-md-4 mb-4">
            <div className="card h-100">
              <div className="card-body">
                <h5 className="card-title">
                  <Link to={`/product/${p.id}`}>{p.name}</Link>
                </h5>
                <p className="card-text text-truncate">{p.description}</p>
                <p className="card-text">
                  <strong>${p.price.toFixed(2)}</strong>
                </p>
                <p className="card-text">
                  <small className="text-muted">{p.category}</small>
                </p>
                <button
                  className="btn btn-success"
                  onClick={() => handleAddToCart(p)}
                >
                  Add to Cart
                </button>
              </div>
            </div>
          </div>
        ))}
        {displayedProducts.length === 0 && (
          <div className="col-12">
            <p className="text-center text-muted">No products found.</p>
          </div>
        )}
      </div>
    </div>
  );
}
