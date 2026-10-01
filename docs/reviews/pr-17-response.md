# Pull request 17 - author response

Date: 2026-10-01. Review round 1 recorded effective head `21ed998`, with the verdict "Changes required".

## P2-1: Cache-write tokens are undercounted

**Result: full merit.**

The author read both pages of the finding on 2026-10-01 (PC-102 of `docs/research/platform-cloud-and-ai.md`). The pricing page lists a cache write of `gpt-6-luna` at 0.125 USD for each million tokens, and the input rate is 0.10 USD. The cost example of the prompt-caching guide reads `usage.input_tokens_details.cache_write_tokens`, and subtracts the cached tokens and the cache-write tokens from the input tokens. So a cache write is part of the input tokens, at its own rate.

At `21ed998`, `Prices.Cost` billed each cache write at the input rate, and the OpenAI provider did not read the field. The reservation of the cap used the input rate for each input byte. So the cost of a call of cache writes was more than its reservation.

The same pricing page gives higher prices above 272K input tokens. The reservation at `21ed998` used the prices of the short context for any size of request. The correction covers this case too.

**Correction.**

- `go/internal/ai/cost.go`: `Prices` holds the cache-write rate, and `Usage` holds the cache-write tokens. `Cost` bills each input token at one of the three input rates. The reservation bills each input byte at the highest input rate.
- `go/internal/ai/role.go`: the prices of the roles hold the cache-write rate of 125 billionths of a dollar for each token. Each role refuses a request of more than 272,000 bytes, because one token holds one byte or more. So each call stays in the short context.
- `go/internal/ai/client.go`: the client refuses a request over the limit before the cap and before the call.
- `go/internal/ai/openai.go`: the provider reads `input_tokens_details.cache_write_tokens`.
- `docs/research/platform-cloud-and-ai.md`: PC-102, and the prices of the short and the long context.

**Regression checks.**

- `TestOpenAICacheWriteCost`: a reply of the local server with cache writes gives the cost of the published rates. The old code gave a lower cost.
- `TestWorstCoversEachRate`: the reservation covers each mix of the input rates. The old reservation was below the cost of a call of cache writes.
- `TestCapBoundary`: the cap permits a call at its reservation, and refuses the call one billionth of a dollar below it.
- `TestRequestLimit`: a request over 272,000 bytes makes no call.
- `make go-test`, `go test -race ./internal/ai`, and `make verify` pass.
