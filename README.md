`datadog_csv`
=============

A simple command line tool to download a CSV file of a specific metrics query from DataDog.


DOWNLOAD
--------
Get the latest binary release for your OS from [https://github.com/cloudops/datadog_csv/releases/latest](https://github.com/cloudops/datadog_csv/releases/latest).

The binary can be run from anywhere on your filesystem.  I would recommend renaming the downloaded binary to `datadog_csv` in order to follow along with the examples below.


USAGE
-----
`datadog_csv` leverages the DataDog API.  In order to use the API, you will need both an `api_key` and an `app_key`, which you can get from [https://app.datadoghq.com/account/settings#api](https://app.datadoghq.com/account/settings#api).


There are a few ways to specify configuration details when running the `datadog_csv` binary.

1. Pass the params on the command line.  
`$ ./datadog_csv -api_key="xyz" -app_key="qrs"`

2. Populate a `config.ini` file and pass the filepath on the command line.  
`$ ./datadog_csv -config="config.ini"`

With the `config.ini` file in the format of:
```
api_key = xyz
app_key = qrs
```

*Note: `-api_key` and `-app_key` are private, so you may not want to enter them on the command line or store them in a config file.  In this case, if you do not specify these options via either of the above methods, the application will prompt you for these details once launched and your input will be masked.*
```
$ ./datadog_csv -config="config.ini"
Enter your API_KEY:
Enter your APP_KEY:
```

**RECOMMENDED USAGE**  
**`$ ./datadog_csv -config="system_cpu_idle.ini" -csv_file="system_cpu_idle.csv"`**

**`system_cpu_idle.ini`**
```
api_key = xyz
app_key = qrs
query = avg:system.cpu.idle{logs}
start = 2018/10/01-00:00
end = 2018/11/01-00:00
```


CLI PARAMS
----------
```
$ ./datadog_csv -h
Usage of ./datadog_csv:
  -api_key string
        API Key to connect to DataDog
  -app_key string
        APP Key for this app in DataDog
  -config string
        Path to ini config for using in go flags. May be relative to the current executable path.
  -configUpdateInterval duration
        Update interval for re-reading config file set via -config flag. Zero disables config file re-reading.
  -csv_file string
        The filepath of the CSV file to output to
  -dumpflags
        Dumps values for all flags defined in the app into stdout in ini-compatible syntax and terminates the app.
  -end string
        The ending point for the date range to query (format: yyyy/mm/dd-hh:mm) (required)
  -interval string
        The preferred data interval. [5m, 10m, 20m, 30m, 1h, 2h, 4h, 8h, 12h, 24h] (default "1h")
  -query string
        The DataDog query to run (required)
  -start string
        The starting point for the date range to query (format: yyyy/mm/dd-hh:mm) (required)
  -v    Version of the binary (optional)
```


DETAILS
-------
A non-obvious detail about the DataDog API is that you can not specify the granularity of the data you receive from the API.  Essentially, the API will give you a maximum of around 300 results per query.  The DataDog backend will automatically down sample query results for longer query range in order to provide a relatively consistent number of result records.  You can learn more about how this works from the [DataDog Docs](https://docs.datadoghq.com/graphing/faq/what-is-the-granularity-of-my-graphs-am-i-seeing-raw-data-or-aggregates-on-my-graph/).

Unfortunately, this down sampling of the data over longer periods of time does not fit my use case.  I need to be able to query data over long periods of time, but with the results all conforming to a specific granularity/interval.  This is where the `-interval` flag comes in.  Basically, it lets you specify on what interval you want the aggregation of data for your results.  In the background `datadog_csv` will potentially move your `-start` date earlier to ensure that your date range is a perfect multiple of a mapped duration which is keyed from the `-interval` value.  If that sentence just made your head explode, just know that the date range you specified may not be exactly what is returned.  In order to be able to set the interval on which I want my data points of the result, I have to do a bunch of math to assign an appropriate overall date range which can then be broken up into smaller equal sized ranges which I can predict to have the appropriate interval.

To hopefully clarify these details a bit more, I have attached the log output of a single query which is produced by calling 3 separate sub-queries.

**`datadog_csv.log`**
```
2018/11/08 23:11:00 Connecting to DataDog...
2018/11/08 23:11:00 Requested date range: 2018/10/01-00:00 to 2018/11/01-00:00
2018/11/08 23:11:00 Querying date range: 2018/09/26-00:00 to 2018/11/01-00:00
2018/11/08 23:11:00 Querying 'avg:system.cpu.idle{logs}' from '2018/09/26-00:00' to '2018/10/08-00:00'...
2018/11/08 23:11:00 Querying 'avg:system.cpu.idle{logs}' from '2018/10/08-00:00' to '2018/10/20-00:00'...
2018/11/08 23:11:00 Querying 'avg:system.cpu.idle{logs}' from '2018/10/20-00:00' to '2018/11/01-00:00'...
```