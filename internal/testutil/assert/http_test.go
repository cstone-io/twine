package assert_test

import (
	"net/http/httptest"
	"testing"

	httpAssert "github.com/cstone-io/twine/internal/testutil/assert"
	"github.com/stretchr/testify/assert"
)

func TestAssertJSONResponse_ValidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(`{"message": "hello"}`))

	httpAssert.AssertJSONResponse(t, w, 200, `{"message": "hello"}`)
}

func TestAssertHTMLResponse_ContainsText(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(200)
	w.Write([]byte(`<html><body><h1>Hello World</h1></body></html>`))

	httpAssert.AssertHTMLResponse(t, w, 200, "Hello World")
}

func TestAssertTextResponse_PlainText(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("plain text response"))

	httpAssert.AssertTextResponse(t, w, 200, "plain text response")
}

func TestAssertHeader_VerifiesHeader(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("X-Custom-Header", "custom-value")
	w.WriteHeader(200)

	httpAssert.AssertHeader(t, w, "X-Custom-Header", "custom-value")
}

func TestAssertStatusCode_VerifiesStatus(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(404)

	httpAssert.AssertStatusCode(t, w, 404)
}

func TestAssertBodyContains_FindsSubstring(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(200)
	w.Write([]byte("This is a test response with some content"))

	httpAssert.AssertBodyContains(t, w, "test response")
}

func TestAssertBodyNotContains_VerifiesAbsence(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(200)
	w.Write([]byte("This is a test response"))

	httpAssert.AssertBodyNotContains(t, w, "unwanted text")
}

func TestAssertAjaxResponse_PartialHTML(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(200)
	w.Write([]byte(`<div class="content">Partial content</div>`))

	httpAssert.AssertAjaxResponse(t, w, 200)
}

func TestAssertAjaxResponse_RejectsFullHTML(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(200)
	w.Write([]byte(`<!DOCTYPE html><html><body>Full page</body></html>`))

	// This should fail because it contains DOCTYPE
	// We'll wrap in a test that expects failure
	tt := &testing.T{}
	httpAssert.AssertAjaxResponse(tt, w, 200)
	assert.True(t, tt.Failed(), "AssertAjaxResponse should reject full HTML documents")
}

func TestAssertRedirect_VerifiesRedirection(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("Location", "/new-location")
	w.WriteHeader(302)

	httpAssert.AssertRedirect(t, w, "/new-location")
}
