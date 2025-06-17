import { useContext, useMemo } from "react";
import { DIContext } from "../contexts/DIProvider.jsx";

// This hook lets your React components easily get services
// It's the bridge between InversifyJS and React
export const useService = (serviceType) => {
  const diContainer = useContext(DIContext);

  if (!diContainer) {
    throw new Error("useService must be used within a DIProvider");
  }

  // useMemo ensures we get the same service instance on every render
  // This is important for performance and consistency
  const service = useMemo(() => {
    return diContainer.get(serviceType);
  }, [diContainer, serviceType]);

  return service;
};
