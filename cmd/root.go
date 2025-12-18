/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"order-book-manager/client"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "order-book-manager",
	Short: "Order Book Manager",
	Long:  `Order Book Manager is connected with Binance to get the latest currency updates and broadcast it to the subscribed users. Default Currency Pair BTCUSDT`,

	Run: func(cmd *cobra.Command, args []string) {
		client.SetCurrencyPair("BTCUSDT") // default currency pair
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
