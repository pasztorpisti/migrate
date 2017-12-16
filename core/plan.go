package core

import "fmt"

const (
	Initial = "initial"
	Latest  = "latest"
)

type PlanInput struct {
	Migrations MigrationEntries
	// ForwardMigrated has the length of Migrations.NumMigrations().
	ForwardMigrated []bool
	// Target is either the full name of a migration (file) or only it's prefix
	// that is a non-negative integer.
	// It can also be one of the following constants: Initial, Latest
	Target      string
	MigrationDB MigrationDB
}

func Plan(input *PlanInput) (Steps, error) {
	numMigrations := input.Migrations.NumMigrations()
	if len(input.ForwardMigrated) != numMigrations {
		panic("len(forwardMigrated) != numMigrations")
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
		if !input.ForwardMigrated[i] {
			continue
		}
		name := input.Migrations.Name(i)
		_, backwardStep, err := input.Migrations.Steps(i)
		if err != nil {
			return nil, fmt.Errorf("error loading backward step for migration %q", name)
		}
		if backwardStep == nil {
			return nil, fmt.Errorf("%q doesn't have a backward step", name)
		}

		updateSystemStep, err := input.MigrationDB.BackwardMigrate(name)
		if err != nil {
			return nil, err
		}

		s := Steps{
			backwardStep,
			updateSystemStep,
		}
		steps = append(steps, &StepTitleAndResult{
			Step:  TransactionIfAllowed{s},
			Title: "backward-migrate " + name,
		})
	}

	// Forward-migrating items that are older than or equal to the target item.
	for i := 0; i <= targetIdx; i++ {
		if input.ForwardMigrated[i] {
			continue
		}

		name := input.Migrations.Name(i)
		forwardStep, _, err := input.Migrations.Steps(i)
		if err != nil {
			return nil, fmt.Errorf("error loading forward step for migration %q", name)
		}

		updateSystemStep, err := input.MigrationDB.ForwardMigrate(name)
		if err != nil {
			return nil, err
		}

		s := Steps{
			forwardStep,
			updateSystemStep,
		}
		steps = append(steps, &StepTitleAndResult{
			Step:  TransactionIfAllowed{s},
			Title: "forward-migrate " + name,
		})
	}

	return steps, nil
}
