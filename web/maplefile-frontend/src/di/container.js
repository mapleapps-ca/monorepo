// src/di/container.js

import { Container } from "inversify";
import "reflect-metadata";

import { TYPES } from "./types.js";
import { LoggerService } from "../services/LoggerService.js";
import { AuthService } from "../services/AuthService.js";
import { CryptoService } from "../services/CryptoService.js";
import { RegistrationService } from "../services/RegistrationService.js";
import { LoginService } from "../services/LoginService.js";
import { TokenService } from "../services/TokenService.js";

// Create the container
export const container = new Container();

// Create service instances manually - this demonstrates dependency management
// Notice how we're building a sophisticated dependency graph step by step:

console.log("Building dependency injection container...");

// 1. LoggerService has no dependencies, so we create it first
// This service provides logging throughout the entire application
const loggerService = new LoggerService();
console.log("✓ LoggerService created (foundation layer)");

// 2. CryptoService depends only on LoggerService
// This service handles all cryptographic operations
const cryptoService = new CryptoService(loggerService);
console.log("✓ CryptoService created (cryptographic layer)");

// 3. TokenService manages authentication tokens and automatic refresh
// It depends on LoggerService for tracking token lifecycle events
const tokenService = new TokenService(loggerService);
console.log("✓ TokenService created (token management layer)");

// 4. LoginService orchestrates the three-step login process
// It depends on both LoggerService and CryptoService
const loginService = new LoginService(loggerService, cryptoService);
console.log("✓ LoginService created (authentication layer)");

// 5. RegistrationService handles user registration and email verification
// It depends on both LoggerService and CryptoService
const registrationService = new RegistrationService(
  loggerService,
  cryptoService,
);
console.log("✓ RegistrationService created (registration layer)");

// 6. AuthService is the main coordinator that brings everything together
// It depends on LoggerService, LoginService, and TokenService
// This service acts as the primary interface between React components and authentication
const authService = new AuthService(loggerService, loginService, tokenService);
console.log("✓ AuthService created (orchestration layer)");

// Bind all the actual instances to the container
// This makes them available to any component that needs them through useService()
container.bind(TYPES.LoggerService).toConstantValue(loggerService);
container.bind(TYPES.CryptoService).toConstantValue(cryptoService);
container.bind(TYPES.TokenService).toConstantValue(tokenService);
container.bind(TYPES.LoginService).toConstantValue(loginService);
container.bind(TYPES.RegistrationService).toConstantValue(registrationService);
container.bind(TYPES.AuthService).toConstantValue(authService);

console.log("✓ All services bound to container successfully!");
console.log("");
console.log("Service Architecture Summary:");
console.log("  Foundation: LoggerService");
console.log("  Security: CryptoService");
console.log("  Token Management: TokenService (with automatic refresh)");
console.log("  Authentication: LoginService (3-step process)");
console.log("  Registration: RegistrationService (with E2E encryption)");
console.log("  Orchestration: AuthService (main coordinator)");
console.log("");
console.log("This dependency injection setup allows for:");
console.log("- Clean separation of concerns");
console.log("- Easy testing with mock services");
console.log("- Flexible service replacement without code changes");
console.log("- Clear dependency relationships");
console.log("- Consistent logging and error handling");
console.log("");

// Initialize the authentication service for the entire application
// This starts token refresh monitoring and checks for existing authentication
authService
  .initialize()
  .then((isAuthenticated) => {
    console.log(
      `Authentication service initialized. User authenticated: ${isAuthenticated}`,
    );
  })
  .catch((error) => {
    console.error(`Error initializing authentication service:`, error);
  });
