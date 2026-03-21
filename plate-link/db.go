package main

import "plate/link/internal/plate"

func buildDependencies() (*plate.Dependencies, error) {
	cfg, err := plate.LoadConfig()
	if err != nil {
		return nil, err
	}
	return plate.NewDependencies(cfg)
}
