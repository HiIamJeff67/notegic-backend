package main

import "github.com/spf13/cobra"

var writeGraphQLEnumMappingValuesToConfig = &cobra.Command{
	Use:   "writeGQLEnumMappingValues",
	Short: "Write gql enum mapping values to config file.",
	Long:  "Use truth enum values defined in golang and database to write them as enum mapping values to the graphql config file.",
	Run: func(_ *cobra.Command, _ []string) {
		// May be implemented when GraphQL enum mappings need generation.
	},
}
