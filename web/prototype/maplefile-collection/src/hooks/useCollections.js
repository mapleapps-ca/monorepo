import { useState, useEffect } from "react";
import { useServices } from "../contexts/ServiceContext";

/**
 * Custom hook for managing collections state
 * Follows Single Responsibility Principle - only handles collection state management
 */
export const useCollections = (filters = {}) => {
  const { collectionService } = useServices();
  const [collections, setCollections] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchCollections = async () => {
    try {
      setLoading(true);
      setError(null);

      let response;
      if (
        filters.include_shared !== undefined ||
        filters.include_owned !== undefined
      ) {
        response = await collectionService.getFilteredCollections(filters);
        // Combine owned and shared collections
        const allCollections = [
          ...(response.owned_collections || []),
          ...(response.shared_collections || []),
        ];
        setCollections(allCollections);
      } else {
        response = await collectionService.getUserCollections();
        setCollections(response.collections || []);
      }
    } catch (err) {
      setError(err.message || "Failed to fetch collections");
      setCollections([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCollections();
  }, [filters.include_owned, filters.include_shared]);

  const refreshCollections = () => {
    fetchCollections();
  };

  return {
    collections,
    loading,
    error,
    refreshCollections,
  };
};

/**
 * Custom hook for managing a single collection
 */
export const useCollection = (collectionId) => {
  const { collectionService } = useServices();
  const [collection, setCollection] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchCollection = async () => {
    if (!collectionId) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const response = await collectionService.getCollection(collectionId);
      setCollection(response);
    } catch (err) {
      setError(err.message || "Failed to fetch collection");
      setCollection(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCollection();
  }, [collectionId]);

  const refreshCollection = () => {
    fetchCollection();
  };

  return {
    collection,
    loading,
    error,
    refreshCollection,
  };
};

/**
 * Custom hook for collection CRUD operations
 */
export const useCollectionOperations = () => {
  const { collectionService, cryptoService } = useServices();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const createCollection = async (collectionData) => {
    try {
      setLoading(true);
      setError(null);

      // Encrypt the collection name
      const encryptedName = cryptoService.encrypt(collectionData.name);

      // Generate collection key
      const collectionKey = cryptoService.generateCollectionKey();

      const payload = {
        encrypted_name: encryptedName,
        collection_type: collectionData.type || "folder",
        encrypted_collection_key: collectionKey,
        parent_id: collectionData.parent_id || null,
      };

      const response = await collectionService.createCollection(payload);
      return response;
    } catch (err) {
      setError(err.message || "Failed to create collection");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const updateCollection = async (collectionId, updateData) => {
    try {
      setLoading(true);
      setError(null);

      // Encrypt the new name if provided
      const payload = { ...updateData };
      if (updateData.name) {
        payload.encrypted_name = cryptoService.encrypt(updateData.name);
      }

      const response = await collectionService.updateCollection(
        collectionId,
        payload,
      );
      return response;
    } catch (err) {
      setError(err.message || "Failed to update collection");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const deleteCollection = async (collectionId) => {
    try {
      setLoading(true);
      setError(null);

      const response = await collectionService.deleteCollection(collectionId);
      return response;
    } catch (err) {
      setError(err.message || "Failed to delete collection");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const shareCollection = async (collectionId, shareData) => {
    try {
      setLoading(true);
      setError(null);

      // In a real app, you'd encrypt the collection key for the recipient
      const encryptedKey = cryptoService.encryptKeyForRecipient(
        shareData.collectionKey,
        shareData.recipientPublicKey,
      );

      const payload = {
        ...shareData,
        encrypted_collection_key: encryptedKey,
      };

      const response = await collectionService.shareCollection(
        collectionId,
        payload,
      );
      return response;
    } catch (err) {
      setError(err.message || "Failed to share collection");
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return {
    loading,
    error,
    createCollection,
    updateCollection,
    deleteCollection,
    shareCollection,
    clearError: () => setError(null),
  };
};
