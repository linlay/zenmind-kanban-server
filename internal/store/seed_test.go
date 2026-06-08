package store

import (
	"context"
	"testing"
)

func TestSeed(t *testing.T) {
	ctx := context.Background()
	s, err := Open(ctx, t.TempDir()+"/kanban.db")
	if err != nil { t.Fatal(err) }
	defer s.Close()

	err = s.SeedWorkflowCatalog(ctx)
	if err != nil { t.Fatal(err) }
	t.Log("Seed OK")
}