// Example usage of CollectionService in React components

// Example 1: Using the service directly via useServices hook
import React, { useState, useEffect } from "react";
import { useServices } from "../hooks/useService.jsx";

const CollectionManagerExample = () => {
  const { collectionService, cryptoService } = useServices();
  const [collections, setCollections] = useState([]);
  const [loading, setLoading] = useState(false);

  // Load collections on mount
  useEffect(() => {
    loadCollections();
  }, []);

  const loadCollections = async () => {
    try {
      setLoading(true);
      const userCollections = await collectionService.listUserCollections();
      setCollections(userCollections);
    } catch (error) {
      console.error("Failed to load collections:", error);
    } finally {
      setLoading(false);
    }
  };

  // Create a new encrypted folder
  const createFolder = async (name, parentId = null) => {
    try {
      // Generate encryption key for the collection
      const collectionKey = cryptoService.generateCollectionKey(); // You'll need to implement this

      // Encrypt the collection name
      const encryptedName = await cryptoService.encryptWithKey(
        name,
        collectionKey,
      );

      // Encrypt the collection key with user's public key
      const encryptedCollectionKey =
        await cryptoService.encryptCollectionKey(collectionKey);

      const collectionData = {
        encrypted_name: encryptedName,
        collection_type: "folder",
        encrypted_collection_key: {
          ciphertext: Array.from(encryptedCollectionKey.ciphertext),
          nonce: Array.from(encryptedCollectionKey.nonce),
          key_version: 1,
          rotated_at: new Date().toISOString(),
        },
        parent_id: parentId,
        ancestor_ids: [], // Will be set by backend
      };

      const newCollection =
        await collectionService.createCollection(collectionData);

      // Refresh collections list
      await loadCollections();

      return newCollection;
    } catch (error) {
      console.error("Failed to create folder:", error);
      throw error;
    }
  };

  // Share a collection with another user
  const shareCollectionWithUser = async (
    collectionId,
    recipientEmail,
    recipientId,
    recipientPublicKey,
  ) => {
    try {
      // Get the collection to access its key
      const collection = await collectionService.getCollection(collectionId);

      // Decrypt the collection key with your private key
      const collectionKey = await cryptoService.decryptCollectionKey(
        collection.encrypted_collection_key,
      );

      // Re-encrypt the collection key with recipient's public key
      const encryptedKeyForRecipient = await cryptoService.encryptWithPublicKey(
        collectionKey,
        recipientPublicKey,
      );

      const shareData = {
        recipient_id: recipientId,
        recipient_email: recipientEmail,
        permission_level: "read_write",
        encrypted_collection_key: Array.from(encryptedKeyForRecipient),
        share_with_descendants: true,
      };

      await collectionService.shareCollection(collectionId, shareData);
      console.log("Collection shared successfully");
    } catch (error) {
      console.error("Failed to share collection:", error);
      throw error;
    }
  };

  return (
    <div>
      <h2>Collection Manager</h2>
      {loading ? (
        <p>Loading collections...</p>
      ) : (
        <div>
          <button onClick={() => createFolder("New Folder")}>
            Create New Folder
          </button>
          <ul>
            {collections.map((collection) => (
              <li key={collection.id}>
                {collection.encrypted_name} ({collection.collection_type})
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

// Example 2: Using the custom useCollections hook
import React from "react";
import useCollections from "../hooks/useCollections.js";

const CollectionExplorerExample = () => {
  const {
    collections,
    sharedCollections,
    isLoading,
    error,
    createCollection,
    updateCollection,
    deleteCollection,
    shareCollection,
    getRootCollections,
    getChildCollections,
    COLLECTION_TYPES,
    PERMISSION_LEVELS,
  } = useCollections();

  const handleCreateAlbum = async () => {
    try {
      const albumData = {
        encrypted_name: "base64_encrypted_album_name",
        collection_type: COLLECTION_TYPES.ALBUM,
        encrypted_collection_key: {
          ciphertext: [
            /* encrypted key bytes */
          ],
          nonce: [
            /* nonce bytes */
          ],
          key_version: 1,
          rotated_at: new Date().toISOString(),
        },
      };

      const newAlbum = await createCollection(albumData);
      console.log("Album created:", newAlbum);
    } catch (error) {
      console.error("Failed to create album:", error);
    }
  };

  const handleRenameCollection = async (collectionId, newName) => {
    try {
      // Encrypt the new name
      const encryptedNewName = "base64_encrypted_new_name"; // Implement encryption

      await updateCollection(collectionId, {
        encrypted_name: encryptedNewName,
        version: 1, // Include current version for optimistic locking
      });
    } catch (error) {
      console.error("Failed to rename collection:", error);
    }
  };

  const renderCollectionTree = (parentId = null, level = 0) => {
    const children = parentId
      ? getChildCollections(parentId)
      : getRootCollections();

    return children.map((collection) => (
      <div key={collection.id} style={{ marginLeft: `${level * 20}px` }}>
        <div>
          📁 {collection.encrypted_name}
          {collection.members?.length > 1 &&
            ` (Shared with ${collection.members.length - 1})`}
        </div>
        {renderCollectionTree(collection.id, level + 1)}
      </div>
    ));
  };

  if (error) {
    return <div>Error: {error}</div>;
  }

  return (
    <div>
      <h2>My Collections</h2>

      <button onClick={handleCreateAlbum} disabled={isLoading}>
        Create New Album
      </button>

      {isLoading ? (
        <p>Loading collections...</p>
      ) : (
        <div>
          <h3>My Collections ({collections.length})</h3>
          {renderCollectionTree()}

          {sharedCollections.length > 0 && (
            <>
              <h3>Shared with Me ({sharedCollections.length})</h3>
              {sharedCollections.map((collection) => (
                <div key={collection.id}>
                  📁 {collection.encrypted_name} (Owner: {collection.owner_id})
                </div>
              ))}
            </>
          )}
        </div>
      )}
    </div>
  );
};

// Example 3: Advanced collection operations
const AdvancedCollectionExample = () => {
  const { collectionService } = useServices();

  // Sync collections for offline support
  const syncAllCollections = async () => {
    let cursor = null;
    let hasMore = true;
    const allSyncedCollections = [];

    while (hasMore) {
      const syncResult = await collectionService.syncCollections(cursor, 1000);
      allSyncedCollections.push(...syncResult.collections);

      cursor = syncResult.next_cursor;
      hasMore = syncResult.has_more;
    }

    console.log(`Synced ${allSyncedCollections.length} collections`);
    return allSyncedCollections;
  };

  // Move collection to a new parent
  const moveCollectionToParent = async (collectionId, newParentId) => {
    try {
      // Get the new parent's hierarchy
      const parentHierarchy =
        await collectionService.getCollectionHierarchy(newParentId);

      const moveData = {
        new_parent_id: newParentId,
        updated_ancestors: parentHierarchy.map((col) => col.id),
        updated_path_segments: parentHierarchy.map((col) => col.encrypted_name),
      };

      await collectionService.moveCollection(collectionId, moveData);
      console.log("Collection moved successfully");
    } catch (error) {
      console.error("Failed to move collection:", error);
    }
  };

  // Batch operations
  const createFolderStructure = async () => {
    try {
      // Create root folders
      const documentsFolder = await collectionService.createCollection({
        encrypted_name: "encrypted_documents",
        collection_type: "folder",
        encrypted_collection_key: {
          /* ... */
        },
      });

      const photosFolder = await collectionService.createCollection({
        encrypted_name: "encrypted_photos",
        collection_type: "album",
        encrypted_collection_key: {
          /* ... */
        },
      });

      // Create subfolders
      await collectionService.batchCreateCollections([
        {
          encrypted_name: "encrypted_work",
          collection_type: "folder",
          parent_id: documentsFolder.id,
          ancestor_ids: [documentsFolder.id],
          encrypted_collection_key: {
            /* ... */
          },
        },
        {
          encrypted_name: "encrypted_personal",
          collection_type: "folder",
          parent_id: documentsFolder.id,
          ancestor_ids: [documentsFolder.id],
          encrypted_collection_key: {
            /* ... */
          },
        },
      ]);

      console.log("Folder structure created");
    } catch (error) {
      console.error("Failed to create folder structure:", error);
    }
  };

  return null; // This is just for example functions
};

export {
  CollectionManagerExample,
  CollectionExplorerExample,
  AdvancedCollectionExample,
};
