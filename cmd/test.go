package cmd

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"time"
)

type Post struct {
	Id      int
	Content string
	Author  string
}

func init() {
	//Root.AddCommand(testCmd)
}

var (
	testCmd = &cobra.Command{
		Use:   "test",
		Short: "test short",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
)

func save(m model) error {
	csvFile, err := os.Create("/Users/a3nv/gitrepo/tools/timer/posts.csv")
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	allPosts := []Post{
		Post{Id: 1, Content: m.start.Format(time.Kitchen), Author: m.start.Format(time.DateOnly)},
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
}

func save2(m model) {
	var exist bool
	if _, err := os.Stat("test.db"); err == nil {
		fmt.Printf("File exists\n")
		exist = true
	} else {
		_, err := os.Create("test.db")
		checkError(err)
		exist = false
	}

	db, err := sql.Open("sqlite3", "test.db")

	checkError(err)
	if !exist {
		createTable(db)
	}

	stmt, err := db.Prepare("INSERT INTO timer (startTime, startDate, duration) VALUES (?, ?, ?)")
	checkError(err)

	id, err := stmt.Exec(m.start.Format(time.Kitchen), m.start.Format(time.DateOnly), m.duration.Milliseconds()/100)
	checkError(err)

	fmt.Println(id)

	rows, err := db.Query("SELECT * FROM timer")
	checkError(err)

	var uuid int
	var startTime string
	var startDate string
	var duration string
	for rows.Next() {
		err = rows.Scan(&uuid, &startTime, &startDate, &duration)
		checkError(err)
		//fmt.Println(uuid)
		fmt.Println(strconv.Itoa(uuid) + " - " + startTime + " - " + startDate + " - " + duration)
	}
	err = rows.Close()
	checkError(err)
}

func createTable(db *sql.DB) {
	timers := `CREATE TABLE timer (
    	id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    	startTime TEXT,
    	startDate TEXT,
    	duration TEXT
    );`
	query, err := db.Prepare(timers)
	checkError(err)
	query.Exec()
	fmt.Println("Table create successfully")
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
