// src/services/api.jsx
import { iamApi, mapleFileApi } from "./apiConfig";
import { authAPI } from "./authApi";
import { userAPI } from "./userApi";
import { collectionsAPI } from "./collectionApi";
import { fileAPI } from "./fileApi";

// Export all the APIs
export { iamApi, mapleFileApi, authAPI, userAPI, collectionsAPI, fileAPI };

// Default export for backward compatibility
export default iamApi;
