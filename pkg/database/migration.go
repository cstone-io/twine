package database

// migrations holds all registered migrations
var migrations = []*Migration{}

// Migration represents a database table migration with dependencies
type Migration struct {
	Model interface{}
	Name  string
	Deps  []*Migration
}

// MigrationBuilder provides a fluent interface for building migrations
type MigrationBuilder struct {
	model interface{}
	name  string
	deps  []*Migration
}

// NewMigrationBuilder creates a new MigrationBuilder instance
func NewMigrationBuilder() *MigrationBuilder {
	return &MigrationBuilder{}
}

// Model sets the model struct for this migration
func (b *MigrationBuilder) Model(model interface{}) *MigrationBuilder {
	b.model = model
	return b
}

// Name sets the name of this migration
func (b *MigrationBuilder) Name(name string) *MigrationBuilder {
	b.name = name
	return b
}

// Deps sets the dependencies for this migration
func (b *MigrationBuilder) Deps(deps ...*Migration) *MigrationBuilder {
	b.deps = deps
	return b
}

// Build constructs the final Migration
func (b *MigrationBuilder) Build() *Migration {
	return &Migration{
		Model: b.model,
		Name:  b.name,
		Deps:  b.deps,
	}
}
