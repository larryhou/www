package db

import (
    "bytes"
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "log"
    "os"
    "reflect"
    "strings"
    "sync"
    "time"
)

type sqlConfig struct {
    Host     string
    Port     string
    Username string
    Password string
    Database string
}

type metadata map[string]int
var (
    db *sql.DB
    ctx struct{
        m map[reflect.Type]metadata
        sync.RWMutex
    }
)

func init() {
    conf := sqlConfig{
        Host: os.Getenv("SQL_HOST"),
        Port: os.Getenv("SQL_PORT"),
        Username: os.Getenv("SQL_USERNAME"),
        Password: os.Getenv("SQL_PASSWORD"),
        Database: os.Getenv("SQL_DATABASE"),
    }
    if len(conf.Username) == 0 { conf.Username, conf.Password = "", "" }
    if len(conf.Database) == 0 { conf.Database = "rapidci" }
    ctx.m = make(map[reflect.Type]metadata)
    buf := &bytes.Buffer{}
    buf.WriteString(fmt.Sprintf("%s:%s@", conf.Username, conf.Password))
    if len(conf.Host) > 0 || len(conf.Port) > 0 {
        buf.WriteString("tcp(")
        if len(conf.Host) == 0 {buf.WriteString("localhost")}else{buf.WriteString(conf.Host)}
        if len(conf.Port) != 0 {buf.WriteString(fmt.Sprintf(":%s", conf.Port))}
        buf.WriteString(")")
    }
    buf.WriteString(fmt.Sprintf("/%s?parseTime=True&loc=Local", conf.Database))
    instance, err := sql.Open("mysql", buf.String())
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

    for _, entity := range args {
        if _, err := stmt.Exec(entity...); err != nil { return err}
    }
    return nil
}

func conv(ctype *sql.ColumnType, value interface{}) interface{} {
    dtype := ctype.DatabaseTypeName()
    if s, ok := value.([]byte); ok {
        if strings.HasSuffix(dtype, "TEXT") || strings.HasSuffix(dtype, "CHAR") {
            if len(s) == 0 {return ""} else {return string(s)}
        }
    }

    return value
}

func construct(t *sql.ColumnType) interface{} {
    switch t.DatabaseTypeName() {
    case "CHAR","VARCHAR","TEXT","LONGTEXT": return new(string)
    case "TIMESTAMP": return new(time.Time)
    default:
        return reflect.New(t.ScanType()).Interface()
    }
}

func LQuery(query string, args ...interface{}) [][]interface{} {
    err := error(nil)
    defer func() {
        if err != nil {log.Printf("LQuery %#v %s %v", query, args, err)}
    }()

    stmt, err := db.Prepare(query)
    if err != nil { return nil }
    defer stmt.Close()

    rows, err := stmt.Query(args...)
    if err != nil { return nil }
    defer rows.Close()

    columns, _ := rows.ColumnTypes()
    values := make([]interface{}, len(columns))
    for i := range values { values[i] = new(interface{}) }

    var result [][]interface{}
    for rows.Next() {
        if err = rows.Scan(values...); err != nil { return nil }
        record := make([]interface{}, len(columns))
        for i, value := range values {
            record[i] = conv(columns[i], *value.(*interface{}))
        }
        result = append(result, record)
    }

    return result
}

func RawQuery(query string, args ...interface{}) [][]interface{} {
    err := error(nil)
    defer func() {
        if err != nil {log.Printf("RawQuery %s %v", args, err)}
    }()

    stmt, err := db.Prepare(query)
    if err != nil { return nil }
    defer stmt.Close()

    rows, err := stmt.Query(args...)
    if err != nil { return nil }
    defer rows.Close()

    columns, _ := rows.ColumnTypes()
    var result [][]interface{}
    for rows.Next() {
        record := make([]interface{}, len(columns))
        for i := range record {record[i] = construct(columns[i])}
        if err = rows.Scan(record...); err != nil { return nil }
        result = append(result, record)
    }

    return result
}

func MapQuery(query string, args ...interface{}) []map[string]interface{} {
    err := error(nil)
    defer func() {
        if err != nil {log.Printf("MapQuery %s %v", args, err)}
    }()

    stmt, err := db.Prepare(query)
    if err != nil { return nil }
    defer stmt.Close()

    rows, err := stmt.Query(args...)
    if err != nil { return nil }
    defer rows.Close()

    columns, _ := rows.ColumnTypes()
    var result []map[string]interface{}
    for rows.Next() {
        values := make([]interface{}, len(columns))
        for i := range values {values[i] = construct(columns[i])}
        if err = rows.Scan(values...); err != nil { return nil }
        record := make(map[string]interface{})
        for i := range values {record[columns[i].Name()] = values[i]}
        result = append(result, record)
    }

    return result
}

func Query(query string, model interface{}, args ...interface{}) ([]interface{}, error) {
    rv := reflect.ValueOf(model).Elem() // derreferenced from ptr
    rt := rv.Type()
    if rt.Kind() != reflect.Struct {return nil, fmt.Errorf("expect struct type but %v", rt.Name())}

    stmt, err := db.Prepare(query)
    if err != nil { return nil, err }
    defer stmt.Close()

    rows, err := stmt.Query(args...)
    if err != nil { return nil, err }
    defer rows.Close()

    ctx.RLock()
    meta, ok := ctx.m[rt]
    ctx.RUnlock()
    if !ok {
        meta = make(metadata)
        for i := 0; i < rt.NumField(); i++ {
            rf := rt.Field(i)
            name := rf.Tag.Get("sql")
            if len(name) == 0 {name = strings.ToLower(rf.Name)}
            if rv.Field(i).CanSet() { meta[name] = i }
        }
        ctx.Lock()
        ctx.m[rt] = meta
        ctx.Unlock()
    }

    var indice []int
    columns, _ := rows.Columns()
    for _, name := range columns {
        if i, ok := meta[name]; ok {indice = append(indice, i)} else {
            indice = append(indice, -1)
        }
    }

    var records []interface{}
    for n:= 0; rows.Next(); n++ {
        var item reflect.Value
        if n == 0 {item = rv} else {item = reflect.New(rt).Elem()}
        var values []interface{}
        for _, i := range indice {
            if i == -1 {values = append(values, new(interface{}))} else {
                values = append(values, item.Field(i).Addr().Interface())
            }
        }
        if err := rows.Scan(values...); err != nil { return nil, err }
        records = append(records, item.Addr().Interface())
    }

    if len(records) == 0 {return nil, fmt.Errorf("no database records: %s %+v", query, args)}
    return records, nil
}

