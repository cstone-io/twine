package assert

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertJSONResponse asserts that the HTTP response has the expected status code,
// content type, and JSON body.
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, status int, expectedJSON string) {
	t.Helper()
	assert.Equal(t, status, w.Code, "unexpected status code")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "unexpected content type")
	assert.JSONEq(t, expectedJSON, w.Body.String(), "unexpected JSON response")
}

// AssertHTMLResponse asserts that the HTTP response has the expected status code
// and contains the expected HTML content.
func AssertHTMLResponse(t *testing.T, w *httptest.ResponseRecorder, status int, contains string) {
	t.Helper()
	assert.Equal(t, status, w.Code, "unexpected status code")
	assert.Contains(t, w.Body.String(), contains, "response body does not contain expected content")
}

// AssertTextResponse asserts that the HTTP response has the expected status code
// and text body.
func AssertTextResponse(t *testing.T, w *httptest.ResponseRecorder, status int, expectedText string) {
	t.Helper()
	assert.Equal(t, status, w.Code, "unexpected status code")
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"), "unexpected content type")
	assert.Equal(t, expectedText, w.Body.String(), "unexpected text response")
}

// AssertHeader asserts that the HTTP response has the expected header value.
func AssertHeader(t *testing.T, w *httptest.ResponseRecorder, header, expectedValue string) {
	t.Helper()
	actual := w.Header().Get(header)
	assert.Equal(t, expectedValue, actual, "unexpected header value for %s", header)
}

// RequireHeader requires that the HTTP response has the expected header value.
// Fails the test immediately if the header doesn't match.
func RequireHeader(t *testing.T, w *httptest.ResponseRecorder, header, expectedValue string) {
	t.Helper()
	actual := w.Header().Get(header)
	require.Equal(t, expectedValue, actual, "unexpected header value for %s", header)
}

// AssertStatusCode asserts that the HTTP response has the expected status code.
func AssertStatusCode(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	assert.Equal(t, expectedStatus, w.Code, "unexpected status code")
}

// RequireStatusCode requires that the HTTP response has the expected status code.
// Fails the test immediately if the status code doesn't match.
func RequireStatusCode(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	require.Equal(t, expectedStatus, w.Code, "unexpected status code")
}

// AssertBodyContains asserts that the HTTP response body contains the expected substring.
func AssertBodyContains(t *testing.T, w *httptest.ResponseRecorder, substring string) {
	t.Helper()
	assert.Contains(t, w.Body.String(), substring, "response body does not contain expected substring")
}

// AssertBodyNotContains asserts that the HTTP response body does not contain the substring.
func AssertBodyNotContains(t *testing.T, w *httptest.ResponseRecorder, substring string) {
	t.Helper()
	assert.NotContains(t, w.Body.String(), substring, "response body contains unexpected substring")
}

// AssertAjaxResponse asserts that the response is an Ajax partial response.
// Checks for appropriate headers and partial HTML content.
func AssertAjaxResponse(t *testing.T, w *httptest.ResponseRecorder, status int) {
	t.Helper()
	assert.Equal(t, status, w.Code, "unexpected status code")
	// Ajax responses should not contain full HTML document structure
	body := w.Body.String()
	assert.NotContains(t, body, "<!DOCTYPE html>", "Ajax response contains full HTML document")
	assert.NotContains(t, body, "<html", "Ajax response contains <html> tag")
}

// AssertRedirect asserts that the response is a redirect to the expected location.
func AssertRedirect(t *testing.T, w *httptest.ResponseRecorder, expectedLocation string) {
	t.Helper()
	assert.True(t, w.Code >= 300 && w.Code < 400, "expected redirect status code (3xx), got %d", w.Code)
	assert.Equal(t, expectedLocation, w.Header().Get("Location"), "unexpected redirect location")
}
