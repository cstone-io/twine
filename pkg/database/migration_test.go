package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMigration_NewMigrationBuilder tests the migration builder
func TestMigration_NewMigrationBuilder(t *testing.T) {
	builder := NewMigrationBuilder()

	assert.NotNil(t, builder)
}

// TestMigration_Builder tests the builder pattern
func TestMigration_Builder(t *testing.T) {
	type TestTable struct {
		ID   int
		Name string
	}

	t.Run("builds migration with all fields", func(t *testing.T) {
		dep := &Migration{Name: "dependency"}

		migration := NewMigrationBuilder().
			Model(&TestTable{}).
			Name("test_table").
			Deps(dep).
			Build()

		assert.NotNil(t, migration)
		assert.NotNil(t, migration.Model)
		assert.Equal(t, "test_table", migration.Name)
		assert.Len(t, migration.Deps, 1)
		assert.Equal(t, dep, migration.Deps[0])
	})

	t.Run("builds migration without dependencies", func(t *testing.T) {
		migration := NewMigrationBuilder().
			Model(&TestTable{}).
			Name("test_table").
			Build()

		assert.NotNil(t, migration)
		assert.Equal(t, "test_table", migration.Name)
		assert.Nil(t, migration.Deps)
	})

	t.Run("builds migration with multiple dependencies", func(t *testing.T) {
		dep1 := &Migration{Name: "dep1"}
		dep2 := &Migration{Name: "dep2"}
		dep3 := &Migration{Name: "dep3"}

		migration := NewMigrationBuilder().
			Model(&TestTable{}).
			Name("test_table").
			Deps(dep1, dep2, dep3).
			Build()

		assert.Len(t, migration.Deps, 3)
		assert.Equal(t, "dep1", migration.Deps[0].Name)
		assert.Equal(t, "dep2", migration.Deps[1].Name)
		assert.Equal(t, "dep3", migration.Deps[2].Name)
	})

	t.Run("builder methods return builder for chaining", func(t *testing.T) {
		builder := NewMigrationBuilder()

		result := builder.Model(&TestTable{})
		assert.Equal(t, builder, result)

		result = builder.Name("test")
		assert.Equal(t, builder, result)

		result = builder.Deps()
		assert.Equal(t, builder, result)
	})
}

// TestMigration_Struct tests the Migration struct
func TestMigration_Struct(t *testing.T) {
	type TestTable struct {
		ID int
	}

	t.Run("stores model correctly", func(t *testing.T) {
		model := &TestTable{}
		migration := &Migration{
			Model: model,
			Name:  "test",
		}

		assert.Equal(t, model, migration.Model)
	})

	t.Run("stores name correctly", func(t *testing.T) {
		migration := &Migration{
			Model: &TestTable{},
			Name:  "users",
		}

		assert.Equal(t, "users", migration.Name)
	})

	t.Run("stores dependencies correctly", func(t *testing.T) {
		dep := &Migration{Name: "dependency"}
		migration := &Migration{
			Model: &TestTable{},
			Name:  "test",
			Deps:  []*Migration{dep},
		}

		assert.Len(t, migration.Deps, 1)
		assert.Equal(t, dep, migration.Deps[0])
	})
}

// TestRegisterMigration tests single migration registration
func TestRegisterMigration(t *testing.T) {
	// Save original migrations and restore after test
	originalMigrations := migrations
	defer func() { migrations = originalMigrations }()

	t.Run("registers migration", func(t *testing.T) {
		migrations = []*Migration{} // Reset

		migration := &Migration{Name: "test"}
		RegisterMigration(migration)

		assert.Len(t, migrations, 1)
		assert.Equal(t, "test", migrations[0].Name)
	})

	t.Run("appends to existing migrations", func(t *testing.T) {
		migrations = []*Migration{
			{Name: "existing"},
		}

		migration := &Migration{Name: "new"}
		RegisterMigration(migration)

		assert.Len(t, migrations, 2)
		assert.Equal(t, "existing", migrations[0].Name)
		assert.Equal(t, "new", migrations[1].Name)
	})
}

