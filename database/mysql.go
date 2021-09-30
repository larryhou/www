package database

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "reflect"
    "strings"
)

var db *sql.DB

func init() {
    instance, err := sql.Open("mysql", `larryhou:~!@#WA1qaz@/rapidci?parseTime=True&loc=Local`)
    if err != nil { panic(err) }
    db = instance
}

func Close() { db.Close() }

func Execute(sql string, args ...interface{}) error  {
    stmt, err := db.Prepare(sql)
    if err != nil { return err }
    defer stmt.Close()

    if _, err = stmt.Exec(args...); err != nil { return err }
    return nil
}

func ExecuteMany(sql string, args ...[]interface{}) error {
    stmt, err := db.Prepare(sql)
    if err != nil { return err }
    defer stmt.Close()

    for _, args := range args {
        if _, err := stmt.Exec(args...); err != nil { return err}
    }
    return nil
}

func parse(ctype *sql.ColumnType, value interface{}) interface{} {
    //log.Printf("%v %v %v\n", ctype.Name(), ctype.DatabaseTypeName(), ctype.ScanType().Name())
    dtype := ctype.DatabaseTypeName()
    if s, ok := value.([]byte); ok {
        if dtype == "TEXT" || strings.HasSuffix(dtype, "CHAR") { return string(s) }
    }

    return value
}

func LQuery(query string, args ...interface{}) ([][]interface{}, error) {
    stmt, err := db.Prepare(query)
    if err != nil { return nil, err }
    defer stmt.Close()

    rows, err := stmt.Query(args...)
    if err != nil { return nil, err }
    defer rows.Close()

    columns, _ := rows.ColumnTypes()
    values := make([]interface{}, len(columns))
    for i, _ := range values { values[i] = new(interface{}) }

    var result [][]interface{}
    for rows.Next() {
        if err := rows.Scan(values...); err != nil { return nil, err }
        record := make([]interface{}, len(columns))
        for i, value := range values {
            record[i] = parse(columns[i], *value.(*interface{}))
        }
        result = append(result, record)
    }

    return result, nil
}

func MQuery(query string, args ...interface{}) ([]map[string]interface{}, error) {
    stmt, err := db.Prepare(query)
    if err != nil { return nil, err }
    defer stmt.Close()

    rows, err := stmt.Query(args...)
    if err != nil { return nil, err }
    defer rows.Close()

    columns, _ := rows.ColumnTypes()
    values := make([]interface{}, len(columns))
    for i, _ := range values { values[i] = new(interface{}) }

    var result []map[string]interface{}
    for rows.Next() {
        if err := rows.Scan(values...); err != nil { return nil, err }
        record := make(map[string]interface{})
        for i, value := range values {
            c := columns[i]
            record[c.Name()] = parse(c, *value.(*interface{}))
        }
        result = append(result, record)
    }

    return result, nil
}

func Query(query string, model interface{}, args ...interface{}) ([]interface{}, error) {
    rv := reflect.ValueOf(model).Elem() // derreferenced from ptr
    rt := rv.Type()

    stmt, err := db.Prepare(query)
    if err != nil { return nil, err }
    defer stmt.Close()

    rows, err := stmt.Query(args...)
    if err != nil { return nil, err }
    defer rows.Close()

    columns, _ := rows.Columns()
    mapping := make(map[string]interface{})
    for i := 0; i < rt.NumField(); i++ {
        name := strings.ToLower(rt.Field(i).Name)
        vf := rv.Field(i)
        if vf.CanSet() { mapping[name] = vf.Addr().Interface() }
    }

    values := make([]interface{}, len(columns))
    for i, name := range columns {
        name = strings.ToLower(name)
        if v, ok := mapping[name]; ok { values[i] = v } else {
            values[i] = new(interface{})
        }
    }

    var result []interface{}
    for rows.Next() {
        if err := rows.Scan(values...); err != nil { return nil, err }
        rc := reflect.New(rt).Elem() // create a new struct for each record
        for i := 0; i < rv.NumField(); i++ {
            vf := rc.Field(i)
            if vf.CanSet() { vf.Set(rv.Field(i)) }
        } // copy
        result = append(result, rc.Addr().Interface())
    }

    if len(result) == 0 {return nil, fmt.Errorf("no database record found: %v", args)}
    return result, nil
}
