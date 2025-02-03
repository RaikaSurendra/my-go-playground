	package cmd

	import (
	  "encoding/csv"
	  "fmt"
	  "os"
	  "strings"

	  "github.com/spf13/cobra"
	  "sncli/internal/snow"
	)

	var schemaCmd = &cobra.Command{
	  Use:   "schema",
	  Short: "Export table schema for ERD",
	  Long: `Export ServiceNow table schemas and relationships for ERD generation.
	Supports scoped application filtering and outputs in CSV format suitable for tools like Lucidchart.`,
	  RunE: runSchema,
	}

	var (
	  scope    string
	  output   string
	  detailed bool
	)

	func init() {
	  rootCmd.AddCommand(schemaCmd)
	  schemaCmd.Flags().StringVarP(&scope, "scope", "s", "", "Application scope to filter tables (required)")
	  schemaCmd.Flags().StringVarP(&output, "output", "o", "tables.csv", "Output CSV file path")
	  schemaCmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Include detailed field information")
	  schemaCmd.MarkFlagRequired("scope")
	}

	func runSchema(cmd *cobra.Command, args []string) error {
			cfg, err := snow.ReadConfig()
			if err != nil {
			    return fmt.Errorf("failed to read config: %w", err)
			}

			client, err := snow.NewClient(cfg.Instance, cfg.Username, cfg.Password)
	  if err != nil {
	    return fmt.Errorf("failed to create ServiceNow client: %v", err)
	  }

			fmt.Printf("Fetching tables for scope: %s...\n", scope)
			tables, err := client.GetTables(scope, detailed)
			if err != nil {
			  return fmt.Errorf("failed to fetch tables: %w", err)
			}

			fmt.Printf("Found %d tables, fetching relationships...\n", len(tables))
			relationships, err := client.GetRelationships(tables)
			if err != nil {
			  return fmt.Errorf("failed to fetch relationships: %w", err)
			}

	  f, err := os.Create(output)
	  if err != nil {
	    return fmt.Errorf("failed to create output file: %v", err)
	  }
	  defer f.Close()

	  w := csv.NewWriter(f)
	  defer w.Flush()

	  // Write header
			header := []string{"Table Name", "Label", "Description", "Super Class", "Properties", "Parent Relationships", "Referenced Relationships"}
			if detailed {
			  header = append(header, "Fields")
			}
	  if err := w.Write(header); err != nil {
	    return fmt.Errorf("failed to write CSV header: %v", err)
	  }

	  // Write table data
			for _, table := range tables {
			  parentRels := make([]string, 0)
			  referenceRels := make([]string, 0)
			  for _, rel := range relationships {
			    if rel.ParentTable == table.Name {
			      parentRels = append(parentRels, fmt.Sprintf("%s (%s)", rel.ChildTable, rel.Type))
			    }
			    if rel.ReferencedTable == table.Name {
			      referenceRels = append(referenceRels, fmt.Sprintf("%s.%s", rel.SourceTable, rel.Field))
			    }
			  }

			  properties := []string{
			    fmt.Sprintf("Access: %s", table.AccessibleFrom),
			    fmt.Sprintf("Extendable: %v", table.Extendable),
			    fmt.Sprintf("Number Prefix: %s", table.NumberPrefix),
			  }

			  record := []string{
			    table.Name,
			    table.Label,
			    table.Description,
			    table.SuperClass,
			    strings.Join(properties, "\n"),
			    strings.Join(parentRels, "\n"),
			    strings.Join(referenceRels, "\n"),
			  }

			  if detailed {
			    fields := make([]string, 0)
			    for _, f := range table.Fields {
			      fieldProps := []string{
			        fmt.Sprintf("Type: %s", f.Type),
			        fmt.Sprintf("Length: %d", f.Length),
			        fmt.Sprintf("Reference: %s", f.Reference),
			      }
			      fields = append(fields, fmt.Sprintf("%s\n  %s", f.Name, strings.Join(fieldProps, ", ")))
			    }
			    record = append(record, strings.Join(fields, "\n"))
			  }

	    if err := w.Write(record); err != nil {
	      return fmt.Errorf("failed to write table record: %v", err)
	    }
	  }

	  fmt.Printf("Successfully exported schema to %s\n", output)
	  return nil
	}

