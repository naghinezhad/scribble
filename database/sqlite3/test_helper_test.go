package sqlite3_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/nasermirzaei89/scribble/database/sqlite3"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()

	ctx := context.Background()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())

	db, err := sqlite3.NewDB(ctx, dsn)
	require.NoError(t, err)

	err = sqlite3.MigrateUp(ctx, db)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := sqlite3.MigrateDown(db)
		require.NoError(t, err)

		err = db.Close()
		require.NoError(t, err)
	})

	return ctx, db
}
