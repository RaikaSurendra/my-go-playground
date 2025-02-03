	package cmd

	      import (
	              "fmt"
	              "strings"
	              "time"

	              "github.com/briandowns/spinner"
	              "github.com/charmbracelet/lipgloss"
	              "github.com/spf13/cobra"
	              "sncli/internal/snow"
	              "sncli/internal/tui"
	      )


	var (
	        successStyle = lipgloss.NewStyle().
	                Foreground(lipgloss.Color("42")).
	                Bold(true)
	        errorStyle = lipgloss.NewStyle().
	                Foreground(lipgloss.Color("196")).
	                Bold(true)
	        infoStyle = lipgloss.NewStyle().
	                Foreground(lipgloss.Color("87"))
	)

	var connectCmd = &cobra.Command{
	  Use:   "connect",
	  Short: "Connect to a ServiceNow instance",
			Run: func(cmd *cobra.Command, args []string) {
			        model := tui.InitialModel()
			        program := tui.CreateProgram(model)

			        // Run the TUI and get the result
			        result, err := program.Run()
			        if err != nil {
			                fmt.Println(errorStyle.Render("✗ Error running program:"), err)
			                return
			        }

			        loginModel, ok := result.(tui.LoginModel)
			        if !ok {
			                fmt.Println(errorStyle.Render("✗ Invalid login model"))
			                return
			        }

			        // Start spinner
			        s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
			        s.Prefix = "  "
			        s.Suffix = " Connecting to ServiceNow instance..."
			        s.Start()

											// Process instance name and construct URL
											instanceName := loginModel.InstanceURL
											// Remove any domain suffix if present
											if strings.Contains(instanceName, ".") {
											    instanceName = strings.Split(instanceName, ".")[0]
											}
											// Construct the full ServiceNow URL
											instanceURL := fmt.Sprintf("https://%s.service-now.com", instanceName)

			        // Simulate API call with delay

																									// Create snow client
																									client, err := snow.NewClient(instanceName, loginModel.Username, loginModel.Password)
																									if err != nil {
																									    s.Stop()
																									    fmt.Println(errorStyle.Render("\n✗ Failed to create client:"), err)
																									    return
																									}

											              // Test connection and authenticate
											              user, err := client.Authenticate()
											              s.Stop()
											              if err != nil {
											                      fmt.Println(errorStyle.Render("\n✗ Authentication failed:"), err)
											                      return
											              }

											              // Save credentials after successful authentication
											              if err := client.SaveConfig(); err != nil {
											                      fmt.Println(errorStyle.Render("\n✗ Failed to save credentials:"), err)
											                      return
											              }

											              // Show success message
											              fmt.Println(successStyle.Render("\n✓ Successfully connected to ServiceNow!"))
											              fmt.Printf(infoStyle.Render("\nInstance: %s\nUser: %s\n"), instanceURL, loginModel.Username)
											              if user != nil {
											                      fmt.Printf(infoStyle.Render("Name: %s\nEmail: %s\n"), user.Name, user.Email)
											              }
			},
	}

