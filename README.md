# FreeSozai

A simple to use text pastebin.

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/edvakf/freesozai)

## usage

POST any text output from command line, then it returns you a URL.

```
$ echo "foo\nbar\nbaz" | curl --data-binary @- http://example.herokuapp.com/
http://freesozai.herokuapp.com/1234abc
```

## environment variables

### `ENDPOINT`

* If set, the post URL will be "http://example.herokuapp.com/myendpoint"
* You might enhance security slightly

### `WEBHOOK_URL`

* To which notification is posted when a new post is created
* You can think of it as an equivalent of `curl --data=$URL_OF_NEWLY_CREATED_POST $WEBHOOK_URL`
