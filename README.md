# IASIPBot (for Mastodon)

This is a Mastodon bot that creates It's Always Sunny in Philadelphia title
cards on demand. It makes use of my [Go-based title card
generator](https://github.com/bclindner/iasipgenerator). It can currently be
found on Mastodon at
[@iasipbot@botsin.space](https://botsin.space/@iasipbot).

# Deploying

I don't really ship binaries for this, so you'll have to compile from source.
That said, it's a Go application, so as long as you have Go installed, it's as
easy as one line in terminal:

```
go get github.com/bclindner/iasipbot_mastodon
```

While that's running, make a new Mastodon account, go into the settings, and
create a new Mastodon application. Fairly straightforward.

You'll also need a TrueType font file for Textile, the font the title cards use.
I've left it out for licensing reasons. Do some searching.

Once you're done with all that, fill in the blanks for this config file with the
path to your font file (relative paths OK) and your shiny new Mastodon app
credentials:

```json
{
	"fontPath": "<PATH_TO_YOUR_TEXTILE_TTF>",
	"credentials": {
		"server": "https://<YOUR_BOT_DOMAIN>",
		"clientID": "<YOUR_BOT_CLIENT_TOKEN",
		"clientSecret": "<YOUR_BOT_CLIENT_SECRET",
		"accessToken": "<YOUR_BOT_ACCESS_TOKEN>"
	}
}
```

From there, just launch it with `iasipbot_mastodon` and it should mumble
something about "entering event loop" and spring to life.


You can also build a container with the supplied Dockerfile. The container
expects you to mount your config to `/srv/config.json` and your font to
wherever you told the config the font is. I personally just put them both in the
same directory, set `"fontPath": "./textile.ttf"`, and mount that folder to
`/srv` directly.

# Using

Triggering the bot can be a little finicky - I made it so you have to be
somewhat deliberate when querying the bot to prevent it from generating
titlecards for every post in a thread. Currently it trims all the mentions from
the start of the message and checks to see if that list of mentions exactly
matches its username.

As such, to trigger the bot, type something like this:

```
@iasipbot@botsin.space "The Gang Uses an IASIP Title Card Generator"
```

You can @ other people inside of a message (so long as it's surrounded in the
usual IASIP quotes, or it's not directly next to the bot mention, i.e. separated
by a new line or two spaces):

```
@iasipbot@botsin.space "@Gargron@mastodon.social Breaks The Website"
```
