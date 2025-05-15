// src/services/collectionApi.js
import { mapleFileApi } from "./apiConfig";
import { cryptoUtils } from "../utils/crypto";

export const collectionsAPI = {
  listCollections: async () => {
    const response = await mapleFileApi.get("/collections");
    return response.data;
  },

  getCollection: async (collectionId, masterKey) => {
    const response = await mapleFileApi.get(`/collections/${collectionId}`);
    const collectionData = response.data;

    if (
      masterKey &&
      collectionData &&
      collectionData.encrypted_collection_key
    ) {
      try {
        const ciphertext = await cryptoUtils.fromBase64(
          collectionData.encrypted_collection_key.ciphertext,
        );
        const nonce = await cryptoUtils.fromBase64(
          collectionData.encrypted_collection_key.nonce,
        );

        console.log("Attempting to decrypt collection key for:", collectionId);
        const decryptedCollectionKey = await cryptoUtils.decryptWithKey(
          ciphertext,
          nonce,
          masterKey,
        );

        collectionData.decryptedCollectionKey = decryptedCollectionKey;
        console.log("Collection key decrypted successfully for:", collectionId);
      } catch (err) {
        console.error(
          `Failed to decrypt collection key for ${collectionId}:`,
          err,
        );
        collectionData.decryptionError =
          "Failed to decrypt collection key. Check master key or data integrity.";
      }
    } else if (
      masterKey &&
      collectionData &&
      !collectionData.encrypted_collection_key
    ) {
      console.warn(
        `Collection ${collectionId} fetched but has no encrypted_collection_key field.`,
      );
      collectionData.decryptionError =
        "Collection key data is missing from server response.";
    }
    return collectionData;
  },

  createCollection: async (name, path, type = "folder", masterKey) => {
    if (!masterKey)
      throw new Error("Master key is required to create a collection.");

    const collectionKey = await cryptoUtils.generateRandomBytes(32); // sodium.crypto_secretbox_KEYBYTES
    const { nonce, ciphertext } = await cryptoUtils.encryptWithKey(
      collectionKey,
      masterKey,
    );

    const payload = {
      name, // In a full E2EE system, name/path might also be encrypted, perhaps with collectionKey
      path,
      type,
      encrypted_collection_key: {
        ciphertext: await cryptoUtils.toBase64(ciphertext),
        nonce: await cryptoUtils.toBase64(nonce),
      },
    };
    const response = await mapleFileApi.post("/collections", payload);
    // Optionally, attach the decrypted collectionKey to the response for immediate use by the caller
    // response.data.decryptedCollectionKey = collectionKey;
    return response.data;
  },

  updateCollection: async (collectionId, updates, masterKey) => {
    // If updates include re-keying or encrypting name/path, masterKey/collectionKey would be needed
    const response = await mapleFileApi.put(
      `/collections/${collectionId}`,
      updates,
    );
    return response.data;
  },

  deleteCollection: async (collectionId) => {
    const response = await mapleFileApi.delete(`/collections/${collectionId}`);
    return response.data;
  },

  listFiles: async (collectionId) => {
    const response = await mapleFileApi.get(
      `/collections/${collectionId}/files`,
    );
    return response.data;
  },
};

export default collectionsAPI;
