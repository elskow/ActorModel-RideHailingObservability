package utils

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// SetupMockDB creates a mock database connection for testing
func SetupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, sqlMock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	return sqlxDB, sqlMock
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}
