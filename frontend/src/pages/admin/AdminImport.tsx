import React, { useState } from "react";
import api from "../../api/client";

export default function AdminImport() {
  const [url, setUrl] = useState("");
  const [result, setResult] = useState("");
  const [loading, setLoading] = useState(false);

  const handleImport = async () => {
    if (!url.trim()) return;
    setLoading(true);
    setResult("");
    try {
      const res = await api.post("/admin/import", { url });
      setResult(`Imported ${res.data.products_imported} of ${res.data.products_found} products.`);
    } catch (err: any) {
      setResult(`Error: ${err.response?.data?.error || "Import failed"}`);
    }
    setLoading(false);
  };

  return (
    <div>
      <h3>Import Products</h3>
      <p className="text-muted">Import products from a JSON endpoint. The URL should return a JSON array of products.</p>
      <div className="row">
        <div className="col-md-6">
          <div className="input-group mb-3">
            <input
              type="text"
              className="form-control"
              placeholder="https://example.com/products.json"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
            />
            <button className="btn btn-primary" onClick={handleImport} disabled={loading}>
              {loading ? "Importing..." : "Import"}
            </button>
          </div>
          {result && (
            <div className={`alert ${result.startsWith("Error") ? "alert-danger" : "alert-success"}`}>
              {result}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
