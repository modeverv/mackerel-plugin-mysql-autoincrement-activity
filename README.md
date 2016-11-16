# mackerel-plugin-mysql-autoincrement-activity

![screenshot](https://cloud.githubusercontent.com/assets/48426/20335552/d59ff07c-ac06-11e6-8d2b-b3b213e1736b.png)

## Description

It is a plugin which monitors the upper limit of auto increment primary key(MySQL only).

## Synopsis

```shell
mackerel-plugin-mysql-autoincrement-activity --database=<DATABASE_NAME> [ --host=<DATABSE_HOST> ] [ --username=<DATABASE_USER> ] [ --password=<DATABASE_PASSWORD> ] [ --<DATABASE_PORT> ] [ --socket=<DATABASE_SOCKET> ] [ --tempfile=<TMPFILE> ]
```

## mackerel-agent.conf

```
[plugin.metrics.mysql-autoincrement-activity]
command = "/path/to/mackerel-plugin-mysql-autoincrement-activity --database=DATABASE_NAME"
```
