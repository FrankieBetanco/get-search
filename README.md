Get Search 
----------
Get Search is designed to search through Rapid7 [https](https://opendata.rapid7.com/sonar.https/) and 
[http](https://opendata.rapid7.com/sonar.http/) GET response data sets. 

Installation
------------
If you already have go installed, simply run 
```
go get github.com/FrankieBetanco/get-search
```

Basic Usage
-----------
Get Search is invoked on the command line by: 
```
get-search -i [input file] -m [number of goroutines to search with at once] -s [term to search for]
```

For example, 
```
get-search -i /path/to/data/http_get_443.json -m 1000 -s foo bar "foo bar"
```
would search through the file http_get_443.json with 1000 goroutines for the term 'foo', 'bar', and 'foo bar'. 
Each goroutine will handle searching through a single page.

Notes
-----
Feel free to look through my code and tell me how to make it better! I'm always trying to improve. 
