package main

import (
	"flag"
	"fmt"
	"regexp"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
)

type AutoIncrementColumn struct {
	Name     string
	Type     string
	Byte     uint
	Limit    uint
	Unsigned bool
	Count    uint
}

func (c *AutoIncrementColumn) parseColumnType(types string) {
	c.Type = c.columnTypeToType(types)
	c.Byte = c.typeToByte(c.Type)
	c.Unsigned = c.columnTypeToUnsigned(types)
	c.Limit = c.calculateLimit(c.Byte)
}

func (c *AutoIncrementColumn) columnTypeToType(types string) string {
	re, _ := regexp.Compile("^[^(]+")
	return re.FindString(types)
}

func (c *AutoIncrementColumn) typeToByte(types string) uint {
	switch types {
	case "bigint":
		return uint(8)
	case "int":
		return uint(4)
	case "mediumint":
		return uint(3)
	case "smallint":
		return uint(2)
	case "tinyint":
		return uint(1)
	default:
		return uint(0)
	}
}

func (c *AutoIncrementColumn) calculateLimit(bytes uint) uint {
	if bytes == 0 {
		return uint(0)
	} else if bytes > 8 {
		panic("received more than bigger bigint data size")
	}

	max := uint(1 << (bytes*8 - 1))
	if c.Unsigned {
		max = max * 2
	}

	return max - 1
}

func (c *AutoIncrementColumn) columnTypeToUnsigned(types string) bool {
	match, _ := regexp.MatchString("unsigned", types)
	return match
}

func NewAutoIncrementColumn(name string, columnType string, count uint) *AutoIncrementColumn {
	c := &AutoIncrementColumn{Name: name, Count: count}
	c.parseColumnType(columnType)
	return c
}

type MySQLAutoIncrementActivityPlugin struct {
	Prefix   string
	GraphKey string
	Dbopts   map[string]string
}

func (plugin MySQLAutoIncrementActivityPlugin) GraphDefinition() map[string](mp.Graphs) {
	var columns []*AutoIncrementColumn = plugin.autoIncrementColumns()
	var metricsDefs []mp.Metrics
	for _, column := range columns {
		metricsDefs = append(metricsDefs, mp.Metrics{Name: column.Name, Label: column.Name, Diff: false})
	}

	return map[string](mp.Graphs){
		fmt.Sprintf("%s.#", plugin.GraphKey): mp.Graphs{
			Label:   plugin.Prefix + " autoincrement activity",
			Unit:    "percentage",
			Metrics: metricsDefs,
		},
	}
}

func (plugin MySQLAutoIncrementActivityPlugin) FetchMetrics() (map[string]interface{}, error) {
	data := make(map[string]interface{})

	cs := plugin.autoIncrementColumns()
	if len(cs) > 0 {
		for _, c := range cs {
			var metricsKey string = fmt.Sprintf("%s.%s.%s", plugin.GraphKey, plugin.Dbopts["database"], c.Name)
			data[metricsKey] = (float64(c.Count) / float64(c.Limit)) * 100.0
		}
	}
	return data, nil
}

var resultCache [](*AutoIncrementColumn)

func (plugin MySQLAutoIncrementActivityPlugin) autoIncrementColumns() [](*AutoIncrementColumn) {
	if len(resultCache) > 0 {
		return resultCache
	}
	rows := plugin.getAutoIncrementColumnData()
	if len(rows) > 0 {
		resultCache = rows
	}

	return resultCache
}

func (plugin MySQLAutoIncrementActivityPlugin) getAutoIncrementColumnData() []*AutoIncrementColumn {

	conn := plugin.databaseClient(plugin.Dbopts)

	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	sql := "" +
		"SELECT" +
		"	DISTINCT" +
		"	`TABLES`.`TABLE_NAME`," +
		"	`COLUMNS`.`COLUMN_TYPE`," +
		"	`TABLES`.`AUTO_INCREMENT`" +
		"FROM" +
		"	`information_schema`.`TABLES`" +
		"	INNER JOIN" +
		"		`information_schema`.`COLUMNS`" +
		"	ON" +
		"		`TABLES`.`TABLE_NAME` = `COLUMNS`.`TABLE_NAME` AND" +
		"		`COLUMNS`.`TABLE_SCHEMA` = '%s'" +
		"WHERE" +
		"	`TABLES`.`TABLE_SCHEMA` = '%s' AND" +
		"	`COLUMNS`.`EXTRA` = 'auto_increment'"

	rows, result, err := conn.Query(sql, plugin.Dbopts["database"], plugin.Dbopts["database"])
	if err != nil {
		panic(err)
	}

	tableName := result.Map("TABLE_NAME")
	columnType := result.Map("COLUMN_TYPE")
	autoIncrement := result.Map("AUTO_INCREMENT")

	var res []*AutoIncrementColumn

	for _, row := range rows {
		res = append(res, NewAutoIncrementColumn(row.Str(tableName), row.Str(columnType), row.Uint(autoIncrement)))
	}
	return res

}

var driver mysql.Conn

func (plugin MySQLAutoIncrementActivityPlugin) databaseClient(options map[string]string) mysql.Conn {
	if driver != nil {
		return driver
	}
	var dest string
	if options["socket"] != "" {
		dest = options["socket"]
	} else {
		dest = fmt.Sprintf("%s:%s", options["host"], options["port"])
	}
	driver = mysql.New("tcp", "", dest, options["user"], options["pass"], options["database"])
	return driver
}

func main() {
	dbHost := flag.String("host", "localhost", "Hostname")
	dbPort := flag.String("port", "3306", "Port")
	dbSocket := flag.String("socket", "", "Port")
	dbUser := flag.String("username", "root", "Username")
	dbPass := flag.String("password", "", "Password")
	dbName := flag.String("database", "", "Database")
	prefix := flag.String("prefix", "mysql", "Prefix")
	tempFile := flag.String("tempfile", "", "Tempfile")
	flag.Parse()

	options := map[string]string{
		"host":     *dbHost,
		"port":     *dbPort,
		"socket":   *dbSocket,
		"user":     *dbUser,
		"pass":     *dbPass,
		"database": *dbName,
	}

	plugin := MySQLAutoIncrementActivityPlugin{
		Dbopts:   options,
		Prefix:   *prefix,
		GraphKey: "mysql.autoincrement.activity",
	}
	helper := mp.NewMackerelPlugin(plugin)
	helper.Tempfile = *tempFile
	helper.Run()
}
