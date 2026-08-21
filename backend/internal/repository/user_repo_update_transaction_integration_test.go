//go:build integration

package repository

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryUpdateJoinsOuterTransaction(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewUserRepository(client, integrationDB)
	user := mustCreateUser(t, client, &service.User{Username: "before-outer-update"})

	t.Cleanup(func() {
		_, _ = integrationDB.Exec(`DELETE FROM auth_identities WHERE user_id = $1`, user.ID)
		_, _ = integrationDB.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
	})

	tx, err := client.Tx(ctx)
	require.NoError(t, err)
	txCtx := dbent.NewTxContext(ctx, tx)

	user.Username = "inside-outer-update"
	require.NoError(t, repo.Update(txCtx, user, service.UserUpdateFields{Username: true}))
	require.NoError(t, tx.Rollback())

	stored, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, "before-outer-update", stored.Username)
}
