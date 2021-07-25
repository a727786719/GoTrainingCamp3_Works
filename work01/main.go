package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

// DAOError DAO 层统一错误, 实现了 Unwrap 接口
type DAOError struct {
	err error
}

func newDAOError(err error) *DAOError {
	return &DAOError{err}
}

func (err DAOError) Error() string {
	return err.err.Error()
}

func (err DAOError) Unwrap() error {
	return errors.Unwrap(err.err)
}

type Student struct {
	ID   int
	Name string
}

func (t Student) String() string {
	return fmt.Sprintf("Student (id=%d, name=%s)", t.ID, t.Name)
}

func main() {
	conn, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/test")
	if err != nil {
		fmt.Println("open db:", err)
		return
	}

	defer conn.Close()

	err = doPrepare(conn)
	if err != nil {
		fmt.Println("do prepare:", err)
		return
	}

	students, err := doQuery(conn)
	if err != nil {
		fmt.Println("do query students:", err)
		return
	}

	fmt.Println(students)

	// 查询ID为3的学生信息, 由于
	stu, err := doQueryOne(conn)
	if err != nil {
		// 判断是否为没有查询到数据并进行处理
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("not found student with id 3")
			// TODO 业务处理
			return
		}
		// TODO 处理其它错误情况
		fmt.Println("do query student:", err)
		return
	}
	fmt.Println(stu)
}

func doPrepare(conn *sql.DB) error {
	_, err := conn.Exec("create table if not exists student(id int not null, name varchar(10) not null)")
	if err != nil {
		return newDAOError(err)
	}

	stmt, err := conn.Prepare("insert into student(id, name) value(?, ?), (?, ?)")
	if err != nil {
		return newDAOError(err)
	}

	_, err = stmt.Exec(1, "张三", 2, "李四")
	if err != nil {
		return newDAOError(err)
	}

	return nil
}

func doQuery(conn *sql.DB) ([]*Student, error) {
	rows, err := conn.Query("select id, name from student where id in (3,4)")
	if err != nil {
		return nil, newDAOError(err)
	}
	defer rows.Close()

	students := make([]*Student, 0, 2)
	for rows.Next() {
		var id = 0
		var name = ""
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, newDAOError(err)
		}

		students = append(students, &Student{ID: id, Name: name})
	}

	return students, nil
}

func doQueryOne(conn *sql.DB) (*Student, error) {
	row := conn.QueryRow("select id, name from student where id in(3, 4)")
	if row.Err() != nil {
		return nil, newDAOError(row.Err())
	}

	var id = 0
	var name = ""
	err := row.Scan(&id, &name)
	if err != nil {
		return nil, newDAOError(err)
	}

	return &Student{ID: id, Name: name}, nil
}
