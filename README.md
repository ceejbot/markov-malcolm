# markov-malcolm

[Malcolm Tucker](https://en.wikipedia.org/wiki/Malcolm_Tucker) markov-chains antisocial bot. Not polite in any way. A Golang learning project that is nothing more than a clone of [the javascript version](https://github.com/ceejbot/malcolm_ebooks).

## Usage

Clone the repo. Create a file named `.env` in the repo directory and put your Mastodon access tokens in it like so:

```sh
MASTO_URL=base-url-of-instance
MASTO_TOKEN=auth-token-to-use
```

Optionally, make a subdirectory named `images` and toss in a few animated gifs of Malcolm in action. He will post a randomly-selected image every so often, along with a randomly-generated comment. No attempt is made to filter files in the `images` directory for actual images.

Then run `go build && ./markov-malcolm`. Malcolm will now be online and swearing incoherently.

## TODO

Replies to mentions next.

## License

ISC
