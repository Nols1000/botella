Ava
===

_Ava is a temporal name since it already clashes with an important JS test runner. New names are welcomed!_

Slack bot with plugins. What's the selling point then? The plugins are Docker containers and the bot interacts with them via stdin/stdout. In plain English that means that you can develop your plugins in whatever language you want and you will just need to pack them as Docker container to be able to use them with Ava.

In the [examples/](examples/) folder you can find a simple plugin that does two important things:

- it will echo whatever message is send to it, thinking out of the box you could do whatever you want with that input (from the container stdin) and output whatever you want instead just an echo.
- it uses environment variables that are set in the configuration file, this will allow you to secretly pass secrets (my apologies for being redundant), for example some AWS access keys or similar.

Usage
-----

You just need to be sure that you have a `ava.yaml` file in the path where you run it. This project could look like:

```yaml
adapters:
  slack:
    key: xoxb-xxx
    channels:
      - general

plugins:
  - image: agonzalezro/ava-test
    environment:
      KEY: this-is-a-secret
    only_mentions: true
```

I think that it's pretty explainable, but basically: 

- `adapters` explain the configuration of the bot adapter, at the moment of writing you can just use Slack, in the future some other adapters as IRC should be supported.
- `plugins` is a list with the plugins that you want to run. In their `environment` section you can set variable that are going to be passed to the Docker containers.

In addition, for each plugin you will be able to set the following configs:

- `only_mentions`: will make the bot just react when its mentioned.
- `only_channels`: will make the bot just react when the message was send in a channel.
- `only_direct_messages`: you know how this goes...

Developing plugins
------------------

Developing a plugin for Ava is extremely simple, you will need to write a program and pack it with Docker.

What are the characteristics that my program needs to follow? It will need to read a line (or several) from stdin and write a response to stdout.

In the [examples/](examples/) folder you will find an example plugin called `ava-test`. This plugin is uploaded to my Docker Hub so you can use it without building it, but it's a good place to learn.

Developing Ava
--------------

# Compiling

You will need `Go` and `glide`:

```bash
$ glide install
$ go build
```

# Testing

```bash
$ go test
```

Tips
----

You will need an Slack Key to use the Slack adapter (the only one available for now).

Go to https://your-org-here.slack.com/services/new/bot and create a new bot, after configuring it you will see an API Token, you will need that one.