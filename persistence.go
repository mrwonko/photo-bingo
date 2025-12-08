package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

func saveState(ctx context.Context, trigger <-chan struct{}) {
	exit := false
	for !exit {
		select {
		case <-ctx.Done():
			logf("State saving canceled; performing final save.")
			exit = true
			// still perform one last save (though it should not be necessary)
		case <-trigger:
			logf("State saving triggered.")
			// drain trigger channel
			drained := false
			for !drained {
				select {
				case <-trigger:
				default:
					drained = true
				}
			}
		}
		var stateJSON []byte
		var err error
		gameState.Read(func(gs GameState) {
			stateJSON, err = json.Marshal(gs)
		})
		if err != nil {
			log.Printf("failed to marshal game state: %s", err)
			continue
		}
		_, err = os.Stat(latestStatePath)
		if err == nil {
			// move latest state to backup, replacing it (FIXME: this assumes the latest state was written successfuly, which may not necessarily be the case)
			err = os.Remove(previousStatePath)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				log.Printf("failed to remove state backup %q: %s", previousStatePath, err)
			}
			err = os.Rename(latestStatePath, previousStatePath)
			if err != nil {
				log.Printf("failed to back up previous state %q to %q: %s", latestStatePath, previousStatePath, err)
				// still continue and overwrite it, we care more about losing the latest updates than the backup
			}
		}
		err = os.WriteFile(latestStatePath, stateJSON, 0600)
		if err != nil {
			log.Printf("failed to write state file %q: %s", latestStatePath, err)
			continue
		}
		logf("state successfully saved to %q", latestStatePath)
	}
}

func loadState() error {
	var loadedState GameState
	stateJSON, err := os.ReadFile(latestStatePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// first launch / no saved state -> done
			return nil
		}
		return fmt.Errorf("reading state file %q: %w", latestStatePath, err)
	}
	err = json.Unmarshal(stateJSON, &loadedState)
	if err != nil {
		return fmt.Errorf("unmarshaling state file %q: %w", latestStatePath, err)
	}
	gameState.Modify(func(gs GameState) GameState {
		return loadedState
	})
	logf("game state loaded from %q", latestStatePath)
	return nil
}
