// src/di/container.js

import { Container } from "inversify";
import "reflect-metadata";

import { TYPES } from "./types.js";
import { LoggerService } from "../services/LoggerService.js";
import { AuthService } from "../services/AuthService.js";
import { CryptoService } from "../services/CryptoService.js";
import { RegistrationService } from "../services/RegistrationService.js";

// Create the container
export const container = new Container();

// Create service instances manually - this is much clearer for learning!
// Notice how we're building the dependency chain step by step:

// 1. LoggerService has no dependencies, so we create it first
const loggerService = new LoggerService();

// 2. CryptoService depends on LoggerService
const cryptoService = new CryptoService(loggerService);

// 3. AuthService depends on LoggerService
const authService = new AuthService(loggerService);

// 4. RegistrationService depends on both LoggerService and CryptoService
const registrationService = new RegistrationService(
  loggerService,
  cryptoService,
);

// Bind all the actual instances to the container
// This makes them available to any component that needs them
container.bind(TYPES.LoggerService).toConstantValue(loggerService);
container.bind(TYPES.CryptoService).toConstantValue(cryptoService);
container.bind(TYPES.AuthService).toConstantValue(authService);
container.bind(TYPES.RegistrationService).toConstantValue(registrationService);

console.log("All services created and bound to container:");
console.log("- LoggerService (foundation logging)");
console.log("- CryptoService (encryption and key management)");
console.log("- AuthService (authentication logic)");
console.log("- RegistrationService (user registration and verification)");
