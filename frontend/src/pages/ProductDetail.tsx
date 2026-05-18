import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import api from "../api/client";
import { useAuth } from "../context/AuthContext";
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

interface Review {
  id: number;
  rating: number;
  comment: string;
  created_at: string;
  user_name: string;
}

export default function ProductDetail() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const { addItem } = useCart();
  const [product, setProduct] = useState<Product | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [rating, setRating] = useState(5);
  const [comment, setComment] = useState("");

  useEffect(() => {
    api.get(`/products/${id}`).then((res) => {
      setProduct(res.data.product);
      setReviews(res.data.reviews || []);
    }).catch(() => {});
  }, [id]);

  const handleAddToCart = () => {
    if (product) {
      addItem({
        product_id: product.id,
        name: product.name,
        price: product.price,
        quantity: 1,
      });
    }
  };

  const handleSubmitReview = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!comment.trim()) return;
    try {
      const res = await api.post("/reviews", {
        product_id: Number(id),
        rating,
        comment,
      });
      setReviews((prev) => [
        {
          id: Date.now(),
          rating,
          comment,
          created_at: new Date().toISOString(),
          user_name: user?.full_name || "Anonymous",
        },
        ...prev,
      ]);
      setComment("");
    } catch {
      alert("Failed to submit review");
    }
  };

  if (!product) {
    return <div className="container mt-4"><p>Loading...</p></div>;
  }

  return (
    <div className="container mt-4">
      <div className="row">
        <div className="col-md-8">
          <h2>{product.name}</h2>
          <p>{product.description}</p>
          <h3 className="text-primary">${product.price.toFixed(2)}</h3>
          <p>Category: {product.category}</p>
          <p>In stock: {product.stock}</p>
          <button className="btn btn-success btn-lg" onClick={handleAddToCart}>
            Add to Cart
          </button>
        </div>
      </div>

      <hr className="my-4" />

      <h4>Reviews</h4>
      {reviews.map((review) => (
        <div key={review.id} className="card mb-2">
          <div className="card-body">
            <h6>
              {review.user_name} - {"★".repeat(review.rating)}{"☆".repeat(5 - review.rating)}
            </h6>
            <p dangerouslySetInnerHTML={{ __html: review.comment }} />
            <small className="text-muted">{new Date(review.created_at).toLocaleDateString()}</small>
          </div>
        </div>
      ))}
      {reviews.length === 0 && (
        <p className="text-muted">No reviews yet. Be the first!</p>
      )}

      {user && (
        <div className="mt-4">
          <h5>Write a Review</h5>
          <form onSubmit={handleSubmitReview}>
            <div className="mb-2">
              <label className="form-label">Rating</label>
              <select
                className="form-select w-auto"
                value={rating}
                onChange={(e) => setRating(Number(e.target.value))}
              >
                {[5, 4, 3, 2, 1].map((v) => (
                  <option key={v} value={v}>
                    {v} stars
                  </option>
                ))}
              </select>
            </div>
            <div className="mb-2">
              <textarea
                className="form-control"
                rows={3}
                placeholder="Share your experience..."
                value={comment}
                onChange={(e) => setComment(e.target.value)}
              />
            </div>
            <button type="submit" className="btn btn-primary">
              Submit Review
            </button>
          </form>
        </div>
      )}
    </div>
  );
}
