import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Proxy API requests to your backend server
      "/maplefile/api": {
        target: "http://localhost:8000", // Update this to your actual backend URL
        changeOrigin: true,
        secure: false,
      },
      "/iam/api": {
        target: "http://localhost:8000", // Update this to your actual backend URL
        changeOrigin: true,
        secure: false,
      },
    },
  },
});