// TestRegisterMigrations tests multiple migration registration
func TestRegisterMigrations(t *testing.T) {
	// Save original migrations and restore after test
	originalMigrations := migrations
	defer func() { migrations = originalMigrations }()

	t.Run("registers multiple migrations", func(t *testing.T) {
		migrations = []*Migration{} // Reset

		m1 := &Migration{Name: "m1"}
		m2 := &Migration{Name: "m2"}
		m3 := &Migration{Name: "m3"}

		RegisterMigrations(m1, m2, m3)

		assert.Len(t, migrations, 3)
		assert.Equal(t, "m1", migrations[0].Name)
		assert.Equal(t, "m2", migrations[1].Name)
		assert.Equal(t, "m3", migrations[2].Name)
	})

	t.Run("appends to existing migrations", func(t *testing.T) {
		migrations = []*Migration{
			{Name: "existing"},
		}

		m1 := &Migration{Name: "new1"}
		m2 := &Migration{Name: "new2"}

		RegisterMigrations(m1, m2)

		assert.Len(t, migrations, 3)
		assert.Equal(t, "existing", migrations[0].Name)
		assert.Equal(t, "new1", migrations[1].Name)
		assert.Equal(t, "new2", migrations[2].Name)
	})

	t.Run("handles empty registration", func(t *testing.T) {
		migrations = []*Migration{
			{Name: "existing"},
		}

		RegisterMigrations()

		assert.Len(t, migrations, 1)
	})
}

// TestMigration_DependencyStructure tests various dependency configurations
func TestMigration_DependencyStructure(t *testing.T) {
	t.Run("linear dependency chain", func(t *testing.T) {
		m1 := &Migration{Name: "m1"}
		m2 := &Migration{Name: "m2", Deps: []*Migration{m1}}
		m3 := &Migration{Name: "m3", Deps: []*Migration{m2}}

		assert.Nil(t, m1.Deps)
		assert.Equal(t, []*Migration{m1}, m2.Deps)
		assert.Equal(t, []*Migration{m2}, m3.Deps)
	})

	t.Run("multiple dependencies", func(t *testing.T) {
		m1 := &Migration{Name: "m1"}
		m2 := &Migration{Name: "m2"}
		m3 := &Migration{Name: "m3", Deps: []*Migration{m1, m2}}

		assert.Len(t, m3.Deps, 2)
		assert.Contains(t, m3.Deps, m1)
		assert.Contains(t, m3.Deps, m2)
	})

	t.Run("diamond dependency", func(t *testing.T) {
		//     m4
		//    /  \
		//   m2  m3
		//    \  /
		//     m1
		m1 := &Migration{Name: "m1"}
		m2 := &Migration{Name: "m2", Deps: []*Migration{m1}}
		m3 := &Migration{Name: "m3", Deps: []*Migration{m1}}
		m4 := &Migration{Name: "m4", Deps: []*Migration{m2, m3}}

		assert.Len(t, m4.Deps, 2)
		assert.Contains(t, m4.Deps, m2)
		assert.Contains(t, m4.Deps, m3)
		assert.Equal(t, m1, m2.Deps[0])
		assert.Equal(t, m1, m3.Deps[0])
	})

	t.Run("complex dependency graph", func(t *testing.T) {
		//      m5
		//     / |
		//   m3  m4
		//   |   |
		//   m1  m2
		m1 := &Migration{Name: "m1"}
		m2 := &Migration{Name: "m2"}
		m3 := &Migration{Name: "m3", Deps: []*Migration{m1}}
		m4 := &Migration{Name: "m4", Deps: []*Migration{m2}}
		m5 := &Migration{Name: "m5", Deps: []*Migration{m3, m4}}

		assert.Len(t, m5.Deps, 2)
		assert.Equal(t, "m3", m5.Deps[0].Name)
		assert.Equal(t, "m4", m5.Deps[1].Name)
	})
}

