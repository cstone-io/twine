package errors

import "net/http"

// Predefined errors follow a naming convention of Err<Description>

var (
	// 1000 level errors are CRITICAL severity
	ErrDefaultCritical = NewErrorBuilder().Code(1000).Severity(ErrCritical).Message("DEFAULT OR UNKNOWN CRITICAL APPLICATION ERROR!!!").Build()
	ErrListenAndServe  = NewErrorBuilder().Code(1001).Severity(ErrCritical).Message("FAILED TO LISTEN AND SERVE").Build()
	ErrShutdownServer  = NewErrorBuilder().Code(1002).Severity(ErrCritical).Message("FAILED TO SHUTDOWN SERVER").Build()

	// 1100 level errors are DATABASE critical errors
	ErrDatabaseDefaultCritical = NewErrorBuilder().Code(1100).Severity(ErrCritical).Message("DEFAULT OR UNKNOWN CRITICAL DATABASE ERROR!!!").Build()
	ErrDatabaseLoad            = NewErrorBuilder().Code(1101).Severity(ErrCritical).Message("FAILED TO LOAD DATABASE").Build()
	ErrDatabaseConn            = NewErrorBuilder().Code(1102).Severity(ErrCritical).Message("FAILED TO CONNECT TO DATABASE").Build()
	ErrDatabaseMigration       = NewErrorBuilder().Code(1103).Severity(ErrCritical).Message("FAILED TO MIGRATE DATABASE").Build()
	ErrDatabaseSeed            = NewErrorBuilder().Code(1104).Severity(ErrCritical).Message("FAILED TO SEED DATABASE").Build()

	// 2000 level errors are ERROR severity
	ErrDefaultError = NewErrorBuilder().Code(2000).Severity(ErrError).Message("Default or unknown error").Build()
	ErrDecodeJSON   = NewErrorBuilder().Code(2001).Severity(ErrError).Message("Failed to decode JSON").Build()
	ErrNotFound     = NewErrorBuilder().Code(2002).Severity(ErrError).HTTPStatus(http.StatusNotFound).Message("Not found").Build()

	// 2100 level errors are for DATABASE errors
	ErrDatabaseDefaultError = NewErrorBuilder().Code(2100).Severity(ErrError).Message("Default or unknown database error").Build()
	ErrDatabaseRead         = NewErrorBuilder().Code(2101).Severity(ErrError).Message("Failed to read from database").Build()
	ErrDatabaseWrite        = NewErrorBuilder().Code(2102).Severity(ErrError).Message("Failed to write to database").Build()
	ErrDatabaseUpdate       = NewErrorBuilder().Code(2103).Severity(ErrError).Message("Failed to update database").Build()
	ErrDatabaseDelete       = NewErrorBuilder().Code(2104).Severity(ErrError).Message("Failed to delete from database").Build()
	ErrMigrateTable         = NewErrorBuilder().Code(2105).Severity(ErrError).Message("Failed to migrate database table").Build()
	ErrSortMigrations       = NewErrorBuilder().Code(2106).Severity(ErrError).Message("Failed to sort migrations").Build()
	ErrSeedObject           = NewErrorBuilder().Code(2107).Severity(ErrError).Message("Failed to seed object").Build()

	// 2200 level errors are for AUTH errors
	ErrAuthDefault    = NewErrorBuilder().Code(2200).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Default or unknown AUTH error").Build()
	ErrHashPassword   = NewErrorBuilder().Code(2201).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Failed to hash password").Build()
	ErrGenerateToken  = NewErrorBuilder().Code(2202).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Failed to generate token").Build()
	ErrGetPermissions = NewErrorBuilder().Code(2203).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Failed to get IAM permissions").Build()
	ErrGetCookie      = NewErrorBuilder().Code(2204).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Failed to get cookie").Build()

	// 2300 level errors are for API errors
	ErrAPIDefault = NewErrorBuilder().Code(2300).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Default or unknown API error").Build()
	ErrAPIGet     = NewErrorBuilder().Code(2301).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Failed to GET data").Build()
	ErrAPIPost    = NewErrorBuilder().Code(2302).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Failed to POST data").Build()
	ErrAPIPut     = NewErrorBuilder().Code(2303).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Failed to PUT data").Build()
	ErrAPIDelete  = NewErrorBuilder().Code(2304).Severity(ErrError).HTTPStatus(http.StatusInternalServerError).Message("Failed to DELETE data").Build()

	// 3000 level errors are MINOR severity
	ErrDefaultMinor = NewErrorBuilder().Code(3000).Severity(ErrMinor).HTTPStatus(http.StatusInternalServerError).Message("Default or unknown warning").Build()
	ErrDecodeForm   = NewErrorBuilder().Code(3001).Severity(ErrMinor).Message("Failed to decode form").Build()

	// 3100 level errors are for DATABASE minor errors
	ErrDatabaseDefaultMinor   = NewErrorBuilder().Code(3100).Severity(ErrMinor).HTTPStatus(http.StatusInternalServerError).Message("Default or unknown database warning").Build()
	ErrDatabaseObjectNotFound = NewErrorBuilder().Code(3101).Severity(ErrMinor).HTTPStatus(http.StatusNotFound).Message("Object not found").Build()

	// 3200 level errors are for AUTH minor errors
	ErrAuthDefaultMinor          = NewErrorBuilder().Code(3200).Severity(ErrMinor).HTTPStatus(http.StatusInternalServerError).Message("Default or unknown AUTH warning").Build()
	ErrAuthInvalidToken          = NewErrorBuilder().Code(3201).Severity(ErrMinor).HTTPStatus(http.StatusUnauthorized).Message("Invalid token").Build()
	ErrAuthExpiredToken          = NewErrorBuilder().Code(3202).Severity(ErrMinor).HTTPStatus(http.StatusUnauthorized).Message("Expired token").Build()
	ErrAuthInvalidCredentials    = NewErrorBuilder().Code(3203).Severity(ErrMinor).HTTPStatus(http.StatusUnauthorized).Message("Invalid credentials").Build()
	ErrPrimaryEmailNotFound      = NewErrorBuilder().Code(3204).Severity(ErrMinor).HTTPStatus(http.StatusNotFound).Message("Primary email not found").Build()
	ErrInsufficientPermissions   = NewErrorBuilder().Code(3205).Severity(ErrMinor).HTTPStatus(http.StatusForbidden).Message("Insufficient permissions").Build()
	ErrAuthMissingHeader         = NewErrorBuilder().Code(3206).Severity(ErrMinor).HTTPStatus(http.StatusBadRequest).Message("Missing Authorization header").Build()
	ErrAuthMissingAuthTypeHeader = NewErrorBuilder().Code(3207).Severity(ErrMinor).HTTPStatus(http.StatusBadRequest).Message("Missing Authorization-Type header").Build()

	// 3300 level errors are for API minor errors
	ErrAPIDefaultMinor       = NewErrorBuilder().Code(3300).Severity(ErrMinor).HTTPStatus(http.StatusInternalServerError).Message("Default or unknown API warning").Build()
	ErrAPIIDMismatch         = NewErrorBuilder().Code(3301).Severity(ErrMinor).HTTPStatus(http.StatusBadRequest).Message("ID in URL does not match ID in body").Build()
	ErrAPIRequestPayload     = NewErrorBuilder().Code(3302).Severity(ErrMinor).HTTPStatus(http.StatusBadRequest).Message("Invalid request payload").Build()
	ErrAPIPathValue          = NewErrorBuilder().Code(3303).Severity(ErrMinor).HTTPStatus(http.StatusBadRequest).Message("Invalid path value").Build()
	ErrAPIObjectNotFound     = NewErrorBuilder().Code(3304).Severity(ErrMinor).HTTPStatus(http.StatusNotFound).Message("Object not found").Build()
	ErrAPIRequestContentType = NewErrorBuilder().Code(3305).Severity(ErrMinor).HTTPStatus(http.StatusUnsupportedMediaType).Message("Unsupported content type").Build()
)
