package main

import "plate/vec/internal/plate"

func buildDependencies() (*plate.Dependencies, error) {
	cfg, err := plate.LoadConfig()
	if err != nil {
		return nil, err
	}
	return plate.NewDependencies(cfg)
}
