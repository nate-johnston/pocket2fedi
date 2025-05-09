# pocket2fedi

## Purpose

Tool to poll the Pocket API for recent saves and post them to a Fediverse account

## AI Code Use

This project is a one-off to see how quickly I could use AI - specifically
Google Gemini - to design a separate standalone application.  Rather than taking
someone else's word for it, I wanted to set a specific task and see how the
quality of code created by Gemini would work.  Here is my prompt:

```
As an expert go programmer, design a program written in go that will use the GET
api from Pocket to pull recent Pocket saves, and then use the Mastodon API to
post each new save as a post to a Mastodon server. The program should include
error handling and informative log messages for debugging. There should also be
a good set of unit tests for the program.
```

After I have completed work on this, I plan to retrospect on the tangible
lessons learned about the quality of the AI contributions - where they are
strong and where they are weak.  

Everything after this point was provided by the AI.

## Running the Code

- Install Dependencies - ```go get github.com/joho/godotenv```
- Set Environment Variables - Create a .env file in the same directory with your
  API keys and server details:
```
POCKET_CONSUMER_KEY=YOUR_POCKET_CONSUMER_KEY
POCKET_ACCESS_TOKEN=YOUR_POCKET_ACCESS_TOKEN
MASTODON_SERVER=YOUR_MASTODON_SERVER_URL
MASTODON_TOKEN=YOUR_MASTODON_ACCESS_TOKEN
```
Replace the placeholders with your actual values. Alternatively, you can set
these as system environment variables.
- Run the Program: `go run .`
- Run the Tests: `go test .`

## Ideas for Future Improvements

- Tracking Posted Items: Implement a mechanism (e.g., a local file or a simple
  database) to track which Pocket items have already been posted to Mastodon.
- This prevents duplicate posts on subsequent runs.  More Informative Posts:
  Include more details from the Pocket item in the Mastodon post (e.g., tags,
  excerpt).
- Rate Limiting: Implement more sophisticated rate limiting to avoid getting
  blocked by either API
- Configuration Management: Consider using a more robust configuration library
  for handling different environments and file formats.
- Error Handling Strategies: Implement retry mechanisms for transient API
  errors.
- Logging Levels: Introduce different logging levels (e.g., debug, info, error)
  for more granular control over the output.
- Concurrency: If you need to process a large number of Pocket saves, consider
  using Go's concurrency features (goroutines and channels) to speed up the
  process.
