// These symbols are like unique name tags for our services
// They prevent naming conflicts and make our code more reliable
export const TYPES = {
  LoggerService: Symbol.for("LoggerService"),
  AuthService: Symbol.for("AuthService"),
};
