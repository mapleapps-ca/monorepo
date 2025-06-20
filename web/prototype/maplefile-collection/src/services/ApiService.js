/**
 * Base API service for handling HTTP requests
 * Follows dependency inversion principle - depends on abstractions, not concretions
 */
class ApiService {
  constructor(
    baseURL = import.meta.env.VITE_API_BASE_URL || "/maplefile/api/v1",
  ) {
    this.baseURL = baseURL;
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

    try {
      const response = await fetch(url, config);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
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
