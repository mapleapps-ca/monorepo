package middleware

import (
	// "context"
	"net/http"
	// "github.com/mapleapps-ca/monorepo/cloud/mapleapps-backend/config/constants"
)

func (mid *middleware) PostJWTProcessorMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ctx := r.Context()

		// // Get our authorization information.
		// isAuthorized, ok := ctx.Value(constants.SessionIsAuthorized).(bool)
		// if ok && isAuthorized {
		// 	sessionID := ctx.Value(constants.SessionID).(string)

		// 	// Lookup our user profile in the session or return 500 error.
		// 	user, err := mid.userGetBySessionIDUseCase.Execute(ctx, sessionID)
		// 	if err != nil {
		// 		http.Error(w, err.Error(), http.StatusInternalServerError)
		// 		return
		// 	}

		// 	// If no user was found then that means our session expired and the
		// 	// user needs to login or use the refresh token.
		// 	if user == nil {
		// 		http.Error(w, "attempting to access a protected endpoint", http.StatusUnauthorized)
		// 		return
		// 	}

		// 	// // If system administrator disabled the user account then we need
		// 	// // to generate a 403 error letting the user know their account has
		// 	// // been disabled and you cannot access the protected API endpoint.
		// 	// if user.State == 0 {
		// 	// 	http.Error(w, "Account disabled - please contact admin", http.StatusForbidden)
		// 	// 	return
		// 	// }

		// 	// Save our user information to the context.
		// 	// Save our user.
		// 	ctx = context.WithValue(ctx, constants.SessionFederatedUser, user)

		// 	// Save individual pieces of the user profile.
		// 	ctx = context.WithValue(ctx, constants.SessionID, sessionID)
		// 	ctx = context.WithValue(ctx, constants.SessionFederatedUserID, user.ID)
		// 	ctx = context.WithValue(ctx, constants.SessionFederatedUserRole, user.Role)
		// 	ctx = context.WithValue(ctx, constants.SessionFederatedUserName, user.Name)
		// 	ctx = context.WithValue(ctx, constants.SessionFederatedUserFirstName, user.FirstName)
		// 	ctx = context.WithValue(ctx, constants.SessionFederatedUserLastName, user.LastName)
		// 	ctx = context.WithValue(ctx, constants.SessionFederatedUserTimezone, user.Timezone)
		// 	// ctx = context.WithValue(ctx, constants.SessionFederatedUserStoreID, user.StoreID)
		// 	// ctx = context.WithValue(ctx, constants.SessionFederatedUserStoreName, user.StoreName)
		// 	// ctx = context.WithValue(ctx, constants.SessionFederatedUserStoreLevel, user.StoreLevel)
		// 	// ctx = context.WithValue(ctx, constants.SessionFederatedUserStoreTimezone, user.StoreTimezone)
		// }

		// fn(w, r.WithContext(ctx))
	}
}
