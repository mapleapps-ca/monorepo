// Hook to access services from ServiceContext
import { useContext } from "react";
import { ServiceContext } from "../contexts/ServiceContext.jsx";

// Create a custom hook to use our services
export const useServices = () => {
  const context = useContext(ServiceContext);
  if (!context) {
    throw new Error("useServices must be used within a ServiceProvider");
  }
  return context;
};
