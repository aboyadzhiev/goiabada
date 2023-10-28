package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/leodip/goiabada/internal/common"
	"github.com/leodip/goiabada/internal/dtos"
)

func (s *Server) handleAdminRoleUsersInRoleRemoveUserPost() http.HandlerFunc {

	type removeResult struct {
		RequiresAuth        bool
		RemovedSuccessfully bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		result := removeResult{
			RequiresAuth: true,
		}

		allowedScopes := []string{"authserver:admin-website"}
		var jwtInfo dtos.JwtInfo
		if r.Context().Value(common.ContextKeyJwtInfo) != nil {
			jwtInfo = r.Context().Value(common.ContextKeyJwtInfo).(dtos.JwtInfo)
		}

		if s.isAuthorizedToAccessResource(jwtInfo, allowedScopes) {
			result.RequiresAuth = false
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
			return
		}

		idStr := chi.URLParam(r, "roleId")
		if len(idStr) == 0 {
			s.jsonError(w, r, errors.New("roleId is required"))
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			s.jsonError(w, r, err)
			return
		}
		role, err := s.database.GetRoleById(uint(id))
		if err != nil {
			s.jsonError(w, r, err)
			return
		}
		if role == nil {
			s.jsonError(w, r, errors.New("role not found"))
			return
		}

		userIdStr := chi.URLParam(r, "userId")
		if len(userIdStr) == 0 {
			s.jsonError(w, r, errors.New("userId is required"))
			return
		}

		userId, err := strconv.ParseUint(userIdStr, 10, 64)
		if err != nil {
			s.jsonError(w, r, err)
			return
		}

		user, err := s.database.GetUserById(uint(userId))
		if err != nil {
			s.jsonError(w, r, err)
			return
		}

		if user == nil {
			s.jsonError(w, r, errors.New("user not found"))
			return
		}

		err = s.database.RemoveUserFromRole(user, role)
		if err != nil {
			s.jsonError(w, r, err)
			return
		}

		result.RemovedSuccessfully = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
