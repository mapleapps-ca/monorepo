// SyncCollectionExample.jsx
import React, { useState, useEffect } from "react";
import { useServices } from "../../../hooks/useService.jsx";

const SyncCollectionExample = () => {
  const { syncCollectionService } = useServices();
  const [syncCollections, setSyncCollections] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const loadSyncCollections = async () => {
      try {
        setIsLoading(true);
        setError(null);
        const collections = await syncCollectionService.getSyncCollections();
        setSyncCollections(collections);
        console.log("SyncCollections updated:", collections);
      } catch (error) {
        setError(error.message || "Failed to load SyncCollections");
      } finally {
        setIsLoading(false);
      }
    };

    loadSyncCollections();
  }, [syncCollectionService]);

  const handleRefresh = async () => {
    try {
      setError(null);
      await syncCollectionService.forceRefreshSyncCollections();
      // Optionally, re-fetch the data after refreshing
      const refreshedCollections =
        await syncCollectionService.getSyncCollections();
      setSyncCollections(refreshedCollections);
    } catch (error) {
      setError(error.message || "Failed to refresh SyncCollections");
    }
  };

  if (isLoading) {
    return (
      <div>
        <h2>Sync Collection Example</h2>
        <p>Loading SyncCollections...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div>
        <h2>Sync Collection Example</h2>
        <p style={{ color: "red" }}>Error: {error}</p>
        <button onClick={handleRefresh}>Retry Refresh</button>
      </div>
    );
  }

  return (
    <div>
      <h2>Sync Collection Example</h2>
      <button
        onClick={handleRefresh}
        disabled={syncCollectionService.getIsLoading()}
      >
        {syncCollectionService.getIsLoading() ? "Refreshing..." : "Refresh"}
      </button>
      {syncCollections.length === 0 ? (
        <p>No SyncCollections found.</p>
      ) : (
        <ul>
          {syncCollections.map((collection) => (
            // Adjust this line to match your data structure
            <li key={collection.id}> {collection.id || "No id"}</li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default SyncCollectionExample;
