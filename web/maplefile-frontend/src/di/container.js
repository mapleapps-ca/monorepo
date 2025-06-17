import { Container } from "inversify";
import "reflect-metadata";

import { TYPES } from "./types.js";
import { LoggerService } from "../services/LoggerService.js";
import { AuthService } from "../services/AuthService.js";

// Create the container
export const container = new Container();

// Create service instances manually - this is much clearer for learning!
const loggerService = new LoggerService();
const authService = new AuthService(loggerService);

// Bind the actual instances to the container
container.bind(TYPES.LoggerService).toConstantValue(loggerService);
container.bind(TYPES.AuthService).toConstantValue(authService);

console.log("Services created and bound to container");
