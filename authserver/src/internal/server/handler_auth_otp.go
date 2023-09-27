package server

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/leodip/goiabada/internal/common"
	"github.com/leodip/goiabada/internal/customerrors"
	"github.com/leodip/goiabada/internal/entities"
	"github.com/leodip/goiabada/internal/enums"
	"github.com/pquerna/otp/totp"
)

func (s *Server) handleAuthOtpGet(otpSecretGenerator otpSecretGenerator) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		sess, err := s.sessionStore.Get(r, common.SessionName)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}
		authContext, err := s.getAuthContext(r)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}

		user, err := s.database.GetUserById(authContext.UserId)
		if err != nil || user == nil {
			s.internalServerError(w, r, err)
			return
		}

		if len(user.OTPSecret) == 0 {
			// must enroll first

			// generate secret
			settings := r.Context().Value(common.ContextKeySettings).(*entities.Settings)
			base64Image, secretKey, err := otpSecretGenerator.GenerateOTPSecret(user, settings)
			if err != nil {
				s.internalServerError(w, r, err)
				return
			}

			bind := map[string]interface{}{
				"error":       nil,
				"csrfField":   csrf.TemplateField(r),
				"base64Image": base64Image,
				"secretKey":   secretKey,
			}

			// save image and secret in the session state
			sess.Values[common.SessionKeyOTPSecret] = secretKey
			sess.Values[common.SessionKeyOTPImage] = base64Image
			err = sess.Save(r, w)
			if err != nil {
				s.internalServerError(w, r, err)
				return
			}

			err = s.renderTemplate(w, r, "/layouts/layout.html", "/auth_otp_enrollment.html", bind)
			if err != nil {
				s.internalServerError(w, r, err)
				return
			}
		} else {

			delete(sess.Values, common.SessionKeyOTPImage)
			delete(sess.Values, common.SessionKeyOTPSecret)
			sess.Save(r, w)

			bind := map[string]interface{}{
				"error":     nil,
				"csrfField": csrf.TemplateField(r),
			}

			err = s.renderTemplate(w, r, "/layouts/layout.html", "/auth_otp.html", bind)
			if err != nil {
				s.internalServerError(w, r, err)
				return
			}
		}
	}
}

func (s *Server) renderOtpPostError(w http.ResponseWriter, r *http.Request, err error, base64Image string, secretKey string) {

	if appError, ok := err.(*customerrors.AppError); ok {
		if appError.StatusCode == http.StatusInternalServerError {
			s.internalServerError(w, r, appError)
			return
		}

		bind := map[string]interface{}{
			"error":     appError.Description,
			"csrfField": csrf.TemplateField(r),
		}

		template := "/auth_otp.html"
		if len(base64Image) > 0 && len(secretKey) > 0 {
			template = "/auth_otp_enrollment.html"
			bind["base64Image"] = base64Image
			bind["secretKey"] = secretKey
		}

		err = s.renderTemplate(w, r, "/layouts/layout.html", template, bind)
		if err != nil {
			s.internalServerError(w, r, err)
		}
	} else {
		s.internalServerError(w, r, err)
	}
}

func (s *Server) handleAuthOtpPost() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		authContext, err := s.getAuthContext(r)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}
		sess, err := s.sessionStore.Get(r, common.SessionName)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}

		base64Image, secretKey := "", ""
		if val, ok := sess.Values[common.SessionKeyOTPImage]; ok {
			base64Image = val.(string)
		}
		if val, ok := sess.Values[common.SessionKeyOTPSecret]; ok {
			secretKey = val.(string)
		}

		user, err := s.database.GetUserById(authContext.UserId)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}

		otpCode := r.FormValue("otp")
		if len(otpCode) == 0 {
			s.renderOtpPostError(w, r, customerrors.NewAppError(nil, "", "OTP code is required.", http.StatusOK),
				base64Image, secretKey)
			return
		}

		incorrectOtpError := customerrors.NewAppError(nil, "", "Incorrect OTP Code. Please ensure that you've entered the code correctly. OTP codes are time-sensitive and change every 30 seconds. Make sure you're using the most recent code generated by your authenticator app.", http.StatusOK)

		if len(user.OTPSecret) > 0 {
			// already has OTP enrolled
			otpValid := totp.Validate(otpCode, user.OTPSecret)
			if !otpValid {
				s.renderOtpPostError(w, r, incorrectOtpError, base64Image, secretKey)
				return
			}
		} else {
			// is enrolling to TOTP now
			otpValid := totp.Validate(otpCode, secretKey)
			if !otpValid {
				s.renderOtpPostError(w, r, incorrectOtpError, base64Image, secretKey)
				return
			}

			// save TOTP secret
			user.OTPSecret = secretKey
			user, err = s.database.UpdateUser(user)
			if err != nil {
				s.renderOtpPostError(w, r, err, base64Image, secretKey)
				return
			}
		}

		// start new session
		_, err = s.startNewUserSession(w, r, user.ID, enums.AuthMethodPassword.String()+" "+enums.AuthMethodOTP.String())
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}

		// redirect to consent
		authContext.UserId = user.ID
		authContext.AcrLevel = enums.AcrLevel2.String()
		authContext.AuthMethods = enums.AuthMethodPassword.String() + " " + enums.AuthMethodOTP.String()
		authContext.AuthCompleted = true
		err = s.saveAuthContext(w, r, authContext)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}

		http.Redirect(w, r, "/auth/consent", http.StatusFound)
	}
}
