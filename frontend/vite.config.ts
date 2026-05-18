import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      "/api": "http://backend:8080",
      "/uploads": "http://backend:8080",
    },
  },
  build: {
    sourcemap: true,
  },
});
