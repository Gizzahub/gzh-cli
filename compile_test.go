package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
)

func main() {
	ctx := context.Background()
	
	// Test basic types
	var repo github.RepositoryInfo
	var config github.RepositoryConfig
	var rate github.RateLimit
	var token github.TokenInfoRecord
	
	fmt.Printf("Types compiled successfully: %T, %T, %T, %T\n", repo, config, rate, token)
	
	// Test interface usage
	var client github.APIClient
	var cloner github.CloneService
	var validator github.TokenValidatorInterface
	
	fmt.Printf("Interfaces compiled successfully: %T, %T, %T\n", client, cloner, validator)
	
	// Test factory functions
	apiConfig := github.DefaultAPIClientConfig()
	cloneConfig := github.DefaultCloneServiceConfig()
	
	fmt.Printf("Factory functions compiled successfully: %T, %T\n", apiConfig, cloneConfig)
	
	fmt.Println("All compilation tests passed!")
	os.Exit(0)
}