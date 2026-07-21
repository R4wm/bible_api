import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  base: "/v2/",
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      "/auth": "http://localhost:8000",
      "/bible": "http://localhost:8000",
      "/admin": "http://localhost:8000",
      "/user": "http://localhost:8000",
      "/health": "http://localhost:8000",
      "/docs": "http://localhost:8000"
    }
  }
});
