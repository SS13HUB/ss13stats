package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: change these to be cmd args instead!
const DB_FIL = "/home/lmas/projects/ss13_se/src/db.sqlite3"

// Dir to save new graphs in
const SAVE_DIR = "/home/lmas/projects/ss13_se/src/static/graphs"

// How far back in time the graphs will go
var LAST_WEEK = time.Now().AddDate(0, 0, -7)
var LAST_MONTH = time.Now().AddDate(0, -1, 0)

var WEEK_DAYS = [7]string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

func main() {
	// open a db connection
	db, err := sql.Open("sqlite3", DB_FIL)
	checkerror(err)
	defer db.Close()

	// loop over each server in db
	rows, err := db.Query("select id, title from gameservers_server")
	checkerror(err)
	defer rows.Close()

	var (
		id    int
		title string
	)
	for rows.Next() {
		err := rows.Scan(&id, &title)
		checkerror(err)
		weeklyhistorygraph(db, id, title)
		monthlyhistorygraph(db, id, title)
		monthlyaveragedaygraph(db, id, title)
	}
	err = rows.Err()
	checkerror(err)
}

func checkerror(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func setuptemppaths(prefix string, title string) (f *os.File, name string, path string) {
	// create a tmp file
	file, err := ioutil.TempFile("", prefix)
	checkerror(err)

	// Make sure we have somewhere to save the stored graphs in
	err = os.MkdirAll(SAVE_DIR, 0777)
	checkerror(err)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(title)))
	path = filepath.Join(SAVE_DIR, fmt.Sprintf("%s%s", prefix, hash))

	return file, file.Name(), path
}

func weeklyhistorygraph(db *sql.DB, id int, title string) {
	prefix := "week-time-"
	ifile, ifilename, ofilename := setuptemppaths(prefix, title)
	defer ifile.Close()

	// Rows for server, newer then LAST_WEEK and only on the hour
	rows, err := db.Query("select created,players from gameservers_serverhistory where server_id = ? and created >= ? and strftime('%M', created) = '00' order by created asc", id, LAST_WEEK)
	checkerror(err)
	defer rows.Close()

	var (
		created time.Time
		players int
	)
	for rows.Next() {
		err := rows.Scan(&created, &players)
		checkerror(err)
		_, err = ifile.WriteString(fmt.Sprintf("%d, %d\n", created.Unix(), players))
		checkerror(err)
	}
	err = rows.Err()
	checkerror(err)

	// run the plotter against the data file
	err = exec.Command("./plot_time.sh", ifilename, ofilename).Run()
	checkerror(err)

	// close and remove the tmp file
	ifile.Close()
	os.Remove(ifilename)
}

func monthlyhistorygraph(db *sql.DB, id int, title string) {
	prefix := "month-time-"
	ifile, ifilename, ofilename := setuptemppaths(prefix, title)
	defer ifile.Close()

	// Rows for server, newer then LAST_WEEK and only every 6th hour
	rows, err := db.Query("select created,players from gameservers_serverhistory where server_id = ? and created >= ? and strftime('%M', created) = '00' and strftime('%H', created) in ('00', '06', '12', '18') order by created asc", id, LAST_MONTH)
	checkerror(err)
	defer rows.Close()

	var (
		created time.Time
		players int
	)
	for rows.Next() {
		err := rows.Scan(&created, &players)
		checkerror(err)
		_, err = ifile.WriteString(fmt.Sprintf("%d, %d\n", created.Unix(), players))
		checkerror(err)
	}
	err = rows.Err()
	checkerror(err)

	// run the plotter against the data file
	err = exec.Command("./plot_time.sh", ifilename, ofilename).Run()
	checkerror(err)

	// close and remove the tmp file
	ifile.Close()
	os.Remove(ifilename)
}

func monthlyaveragedaygraph(db *sql.DB, id int, title string) {
	prefix := "month-avg_day-"
	ifile, ifilename, ofilename := setuptemppaths(prefix, title)
	defer ifile.Close()

	// get the server's data and write it to the file
	// TODO: Move sunday (first day in list at 0) to the end...
	rows, err := db.Query("select strftime('%w', created) as weekday, avg(players) from gameservers_serverhistory where server_id = ? and created >= ? group by weekday;", id, LAST_MONTH)
	checkerror(err)
	defer rows.Close()

	var (
		day     int
		players float64
	)
	for rows.Next() {
		err := rows.Scan(&day, &players)
		checkerror(err)
		_, err = ifile.WriteString(fmt.Sprintf("%s, %f\n", WEEK_DAYS[day], players))
		checkerror(err)
	}
	err = rows.Err()
	checkerror(err)

	// run the plotter against the data file
	err = exec.Command("./plot_bar.sh", ifilename, ofilename).Run()
	checkerror(err)

	// close and remove the tmp file
	ifile.Close()
	os.Remove(ifilename)
}