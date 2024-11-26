/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/clickonetwo/automations/dialpad/internal/event"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump the received events",
	Long: `After running receive for a while, you can use this command
to view the events that have been received. A number of different options
are supported for viewing and filtering the received events.`,
	Run: func(cmd *cobra.Command, args []string) {
		env, _ := cmd.InheritedFlags().GetString("env")
		dump(env)
	},
}

func init() {
	eventsCmd.AddCommand(dumpCmd)
}

func dump(env string) {
	_ = storage.PushConfig(env)
	defer storage.PopConfig()
	actions, err := storage.FetchRangeInterval(context.Background(), event.ActionHooks, 0, -1)
	if err != nil {
		panic(err)
	}
	ignores, err := storage.FetchRangeInterval(context.Background(), event.IgnoreHooks, 0, -1)
	if err != nil {
		panic(err)
	}
	all, err := merge(actions, ignores)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(all); err != nil {
		panic(err)
	}
}

func merge(left, right []string) ([]map[string]any, error) {
	results := make([]map[string]any, len(left)+len(right))
	var l, r map[string]any
	var lStart, rStart float64
	for i, j, k := 0, 0, 0; i < len(left) || j < len(right); k++ {
		if l == nil && i < len(left) {
			if err := json.Unmarshal([]byte(left[i]), &l); err != nil {
				return nil, err
			}
			lStart = l["event_timestamp"].(float64)
		}
		if r == nil && j < len(right) {
			if err := json.Unmarshal([]byte(right[j]), &r); err != nil {
				return nil, err
			}
			rStart = r["event_timestamp"].(float64)
		}
		if l != nil && r != nil {
			if lStart <= rStart {
				results[k], l = l, nil
				i++
			} else {
				results[k], r = r, nil
				j++
			}
		} else if l != nil {
			results[k], l = l, nil
			i++
		} else {
			results[k], r = r, nil
			j++
		}
	}
	return results, nil
}
