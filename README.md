# FreeSozai

A simple to use text pastebin.

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/edvakf/freesozai)

## usage

POST any text output from command line, then it returns you a URL.

```
$ echo "foo\nbar\nbaz" | curl --data-binary @- http://freesozai.herokuapp.com/
http://freesozai.herokuapp.com/1234abc
```
