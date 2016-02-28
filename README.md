# What is github webhook?

This is a small http server written in **golang** to listen for [github
webhook push
events](https://developer.github.com/v3/activity/events/types/#pushevent).
In order to run deployment or other tasks based on new commits pushed to
the repository.

It is a ready drop in replacement, just copy the code and adapt to your
needs if there are not enough options provided by this default
implementation.

## Install

Given we will run it on our server: **http://my-server.com** and our
repository owner is **golang-lt** and name is **gophers.lt**. Then the
configuration file should look like:

``` json
{
  "webhooks": [
    {
      "repository": "golang-lt/gophers.lt",
      "secret": "secret configured in github webhook",
      "command": {
        "workdir": "/home/gopher/gophers.lt",
        "exec": "./deploy.sh"
      }
    }
  ]
}
```

Register a webhook in your github repository settings/webhooks. Point to
**http://my-server.com:9000** or proxy it through webserver.

    go build github-webhook
    ./github-webhook your-config.json

Try the webhook from your github account. By default only **push** events
are read and only **master** branch changes are taken to account.

The command receives some arguments, for example:

``` bash
#!/bin/sh

USERNAME=$1
USERMAIL=$2
COMMIT_ID=$3
COMMIT_MSG=$4
COMMIT_DATE=$5

git fetch && git checkout $COMMIT_ID
sudo supervisorctl restart project
```

