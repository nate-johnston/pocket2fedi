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
post each new save as a post to a Mastodon server. Use the
https://github.com/mattn/go-mastodon and https://github.com/motemen/go-pocket
modules. The program should include error handling and informative log messages
for debugging. There should also be a good set of unit tests for the program.
```

After I have completed work on this, I plan to retrospect on the tangible
lessons learned about the quality of the AI contributions - where they are
strong and where they are weak.  

Everything after this point was provided by the AI.

## Running the Code

- Install Dependencies - ```go mod install```
- Set environment variables
```
export POCKET_CONSUMER_KEY="YOUR_POCKET_CONSUMER_KEY"
export POCKET_ACCESS_TOKEN="YOUR_POCKET_ACCESS_TOKEN"
export MASTODON_SERVER="YOUR_MASTODON_SERVER_URL"
export MASTODON_TOKEN="YOUR_MASTODON_ACCESS_TOKEN"
```
Replace the placeholders with your actual values. Alternatively, you can set
these as system environment variables.
- Run the Program: `go run .`
- Run the Tests: `go test ./...`

## Ideas for Future Improvements

- Error Handling: The code includes basic error handling, but you might want to implement more sophisticated error logging and potentially retry mechanisms for network-related issues.
- Rate Limiting: Be aware of the API rate limits for both Pocket and Mastodon. The included time.Sleep is a basic measure; you might need a more robust rate limiting strategy for frequent execution.
- More Detailed Pocket Data: The current implementation fetches basic details. You can adjust the DetailType in the api.RetrieveInput to get more information from Pocket if needed.
- Mastodon Formatting: You might want to customize the format of the Mastodon posts further.
- Authentication: This program assumes you already have Pocket and Mastodon access tokens. The go-pocket/auth package can be used to implement the initial OAuth flow if you need to obtain these tokens programmatically.
- State Management: If you want to avoid posting the same Pocket saves repeatedly, you'll need to implement some form of state management (e.g., storing the IDs of already posted items).
- Error Handling Strategies: Implement retry mechanisms for transient API errors.
- Logging Levels: Introduce different logging levels (e.g., debug, info, error) for more granular control over the output.
- Concurrency: If you need to process a large number of Pocket saves, consider using Go's concurrency features (goroutines and channels) to speed up the process.
