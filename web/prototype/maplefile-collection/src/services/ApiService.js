/**
 * Base API service for handling HTTP requests
 * Follows dependency inversion principle - depends on abstractions, not concretions
 */
class ApiService {
  constructor(baseURL = null) {
    // In development, use relative URLs to leverage Vite proxy
    // In production, use environment variable or fallback
    if (baseURL) {
      this.baseURL = baseURL;
      console.log("ApiService: Using provided baseURL:", baseURL);
    } else if (import.meta.env.DEV) {
      // Development: use relative URL that will be proxied by Vite
      this.baseURL = "/maplefile/api/v1";
      console.log("ApiService: Using DEV relative baseURL:", this.baseURL);
    } else {
      // Production: use environment variable or default
      this.baseURL = import.meta.env.VITE_API_BASE_URL || "/maplefile/api/v1";
      console.log("ApiService: Using PROD baseURL:", this.baseURL);
    }

    console.log("ApiService initialized with final baseURL:", this.baseURL);
    console.log("Environment check - DEV mode:", import.meta.env.DEV);
    console.log(
      "Environment check - VITE_API_BASE_URL:",
      import.meta.env.VITE_API_BASE_URL,
    );
  }

  /**
   * Makes HTTP requests with proper error handling
   * @param {string} endpoint
   * @param {object} options
   * @returns {Promise<object>}
   */
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;

    const config = {
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
      ...options,
    };

    console.log("🚀 ApiService.request() called:");
    console.log("  - baseURL:", this.baseURL);
    console.log("  - endpoint:", endpoint);
    console.log("  - final URL:", url);
    console.log("  - method:", config.method || "GET");

    try {
      const response = await fetch(url, config);

      console.log("✅ Response received:", response.status, "for URL:", url);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        console.error("❌ API Error:", response.status, errorData);

        throw new ApiError(
          response.status,
          errorData.error?.message || response.statusText,
          errorData.error?.details || {},
        );
      }

      // Handle empty responses (like DELETE operations)
      const contentType = response.headers.get("content-type");
      if (contentType && contentType.includes("application/json")) {
        return await response.json();
      }

      return {};
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }

      // Network or other errors
      console.error("🔥 Network error for URL:", url, "Error:", error);
      throw new ApiError(0, "Network error or request failed", {
        originalError: error.message,
      });
    }
  }

  /**
   * GET request
   */
  async get(endpoint, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const url = queryString ? `${endpoint}?${queryString}` : endpoint;

    return this.request(url, {
      method: "GET",
    });
  }

  /**
   * POST request
   */
  async post(endpoint, data = {}) {
    return this.request(endpoint, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  /**
   * PUT request
   */
  async put(endpoint, data = {}) {
    return this.request(endpoint, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  /**
   * DELETE request
   */
  async delete(endpoint) {
    return this.request(endpoint, {
      method: "DELETE",
    });
  }

  /**
   * Health check endpoint to test API connectivity
   */
  async healthCheck() {
    try {
      // Try a simple GET request to test connectivity
      const response = await fetch(`${this.baseURL}/collections`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      console.log("Health check response status:", response.status);
      return {
        status: response.status,
        ok: response.ok,
        baseURL: this.baseURL,
      };
    } catch (error) {
      console.error("Health check failed:", error);
      return {
        status: 0,
        ok: false,
        error: error.message,
        baseURL: this.baseURL,
      };
    }
  }
}

/**
 * Custom error class for API errors
 */
class ApiError extends Error {
  constructor(status, message, details = {}) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.details = details;
  }
}

export default ApiService;
export { ApiError };
