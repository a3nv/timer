package cmd

import (
	"encoding/csv"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

type Post struct {
	Id      int
	Content string
	Author  string
}

var (
	testCmd = &cobra.Command{
		Use:   "test",
		Short: "test short",
		RunE: func(cmd *cobra.Command, args []string) error {
			csvFile, err := os.Create("/Users/a3nv/gitrepo/tools/timer/posts.csv")
			if err != nil {
				panic(err)
			}
			defer csvFile.Close()

			allPosts := []Post{
				Post{Id: 1, Content: "test", Author: "Bonny"},
			}

			writer := csv.NewWriter(csvFile)
			for _, post := range allPosts {
				line := []string{strconv.Itoa(post.Id), post.Content, post.Author}
				err := writer.Write(line)
				if err != nil {
					panic(err)
				}
			}
			writer.Flush()
			return nil
		},
	}
)
