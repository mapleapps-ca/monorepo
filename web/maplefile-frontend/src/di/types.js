// src/di/types.js

// These symbols are like unique name tags for our services
// They prevent naming conflicts and make our code more reliable
export const TYPES = {
  LoggerService: Symbol.for("LoggerService"),
  AuthService: Symbol.for("AuthService"),
  CryptoService: Symbol.for("CryptoService"),
  RegistrationService: Symbol.for("RegistrationService"),
  LoginService: Symbol.for("LoginService"),
  TokenService: Symbol.for("TokenService"),
};
