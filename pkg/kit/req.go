package kit

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cstone-io/twine/pkg/errors"
)

// Decode decodes the request body into v based on Content-Type
func (k *Kit) Decode(v any) error {
	contentType := k.GetHeader("Content-Type")

	switch {
	case contentType == "application/json":
		return k.decodeJSON(v)
	case contentType == "application/x-www-form-urlencoded":
		return k.decodeForm(v)
	default:
		return errors.ErrAPIRequestContentType
	}
}

func (k *Kit) decodeJSON(v any) error {
	if err := json.NewDecoder(k.Request.Body).Decode(v); err != nil {
		return errors.ErrDecodeJSON
	}
	return nil
}

func (k *Kit) decodeForm(v any) error {
	if err := k.Request.ParseForm(); err != nil {
		return err
	}

	val := reflect.ValueOf(v).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag.Get("form")
		if tag != "" && field.CanSet() {
			formValue := k.Request.FormValue(tag)

			switch field.Kind() {
			case reflect.String:
				field.SetString(formValue)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if formValue != "" {
					value := reflect.ValueOf(formValue)
					field.SetInt(value.Int())
				}
			case reflect.Float32, reflect.Float64:
				if formValue != "" {
					value := reflect.ValueOf(formValue)
					field.SetFloat(value.Float())
				}
			case reflect.Bool:
				if formValue != "" {
					value := reflect.ValueOf(formValue)
					field.SetBool(value.Bool())
				}
			case reflect.Struct:
				nestedPtr := reflect.New(field.Type())
				if err := k.decodeForm(nestedPtr.Interface()); err != nil {
					return err
				}
				field.Set(nestedPtr.Elem())
			case reflect.Slice:
				sliceType := field.Type().Elem()
				slice := reflect.MakeSlice(field.Type(), 0, 0)
				for _, formValue := range k.Request.Form[tag] {
					elem := reflect.New(sliceType).Elem()
					elem.SetString(formValue)
					slice = reflect.Append(slice, elem)
				}
				field.Set(slice)
			}
		}
	}

	return nil
}

// PathValue extracts a path parameter by key
func (k *Kit) PathValue(key string) string {
	return k.Request.PathValue(key)
}

// Authorization extracts the authorization token from cookie or header
func (k *Kit) Authorization() (string, error) {
	cookie, err := k.GetCookie("token")
	if err == nil && cookie != "" {
		return cookie, nil
	}

	authHeader := k.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.ErrAuthMissingHeader
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", errors.ErrAuthInvalidToken
	}

	return tokenString, nil
}

// GetHeader returns a request header value
func (k *Kit) GetHeader(key string) string {
	return k.Request.Header.Get(key)
}

// SetContext sets a context value on the request
func (k *Kit) SetContext(key, value string) {
	k.Request = k.Request.WithContext(context.WithValue(k.Request.Context(), key, value))
}

// GetContext retrieves a context value from the request
func (k *Kit) GetContext(key string) string {
	val := k.Request.Context().Value(key)
	if val == nil {
		return ""
	}
	return val.(string)
}

// SetCookie sets an HTTP cookie
func (k *Kit) SetCookie(key, value string) {
	http.SetCookie(k.Response, &http.Cookie{
		Name:     key,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(12 * time.Hour),
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // TODO: configure for dev and production
		HttpOnly: true,
	})
}

// GetCookie retrieves a cookie value
func (k *Kit) GetCookie(key string) (string, error) {
	cookie, err := k.Request.Cookie(key)
	if err != nil {
		return "", errors.ErrGetCookie.Wrap(err)
	}
	return cookie.Value, nil
}
