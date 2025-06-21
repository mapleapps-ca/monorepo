// monorepo/web/maplefile-frontend/src/services/MeService.js
class MeService {
  constructor(authService) {
    // MeService depends on AuthService to get the current user
    this.authService = authService;
  }
}

export default MeService;
