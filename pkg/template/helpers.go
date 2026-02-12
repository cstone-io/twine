package template

import (
	"html/template"
	"time"
)

// FuncMap returns the default template functions
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"formatDate":     formatDate,
		"formatDateTime": formatDateTime,
		"add":            add,
		"sub":            sub,
		"mul":            mul,
		"div":            div,
		"mod":            mod,
		"eq":             eq,
		"ne":             ne,
		"lt":             lt,
		"le":             le,
		"gt":             gt,
		"ge":             ge,
		"asset":          asset,
	}
}

// formatDate formats a time.Time as a date string
func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// formatDateTime formats a time.Time as a date-time string
func formatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// Math functions
func add(a, b int) int      { return a + b }
func sub(a, b int) int      { return a - b }
func mul(a, b int) int      { return a * b }
func div(a, b int) int      { return a / b }
func mod(a, b int) int      { return a % b }

// Comparison functions
func eq(a, b any) bool  { return a == b }
func ne(a, b any) bool  { return a != b }
func lt(a, b int) bool  { return a < b }
func le(a, b int) bool  { return a <= b }
func gt(a, b int) bool  { return a > b }
func ge(a, b int) bool  { return a >= b }

// asset returns the path to a static asset
func asset(name string) string {
	return "/public/assets/" + name
}
