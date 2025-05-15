// src/services/fileApi.js
import { mapleFileApi } from "./apiConfig";
import { cryptoUtils, initSodium } from "../utils/crypto"; // Import initSodium
import { collectionsAPI } from "./collectionApi"; // Import collectionsAPI
import _sodium from "libsodium-wrappers-sumo";

// File API functions
export const fileAPI = {
  listFiles: async (collectionId) => {
    const response = await mapleFileApi.get(
      `/collections/${collectionId}/files`,
    );
    return response.data;
  },

  getFile: async (fileId) => {
    const response = await mapleFileApi.get(`/files/${fileId}`);
    console.log(
      `[fileAPI.getFile metadata for ${fileId}]:`,
      JSON.stringify(response.data, null, 2),
    );
    return response.data; // This is FileResponseDTO
  },

  storeEncryptedFileData: async (fileId, encryptedDataWithNonce) => {
    const blob = new Blob([encryptedDataWithNonce], {
      type: "application/octet-stream",
    });
    const formData = new FormData();
    formData.append("file", blob, "encrypted_file.bin"); // Filename is illustrative
    // fileId here is the server's metadata ID for the file
    const response = await mapleFileApi.post(`/files/${fileId}/data`, formData);
    return response.data;
  },

  getEncryptedFileData: async (fileId /* server metadata ID */) => {
    const response = await mapleFileApi.get(`/files/${fileId}/data`, {
      responseType: "blob",
    });
    return response.data; // This is a Blob
  },

  deleteFile: async (fileId) => {
    const response = await mapleFileApi.delete(`/files/${fileId}`);
    return response.data;
  },

  uploadFile: async (
    fileInstance,
    collectionId,
    decryptedCollectionKey,
    onProgress,
  ) => {
    const sodium = await initSodium(); // Use initSodium
    onProgress(5);

    // 1. Generate File Key
    const fileKey = await cryptoUtils.generateRandomBytes(
      sodium.crypto_secretbox_KEYBYTES,
    );
    onProgress(10);

    // 2. Encrypt File Key with (decrypted) Collection Key
    const { nonce: efkNonce, ciphertext: efkCiphertext } =
      await cryptoUtils.encryptWithKey(fileKey, decryptedCollectionKey);
    onProgress(20);

    // 3. Read file content
    const fileContentBytes = new Uint8Array(await fileInstance.arrayBuffer());
    onProgress(30);

    // 4. Encrypt File Content with File Key
    const { nonce: contentNonce, ciphertext: contentCiphertext } =
      await cryptoUtils.encryptWithKey(fileContentBytes, fileKey);
    const combinedEncryptedContent = cryptoUtils.combineNonceAndCiphertext(
      contentNonce,
      contentCiphertext,
    );
    onProgress(50);

    // 5. Prepare and Encrypt Metadata
    const metadata = {
      name: fileInstance.name,
      type: fileInstance.type || "application/octet-stream",
      original_size: fileInstance.size,
      lastModified: fileInstance.lastModified,
    };
    const metadataString = JSON.stringify(metadata);
    const metadataBytes = await cryptoUtils.stringToBytes(metadataString);
    const { nonce: metaNonce, ciphertext: metaCiphertext } =
      await cryptoUtils.encryptWithKey(metadataBytes, fileKey);
    const combinedEncryptedMetadata = cryptoUtils.combineNonceAndCiphertext(
      metaNonce,
      metaCiphertext,
    );
    const base64EncryptedMetadata = await cryptoUtils.toBase64(
      combinedEncryptedMetadata,
    );
    onProgress(60);

    const clientSideFileId = await cryptoUtils.toBase64(
      await cryptoUtils.generateRandomBytes(16),
    );
    const encryptedContentHashBytes = sodium.crypto_generichash(
      32,
      combinedEncryptedContent,
    );
    const encryptedContentHash = await cryptoUtils.toBase64(
      encryptedContentHashBytes,
    );

    // 6. Prepare payload for creating file metadata record (POST /files)
    // This corresponds to CreateFileRequestDTO in the backend service
    const createFilePayload = {
      collection_id: collectionId,
      file_id: clientSideFileId, // This is the client-generated ID for the content
      encrypted_size: combinedEncryptedContent.length,
      encrypted_original_size: metadata.original_size.toString(),
      encrypted_metadata: base64EncryptedMetadata, // base64(nonce || encrypted_metadata_json_bytes)
      encrypted_file_key: {
        // Corresponds to keys.EncryptedFileKey on backend
        ciphertext: await cryptoUtils.toBase64(efkCiphertext),
        nonce: await cryptoUtils.toBase64(efkNonce),
      },
      encryption_version: "1.0",
      encrypted_hash: encryptedContentHash, // Hash of the combined (contentNonce + encryptedContentCiphertext)
      // encrypted_thumbnail could be added here if generated and encrypted
    };
    onProgress(70);

    // 7. API Call: Create file metadata record
    const createResponse = await mapleFileApi.post("/files", createFilePayload);
    const createdFileRecord = createResponse.data; // This is FileResponseDTO from backend
    onProgress(80);

    // 8. API Call: Upload encrypted file content (POST /files/{SERVER_METADATA_ID}/data)
    // The ID from createdFileRecord is the MongoDB ObjectID of the metadata document.
    await fileAPI.storeEncryptedFileData(
      createdFileRecord.id,
      combinedEncryptedContent,
    );
    onProgress(100);

    return createdFileRecord;
  },

  downloadFile: async (fileId /* server metadata ID */, masterKey) => {
    const sodium = await initSodium(); // Use initSodium
    console.log(`E2EE Download: Starting for server file ID: ${fileId}`);

    // 1. Get File Metadata (FileResponseDTO)
    const fileRecord = await fileAPI.getFile(fileId);
    if (!fileRecord) throw new Error("File metadata not found.");

    // 2. Get Collection Data and Decrypt Collection Key
    const collectionData = await collectionsAPI.getCollection(
      fileRecord.collection_id,
      masterKey,
    );
    if (!collectionData || !collectionData.decryptedCollectionKey) {
      throw new Error(
        `Could not retrieve or decrypt collection key for collection ${fileRecord.collection_id}. ${collectionData?.decryptionError || ""}`,
      );
    }
    const decryptedCollectionKey = collectionData.decryptedCollectionKey;

    // 3. Decrypt File Key
    // encrypted_file_key: { ciphertext: base64String, nonce: base64String }
    const efkCiphertext = await cryptoUtils.fromBase64(
      fileRecord.encrypted_file_key.ciphertext,
    );
    const efkNonce = await cryptoUtils.fromBase64(
      fileRecord.encrypted_file_key.nonce,
    );
    const decryptedFileKey = await cryptoUtils.decryptWithKey(
      efkCiphertext,
      efkNonce,
      decryptedCollectionKey,
    );
    console.log("E2EE Download: File key decrypted");

    // 4. Download Encrypted File Content (Blob of nonce || ciphertext)
    const encryptedContentBlob = await fileAPI.getEncryptedFileData(
      fileRecord.id,
    );
    const encryptedContentWithNonce = new Uint8Array(
      await encryptedContentBlob.arrayBuffer(),
    );

    // 5. Split Nonce and Ciphertext for File Content
    const { nonce: contentNonce, ciphertext: contentCiphertext } =
      await cryptoUtils.splitNonceAndCiphertext(encryptedContentWithNonce);

    // 6. Decrypt File Content
    const decryptedFileContent = await cryptoUtils.decryptWithKey(
      contentCiphertext,
      contentNonce,
      decryptedFileKey,
    );
    console.log(
      "E2EE Download: File content decrypted",
      decryptedFileContent.length,
      "bytes",
    );

    // 7. Decrypt Metadata
    let originalMetadata = {
      name: `file_${fileId.slice(-6)}.dat`,
      type: "application/octet-stream",
    };
    if (fileRecord.encrypted_metadata) {
      try {
        const combinedEncMetadata = await cryptoUtils.fromBase64(
          fileRecord.encrypted_metadata,
        );
        const { nonce: metaNonce, ciphertext: metaCiphertext } =
          await cryptoUtils.splitNonceAndCiphertext(combinedEncMetadata);
        const decryptedMetaBytes = await cryptoUtils.decryptWithKey(
          metaCiphertext,
          metaNonce,
          decryptedFileKey,
        );
        originalMetadata = JSON.parse(
          await cryptoUtils.bytesToString(decryptedMetaBytes),
        );
        console.log("E2EE Download: Metadata decrypted", originalMetadata);
      } catch (metaErr) {
        console.warn(
          "E2EE Download: Failed to decrypt metadata, using defaults.",
          metaErr,
        );
      }
    }

    // 8. Trigger browser download
    const finalBlob = new Blob([decryptedFileContent], {
      type: originalMetadata.type,
    });
    const url = URL.createObjectURL(finalBlob);
    const a = document.createElement("a");
    a.href = url;
    a.download = originalMetadata.name;
    document.body.appendChild(a);
    a.click();
    setTimeout(() => {
      // Cleanup
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }, 100);
    console.log("E2EE Download: Download triggered for", originalMetadata.name);

    return { success: true, fileName: originalMetadata.name };
  },
};

export default fileAPI;
