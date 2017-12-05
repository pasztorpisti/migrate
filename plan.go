package migrate

import (
	"fmt"
	"sort"
	"strconv"
)

const (
	Initial = "initial"
	Latest  = "latest"
)

type PlanInput struct {
	Migrations           []*Migration
	ForwardMigratedNames []string
	// Target is either the full name of a migration (file) or only it's prefix
	// that is a non-negative integer.
	// It can also be one of the following constants: Initial, Latest
	Target      string
	MigrationDB MigrationDB
}

func Plan(input *PlanInput) (Steps, error) {
	// First we create a list of migration items in which each item has a
	// unique ID and belongs to one of the following 3 categories:
	//
	// 1. The migration exists and it has been forward migrated
	// 2. The migration exists but it hasn't yet been forward migrated
	//    (or it has been backward migrated)
	// 3. The migration doesn't exist but it has been forward migrated in the
	//    past probably on a different branch of your repo.

	type item struct {
		id              MigrationID
		migration       *Migration // nil if not exist
		forwardMigrated bool
	}

	migrations := input.Migrations
	items := make([]*item, len(migrations))
	idSet := make(map[int64]*item, len(migrations))
	for i, migration := range migrations {
		var id MigrationID
		if err := id.SetName(migration.Name); err != nil {
			return nil, fmt.Errorf("error parsing migration name: %s", err)
		}

		_, ok := idSet[id.Number]
		if ok {
			return nil, fmt.Errorf("duplicate migration id: %v", id.Number)
		}

		it := &item{
			id:        id,
			migration: migration,
		}
		idSet[id.Number] = it
		items[i] = it
	}

	for _, name := range input.ForwardMigratedNames {
		var id MigrationID
		if err := id.SetName(name); err != nil {
			return nil, fmt.Errorf("error processing forward migration name: %s", err)
		}

		if item, ok := idSet[id.Number]; ok {
			if item.id.Name != name {
				return nil, fmt.Errorf("a migration ID (%q) is conflicting with a previously forward migraited ID (%q)", item.id.Name, name)
			}
			item.forwardMigrated = true
			continue
		}

		it := &item{
			id:              id,
			forwardMigrated: true,
		}
		idSet[id.Number] = it
		items = append(items, it)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].id.Number < items[j].id.Number
	})

	// Finding the index of the target migration in the sorted item list.
	var targetIdx int
	switch input.Target {
	case Initial:
		targetIdx = -1
	case Latest:
		targetIdx = len(items) - 1
	default:
		targetIdx = -1
		numericTarget, err := strconv.ParseInt(input.Target, 10, 64)
		if err != nil {
			numericTarget = -1
		}
		for i, item := range items {
			if item.id.Number == numericTarget || item.id.Name == input.Target {
				targetIdx = i
				break
			}
		}
		if targetIdx == -1 {
			return nil, fmt.Errorf("invalid target migration: %v", input.Target)
		}
	}

	var steps Steps

	// Backward-migrating items that are newer than the target item in reverse order.
	for i := len(items) - 1; i > targetIdx; i-- {
		item := items[i]
		if !item.forwardMigrated {
			continue
		}
		if item.migration == nil {
			return nil, fmt.Errorf("don't know how to backward-migrate %q (missing migration file?)", item.id.Name)
		}
		if item.migration.Backward == nil {
			return nil, fmt.Errorf("%q can't be backward migrated", item.id.Name)
		}

		updateMetaStep, err := input.MigrationDB.BackwardMigrate(item.id.Name)
		if err != nil {
			return nil, err
		}

		s := Steps{
			item.migration.Backward,
			updateMetaStep,
		}
		steps = append(steps, &StepTitleAndResult{
			Step:  TransactionIfAllowed{s},
			Title: "backward-migrate " + item.id.Name,
		})
	}

	// Forward-migrating items that are older than or equal to the target item.
	for i := 0; i <= targetIdx; i++ {
		item := items[i]
		if item.forwardMigrated {
			continue
		}
		if item.migration == nil {
			return nil, fmt.Errorf("don't know how to forward-migrate %q (missing migration file?)", item.id.Name)
		}

		updateMetaStep, err := input.MigrationDB.ForwardMigrate(item.id.Name)
		if err != nil {
			return nil, err
		}

		s := Steps{
			item.migration.Forward,
			updateMetaStep,
		}
		steps = append(steps, &StepTitleAndResult{
			Step:  TransactionIfAllowed{s},
			Title: "forward-migrate " + item.id.Name,
		})
	}

	return steps, nil
}