// TestMigration_Integration tests realistic migration scenarios
func TestMigration_Integration(t *testing.T) {
	t.Run("typical web app migrations", func(t *testing.T) {
		type User struct {
			BaseModel
			Email string
		}

		type Post struct {
			BaseModel
			UserID string
			Title  string
		}

		type Comment struct {
			BaseModel
			PostID  string
			UserID  string
			Content string
		}

		// Users table has no dependencies
		userMigration := NewMigrationBuilder().
			Model(&User{}).
			Name("users").
			Build()

		// Posts depend on users (for foreign key)
		postMigration := NewMigrationBuilder().
			Model(&Post{}).
			Name("posts").
			Deps(userMigration).
			Build()

		// Comments depend on both posts and users
		commentMigration := NewMigrationBuilder().
			Model(&Comment{}).
			Name("comments").
			Deps(postMigration, userMigration).
			Build()

		assert.Equal(t, "users", userMigration.Name)
		assert.Nil(t, userMigration.Deps)

		assert.Equal(t, "posts", postMigration.Name)
		assert.Len(t, postMigration.Deps, 1)

		assert.Equal(t, "comments", commentMigration.Name)
		assert.Len(t, commentMigration.Deps, 2)
	})

	t.Run("polymorphic relationship migrations", func(t *testing.T) {
		type Tag struct {
			BaseModel
			Name string
		}

		type Taggable struct {
			TagID        string
			TaggableID   string
			TaggableType string
		}

		type Article struct {
			BaseModel
			Title string
		}

		// Tags and articles can be created independently
		tagMigration := &Migration{Model: &Tag{}, Name: "tags"}
		articleMigration := &Migration{Model: &Article{}, Name: "articles"}

		// Taggable join table depends on both
		taggableMigration := &Migration{
			Model: &Taggable{},
			Name:  "taggables",
			Deps:  []*Migration{tagMigration, articleMigration},
		}

		assert.Len(t, taggableMigration.Deps, 2)
	})

	t.Run("multi-tenant migrations", func(t *testing.T) {
		type Tenant struct {
			BaseModel
			Name string
		}

		type TenantUser struct {
			BaseModel
			TenantID string
			Email    string
		}

		type TenantData struct {
			BaseModel
			TenantID string
			Data     string
		}

		tenantMigration := &Migration{Model: &Tenant{}, Name: "tenants"}

		// Both depend on tenant
		userMigration := &Migration{
			Model: &TenantUser{},
			Name:  "tenant_users",
			Deps:  []*Migration{tenantMigration},
		}

		dataMigration := &Migration{
			Model: &TenantData{},
			Name:  "tenant_data",
			Deps:  []*Migration{tenantMigration},
		}

		assert.Equal(t, tenantMigration, userMigration.Deps[0])
		assert.Equal(t, tenantMigration, dataMigration.Deps[0])
	})
}

// TestMigration_BuilderValidation tests edge cases in builder
func TestMigration_BuilderValidation(t *testing.T) {
	t.Run("empty migration", func(t *testing.T) {
		migration := NewMigrationBuilder().Build()

		assert.NotNil(t, migration)
		assert.Nil(t, migration.Model)
		assert.Empty(t, migration.Name)
		assert.Nil(t, migration.Deps)
	})

	t.Run("nil model", func(t *testing.T) {
		migration := NewMigrationBuilder().
			Model(nil).
			Name("test").
			Build()

		assert.Nil(t, migration.Model)
		assert.Equal(t, "test", migration.Name)
	})

	t.Run("empty name", func(t *testing.T) {
		type TestTable struct{}

		migration := NewMigrationBuilder().
			Model(&TestTable{}).
			Name("").
			Build()

		assert.NotNil(t, migration.Model)
		assert.Empty(t, migration.Name)
	})

	t.Run("nil dependencies", func(t *testing.T) {
		type TestTable struct{}

		migration := NewMigrationBuilder().
			Model(&TestTable{}).
			Name("test").
			Deps(nil).
			Build()

		assert.NotNil(t, migration)
		// Deps() with nil should still set the deps field
		assert.NotNil(t, migration.Deps)
		assert.Len(t, migration.Deps, 1)
		assert.Nil(t, migration.Deps[0])
	})
}

// TestMigration_GlobalState tests the global migrations variable
func TestMigration_GlobalState(t *testing.T) {
	// Save and restore original state
	originalMigrations := migrations
	defer func() { migrations = originalMigrations }()

	t.Run("migrations is package-level variable", func(t *testing.T) {
		migrations = []*Migration{}

		m1 := &Migration{Name: "m1"}
		m2 := &Migration{Name: "m2"}

		RegisterMigration(m1)
		RegisterMigration(m2)

		// Both should be in the same slice
		assert.Len(t, migrations, 2)
	})

	t.Run("multiple register calls accumulate", func(t *testing.T) {
		migrations = []*Migration{}

		for i := 0; i < 5; i++ {
			RegisterMigration(&Migration{Name: "migration"})
		}

		assert.Len(t, migrations, 5)
	})
}
