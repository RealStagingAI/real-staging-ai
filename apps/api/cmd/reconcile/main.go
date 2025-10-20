package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/real-staging-ai/api/internal/config"
	"github.com/real-staging-ai/api/internal/logging"
	"github.com/real-staging-ai/api/internal/reconcile"
	"github.com/real-staging-ai/api/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "images":
		runReconcileImages(os.Args[2:])
	case "cleanup":
		runCleanup(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Reconcile CLI - Database and storage reconciliation tools")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  reconcile <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  images   Check and fix image storage inconsistencies")
	fmt.Println("  cleanup  Delete stuck queued images")
	fmt.Println()
	fmt.Println("Run 'reconcile <command> -help' for command-specific options")
}

func runReconcileImages(args []string) {
	fs := flag.NewFlagSet("images", flag.ExitOnError)
	var (
		batchSize   = fs.Int("batch-size", 100, "Number of images to check per batch")
		concurrency = fs.Int("concurrency", 5, "Number of concurrent S3 checks")
		dryRun      = fs.Bool("dry-run", false, "Don't apply changes, only report what would be done")
		projectID   = fs.String("project-id", "", "Optional: filter by project ID")
		status      = fs.String("status", "", "Optional: filter by status (queued, processing, ready, error)")
	)
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	logger := logging.Default()

	logger.Info(ctx, "starting reconciliation CLI")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error(ctx, "failed to load configuration", "error", err)
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	db, err := storage.NewDefaultDatabase(&cfg.DB)
	if err != nil {
		logger.Error(ctx, "failed to connect to database", "error", err)
		fmt.Fprintf(os.Stderr, "Error: failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Debug(ctx, "database connection established")

	logger.Debug(ctx, "initializing S3 service", "bucket", cfg.S3.BucketName)
	s3Service, err := storage.NewDefaultS3Service(ctx, &cfg.S3)
	if err != nil {
		logger.Error(ctx, "failed to initialize S3 service", "error", err, "bucket", cfg.S3.BucketName)
		fmt.Fprintf(os.Stderr, "Error: failed to initialize S3 service: %v\n", err)
		return
	}

	// Create reconcile service
	svc := reconcile.NewDefaultService(db, s3Service)

	// Build options
	opts := reconcile.ReconcileOptions{
		Limit:       *batchSize,
		Concurrency: *concurrency,
		DryRun:      *dryRun,
	}
	if *projectID != "" {
		opts.ProjectID = projectID
	}
	if *status != "" {
		opts.Status = status
	}

	logger.Info(ctx, "starting reconciliation run",
		"dry_run", *dryRun,
		"batch_size", *batchSize,
		"concurrency", *concurrency,
		"project_id", opts.ProjectID,
		"status", opts.Status,
	)

	fmt.Printf("Starting reconciliation (dry_run=%v, batch_size=%d, concurrency=%d)\n", *dryRun, *batchSize, *concurrency)
	if opts.ProjectID != nil {
		fmt.Printf("  Filtering by project_id: %s\n", *opts.ProjectID)
	}
	if opts.Status != nil {
		fmt.Printf("  Filtering by status: %s\n", *opts.Status)
	}

	// Run reconciliation
	result, err := svc.ReconcileImages(ctx, opts)
	if err != nil {
		logger.Error(ctx, "reconciliation failed", "error", err)
		fmt.Fprintf(os.Stderr, "Error: reconciliation failed: %v\n", err)
		return
	}

	logger.Info(ctx, "reconciliation completed",
		"checked", result.Checked,
		"missing_original", result.MissingOrig,
		"missing_staged", result.MissingStaged,
		"updated", result.Updated,
		"dry_run", result.DryRun,
	)

	// Print results
	fmt.Println("\nReconciliation Results:")
	fmt.Printf("  Checked:         %d images\n", result.Checked)
	fmt.Printf("  Missing original: %d\n", result.MissingOrig)
	fmt.Printf("  Missing staged:   %d\n", result.MissingStaged)
	fmt.Printf("  Updated:         %d\n", result.Updated)
	fmt.Printf("  Dry run:         %v\n", result.DryRun)

	if len(result.Examples) > 0 {
		fmt.Println("\nExample errors (up to 10):")
		for _, ex := range result.Examples {
			fmt.Printf("  - Image %s (status=%s): %s\n", ex.ImageID, ex.Status, ex.Error)
		}
	}

	// Output JSON for scripting
	if jsonOutput := os.Getenv("JSON_OUTPUT"); jsonOutput == "1" {
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println("\nJSON Output:")
		fmt.Println(string(jsonBytes))
	}

	if !*dryRun && result.Updated > 0 {
		fmt.Println("\nNote: Changes have been applied to the database.")
	} else if *dryRun && result.Updated > 0 {
		fmt.Println("\nNote: This was a dry run. No changes were applied.")
	}
}

func runCleanup(args []string) {
	fs := flag.NewFlagSet("cleanup", flag.ExitOnError)
	var (
		olderThanHours = fs.Int("older-than-hours", 1, "Delete images stuck in queued status for more than this many hours")
	)
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	logger := logging.Default()

	logger.Info(ctx, "starting cleanup CLI", "older_than_hours", *olderThanHours)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error(ctx, "failed to load configuration", "error", err)
		fmt.Fprintf(os.Stderr, "Error: failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	db, err := storage.NewDefaultDatabase(&cfg.DB)
	if err != nil {
		logger.Error(ctx, "failed to connect to database", "error", err)
		fmt.Fprintf(os.Stderr, "Error: failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Debug(ctx, "database connection established")

	// Initialize S3 service (needed for service interface, but won't be used for cleanup)
	s3Service, err := storage.NewDefaultS3Service(ctx, &cfg.S3)
	if err != nil {
		logger.Error(ctx, "failed to initialize S3 service", "error", err)
		fmt.Fprintf(os.Stderr, "Error: failed to initialize S3 service: %v\n", err)
		db.Close()
		os.Exit(1) //nolint:gocritic // db.Close() called explicitly before exit
	}

	// Create reconcile service
	svc := reconcile.NewDefaultService(db, s3Service)

	fmt.Printf("Cleaning up images stuck in queued status for more than %d hour(s)...\n", *olderThanHours)

	// Run cleanup
	result, err := svc.CleanupStuckQueuedImages(ctx, *olderThanHours)
	if err != nil {
		logger.Error(ctx, "cleanup failed", "error", err)
		fmt.Fprintf(os.Stderr, "Error: cleanup failed: %v\n", err)
		db.Close()
		os.Exit(1)
	}

	logger.Info(ctx, "cleanup completed",
		"deleted", result.Deleted,
		"threshold", result.Threshold,
	)

	// Print results
	fmt.Println("\nCleanup Results:")
	fmt.Printf("  Deleted:   %d images\n", result.Deleted)
	fmt.Printf("  Threshold: %s\n", result.Threshold)

	if len(result.ImageIDs) > 0 {
		fmt.Println("\nDeleted Image IDs:")
		for _, id := range result.ImageIDs {
			fmt.Printf("  - %s\n", id)
		}
	}

	// Output JSON for scripting
	if jsonOutput := os.Getenv("JSON_OUTPUT"); jsonOutput == "1" {
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println("\nJSON Output:")
		fmt.Println(string(jsonBytes))
	}

	if result.Deleted > 0 {
		fmt.Println("\nNote: Images have been deleted from the database.")
	} else {
		fmt.Println("\nNote: No stuck queued images found.")
	}
}
