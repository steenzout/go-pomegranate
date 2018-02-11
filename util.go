package pomegranate

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// This file should contain only private, mostly pure functions.  They should
// not interact with the filesystem or database.

func nameInMigrationList(name string, migrations []Migration) bool {
	for _, mig := range migrations {
		if name == mig.Name {
			return true
		}
	}
	return false
}

func nameInHistory(name string, history []MigrationRecord) bool {
	for _, mig := range history {
		if name == mig.Name {
			return true
		}
	}
	return false
}

func getConfirm(toRun []Migration, forwardBack string) error {
	names := []string{}
	for _, mig := range toRun {
		names = append(names, mig.Name)
	}
	fmt.Printf(
		"%s migrations that will be run:\n%s\nRun these migrations? (y/n) ",
		forwardBack,
		strings.Join(names, "\n"),
	)
	reader := bufio.NewReader(os.Stdin)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	resp = strings.TrimSpace(resp)
	if resp == "y" {
		return nil
	} else {
		fmt.Printf("Invalid option: %s\n", resp)
	}
	return errors.New("migration cancelled")
}

// getForwardMigrations takes a history of already run migrations, and the list
// of all migrations, and returns all that haven't been run yet.  Error if the
// history is out of sync with the allMigrations list.
func getForwardMigrations(history []MigrationRecord, allMigrations []Migration) ([]Migration, error) {
	historyCount := len(history)
	migCount := len(allMigrations)
	if historyCount > migCount {
		return nil, errors.New("migration history longer than static list")
	}

	for i := 0; i < historyCount; i++ {
		if history[i].Name != allMigrations[i].Name {
			return nil, fmt.Errorf(
				"migration %d from history (%s) does not match name from static list (%s)",
				i+1, history[i].Name, allMigrations[i].Name,
			)
		}
	}
	return allMigrations[historyCount:], nil
}

func trimMigrationsTail(newtail string, migrations []Migration) ([]Migration, error) {
	trimmed := []Migration{}
	for _, mig := range migrations {
		trimmed = append(trimmed, mig)
		if mig.Name == newtail {
			return trimmed, nil
		}
	}
	return nil, fmt.Errorf("migration %s not found", newtail)
}

// getMigrationsToReverse takes the name that you're rolling back to, history of
// all migrations run so far, and an ordered list of all possible migrations.
func getMigrationsToReverse(name string, history []MigrationRecord, allMigrations []Migration) ([]Migration, error) {
	// get name of most recent migration
	latest := history[len(history)-1].Name
	// trim allMigrations to ignore anything newer than latest in history.
	reversableMigrations, err := trimMigrationsTail(latest, allMigrations)
	if err != nil {
		return nil, err
	}

	// reversableMigrations and history should now be the same length
	if le, lh := len(reversableMigrations), len(history); le != lh {
		return nil, fmt.Errorf(
			"history in DB has %d migrations, but we have source for %d migrations up to and including %s",
			lh, le, latest,
		)
	}
	// loop backward over history and allmigrations, asserting that names match,
	// and building list of migrations that need running, until we get to the name
	// we're looking for.
	// If we fall off the end, error.
	toRun := []Migration{}
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Name != reversableMigrations[i].Name {
			return nil, fmt.Errorf(
				"migration %d from history (%s) does not match name from static list (%s)",
				i+1, history[i].Name, reversableMigrations[i].Name,
			)
		}
		toRun = append(toRun, reversableMigrations[i])
		if history[i].Name == name {
			return toRun, nil
		}
	}
	return nil, fmt.Errorf("migration %s not found", name)
}
