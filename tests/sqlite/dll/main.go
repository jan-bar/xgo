package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("打开数据")
	db, err := sql.Open("sqlite3", "./foo.db")
	checkErr(err)

	fmt.Println("生成数据表")
	sql_table := `
CREATE TABLE IF NOT EXISTS "userinfo" (
   "uid" INTEGER PRIMARY KEY AUTOINCREMENT,
   "username" VARCHAR(64) NULL,
   "departname" VARCHAR(64) NULL,
   "created" TIMESTAMP default (datetime('now', 'localtime'))  
);
CREATE TABLE IF NOT EXISTS "userdeatail" (
   "uid" INT(10) NULL,
   "intro" TEXT NULL,
   "profile" TEXT NULL,
   PRIMARY KEY (uid)
);
   `
	_, err = db.Exec(sql_table)
	checkErr(err)

	//插入数据
	fmt.Print("插入数据, ID=")
	stmt, err := db.Prepare("INSERT INTO userinfo(username, departname)  values(?, ?)")
	checkErr(err)
	res, err := stmt.Exec("astaxie", "研发部门")
	checkErr(err)
	id, err := res.LastInsertId()
	checkErr(err)
	fmt.Println(id)

	//更新数据
	fmt.Print("更新数据 ")
	stmt, err = db.Prepare("update userinfo set username=? where uid=?")
	checkErr(err)
	res, err = stmt.Exec("astaxieupdate", id)
	checkErr(err)
	affect, err := res.RowsAffected()
	checkErr(err)
	fmt.Println(affect)

	//查询数据
	fmt.Println("查询数据")
	rows, err := db.Query("SELECT * FROM userinfo")
	checkErr(err)
	defer rows.Close()
	for rows.Next() {
		var uid int
		var username string
		var department string
		var created string
		err = rows.Scan(&uid, &username, &department, &created)
		checkErr(err)
		fmt.Println(uid, username, department, created)
	}

	//删除数据
	fmt.Println("删除数据")
	stmt, err = db.Prepare("delete from userinfo where uid=?")
	checkErr(err)
	res, err = stmt.Exec(id)
	checkErr(err)
	affect, err = res.RowsAffected()
	checkErr(err)
	fmt.Println(affect)

	var version string
	err = db.QueryRow(`select SQLITE_VERSION()`).Scan(&version)
	checkErr(err)
	fmt.Println(version)

	err = db.QueryRow(`select SQLITE_SOURCE_ID()`).Scan(&version)
	checkErr(err)
	fmt.Println(version)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
