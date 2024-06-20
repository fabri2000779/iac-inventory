package main

import (
	"context"
	"drifter/aws"
	"drifter/controller"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()
	cfg, err := aws.LoadAWSConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	drft, err := controller.Run(cfg)
	if err != nil {
		log.Fatalf("controller run failed: %v", err)
	}

	fmt.Println(string(drft))
}
