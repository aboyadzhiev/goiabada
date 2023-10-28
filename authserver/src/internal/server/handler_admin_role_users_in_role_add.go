package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/leodip/goiabada/internal/common"
	"github.com/leodip/goiabada/internal/dtos"
	"github.com/leodip/goiabada/internal/lib"
)

func (s *Server) handleAdminRoleUsersInRoleAddGet() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		allowedScopes := []string{"authserver:admin-website"}
		var jwtInfo dtos.JwtInfo
		if r.Context().Value(common.ContextKeyJwtInfo) != nil {
			jwtInfo = r.Context().Value(common.ContextKeyJwtInfo).(dtos.JwtInfo)
		}

		if !s.isAuthorizedToAccessResource(jwtInfo, allowedScopes) {
			s.redirToAuthorize(w, r, "system-website", lib.GetBaseUrl()+r.RequestURI)
			return
		}

		idStr := chi.URLParam(r, "roleId")
		if len(idStr) == 0 {
			s.internalServerError(w, r, errors.New("roleId is required"))
			return
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}
		role, err := s.database.GetRoleById(uint(id))
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}
		if role == nil {
			s.internalServerError(w, r, errors.New("role not found"))
			return
		}

		bind := map[string]interface{}{
			"roleId":         role.Id,
			"roleIdentifier": role.RoleIdentifier,
			"description":    role.Description,
			"csrfField":      csrf.TemplateField(r),
		}

		err = s.renderTemplate(w, r, "/layouts/menu_layout.html", "/admin_roles_users_in_roles_add.html", bind)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}
	}
}

func (s *Server) handleAdminRoleUsersInRoleSearchGet() http.HandlerFunc {

	type userResult struct {
		Id          uint
		Subject     string
		Username    string
		Email       string
		GivenName   string
		MiddleName  string
		FamilyName  string
		AddedToRole bool
	}

	type searchResult struct {
		RequiresAuth bool
		Users        []userResult
	}

	return func(w http.ResponseWriter, r *http.Request) {
		result := searchResult{
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

		query := strings.TrimSpace(r.URL.Query().Get("query"))
		if len(query) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)
			return
		}

		users, err := s.database.SearchUsers(query)
		if err != nil {
			s.jsonError(w, r, err)
			return
		}

		usersResult := make([]userResult, 0)
		for _, user := range users {

			userInRole := false
			for _, userRole := range user.Roles {
				if userRole.Id == role.Id {
					userInRole = true
					break
				}
			}

			usersResult = append(usersResult, userResult{
				Id:          user.Id,
				Subject:     user.Subject.String(),
				Username:    user.Username,
				Email:       user.Email,
				GivenName:   user.GivenName,
				MiddleName:  user.MiddleName,
				FamilyName:  user.FamilyName,
				AddedToRole: userInRole,
			})
		}

		result.Users = usersResult
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func (s *Server) handleAdminRoleUsersInRoleAddPost() http.HandlerFunc {

	type addResult struct {
		RequiresAuth      bool
		AddedSuccessfully bool
	}

	return func(w http.ResponseWriter, r *http.Request) {
		result := addResult{
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

		userIdStr := r.URL.Query().Get("userId")
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

		err = s.database.AddUserToRole(user, role)
		if err != nil {
			s.jsonError(w, r, err)
			return
		}

		result.AddedSuccessfully = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
