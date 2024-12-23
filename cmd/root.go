package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/devopsext/utils"
	"github.com/dshaparenko/advisor/common"
	"github.com/dshaparenko/advisor/processor"
	"github.com/spf13/cobra"
)

var APPNAME = "ADVISOR"

var runProcessorOptions = processor.RunProcessorOptions{
	Timeout:     envGet("RUN_TIMEOUT", 60).(int),
	Concurrency: envGet("RUN_CONCURRENCY", 2).(int),
	ThanosURL:   envGet("THANOS_URL", "").(string),
	Quantile:    envGet("QUANTILE", "0.90").(string),
	Mode:        envGet("MODE", "sum_irate").(string),
	LimitMargin: envGet("LIMIT_MARGIN", "1.2").(string),
	Owners:      strings.Split(envGet("OWNERS", []string{"bla", "bla2"}).(string), ","),
	PodName:     envGet("POD_NAME", "").(string),
}

func envGet(s string, def interface{}) interface{} {
	return utils.EnvGet(fmt.Sprintf("%s_%s", APPNAME, s), def)
}

func interceptSyscall() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		fmt.Println("Exiting...")
		os.Exit(1)
	}()
}

func Execute() {

	var rootCmd = &cobra.Command{
		Use:   "advisor",
		Short: "Advisor",
		Long:  `Advisor is a tool for providing advice on k8s requests and limits`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to Advisor")

			processors := common.NewProcessors()
			runProcessor, err := processor.NewRunProcessor(runProcessorOptions)
			if err != nil {
				fmt.Println("Error creating RunProcessor:", err)
				os.Exit(1)
			}
			processors.Add(runProcessor)

			p := processors.Find("RunProcessor")
			if rp, ok := p.(*processor.RunProcessor); ok {
				rp.Run(runProcessorOptions.PodName)

			} else {
				fmt.Println("Processor not found")
			}
		},
	}

	flags := rootCmd.PersistentFlags()
	flags.IntVarP(&runProcessorOptions.Timeout, "timeout", "t", runProcessorOptions.Timeout, "Timeout")
	flags.IntVarP(&runProcessorOptions.Concurrency, "concurrency", "c", runProcessorOptions.Concurrency, "Concurrency")
	flags.StringVarP(&runProcessorOptions.ThanosURL, "thanos-url", "u", runProcessorOptions.ThanosURL, "Thanos URL")
	flags.StringSliceVar(&runProcessorOptions.Owners, "owners", runProcessorOptions.Owners, "Owners")
	flags.StringVarP(&runProcessorOptions.PodName, "pod-name", "p", runProcessorOptions.PodName, "Pod Name")
	flags.StringVarP(&runProcessorOptions.Quantile, "quantile", "q", runProcessorOptions.Quantile, "Quantile")
	flags.StringVarP(&runProcessorOptions.Mode, "mode", "m", runProcessorOptions.Mode, "Mode")
	flags.StringVarP(&runProcessorOptions.LimitMargin, "limit-margin", "l", runProcessorOptions.LimitMargin, "Limit Margin")

	interceptSyscall()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
