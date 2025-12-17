/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"order-book-manager/client"

	"github.com/spf13/cobra"
)

// currpairCmd represents the currpair command
var currpairCmd = &cobra.Command{
	Use:   "currpair",
	Short: "Set the Currency Pair ex: ETHUSDT",
	Long:  `Set the Currency Pair ex: ETHUSDT`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("currpair called", args[0])
		client.SetCurrencyPair(args[0])
	},
}

func init() {
	rootCmd.AddCommand(currpairCmd)
}
