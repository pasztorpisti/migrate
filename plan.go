package migrate

import (
	"errors"
	"fmt"
)

var ErrMissingPastMigrations = errors.New("there are unapplied migrations before the latest applied migration")

const (
	Initial = "initial"
	Latest  = "latest"
)

type PlanInput struct {
	Migrations           MigrationEntries
	ForwardMigratedNames []string
	// Target is either the full name of a migration (file) or only it's prefix
	// that is a non-negative integer.
	// It can also be one of the following constants: Initial, Latest
	Target      string
	MigrationDB MigrationDB
}

func Plan(input *PlanInput) (Steps, error) {
	numMigrations := input.Migrations.NumMigrations()

	forwardMigrated := make([]bool, numMigrations)
	for _, name := range input.ForwardMigratedNames {
		index, ok := input.Migrations.IndexForName(name)
		// We don't accept aliases as forward migrated names.
		// This is why we check for (name != input.Migrations.Name(index)).
		if !ok || name != input.Migrations.Name(index) {
			return nil, fmt.Errorf("can't find migration file for forward migrated item %q", name)
		}
		forwardMigrated[index] = true
	}

	if !input.Migrations.AllowsPastMigrations() {
		allowForwardMigrated := true
		for _, fm := range forwardMigrated {
			if fm {
				if !allowForwardMigrated {
					return nil, ErrMissingPastMigrations
				}
			} else {
				allowForwardMigrated = false
			}
		}
	}

	// Finding the index of the target migration in the sorted item list.
	var targetIdx int
	switch input.Target {
	case Initial:
		targetIdx = -1
	case Latest:
		targetIdx = numMigrations - 1
	default:
		if i, ok := input.Migrations.IndexForName(input.Target); ok {
			targetIdx = i
		} else {
			return nil, fmt.Errorf("invalid target migration: %v", input.Target)
		}
	}

	var steps Steps

	// Backward-migrating items that are newer than the target item in reverse order.
	for i := numMigrations - 1; i > targetIdx; i-- {
		if !forwardMigrated[i] {
			continue
		}
		name := input.Migrations.Name(i)
		_, backwardStep, err := input.Migrations.Steps(i)
		if err != nil {
			fmt.Errorf("error loading backward step for migration %q", name)
		}
		if backwardStep == nil {
			return nil, fmt.Errorf("%q doesn't have a backward step", name)
		}

		updateMetaStep, err := input.MigrationDB.BackwardMigrate(name)
		if err != nil {
			return nil, err
		}

		s := Steps{
			backwardStep,
			updateMetaStep,
		}
		steps = append(steps, &StepTitleAndResult{
			Step:  TransactionIfAllowed{s},
			Title: "backward-migrate " + name,
		})
	}

	// Forward-migrating items that are older than or equal to the target item.
	for i := 0; i <= targetIdx; i++ {
		if forwardMigrated[i] {
			continue
		}

		name := input.Migrations.Name(i)
		forwardStep, _, err := input.Migrations.Steps(i)
		if err != nil {
			fmt.Errorf("error loading forward step for migration %q", name)
		}

		updateMetaStep, err := input.MigrationDB.ForwardMigrate(name)
		if err != nil {
			return nil, err
		}

		s := Steps{
			forwardStep,
			updateMetaStep,
		}
		steps = append(steps, &StepTitleAndResult{
			Step:  TransactionIfAllowed{s},
			Title: "forward-migrate " + name,
		})
	}

	return steps, nil
}
