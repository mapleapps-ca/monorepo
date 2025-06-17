import { createContext } from "react";
import { container } from "../di/container.js";

// React Context that provides the InversifyJS container to all components
export const DIContext = createContext(null);

// Provider component that makes dependency injection available throughout your app
export const DIProvider = ({ children }) => {
  return <DIContext.Provider value={container}>{children}</DIContext.Provider>;
};
