# AI runtime — mandatory rules when the feature calls an LLM

If the feature calls an LLM (or any AI service) at runtime, these rules are non-optional. They apply in addition to the AI guardrails in `developing-features/references/security.md`.

---

## Phase 2 — AI runtime profile

The architecture plan must include an explicit AI runtime profile:

```
AI RUNTIME PROFILE:
  Call placement:   <synchronous in request path | async via queue | background job>
  Timeout:          <explicit value, not the SDK default — e.g., 5s sync / 60s async>
  Circuit breaker:  <thresholds — e.g., open after 5 failures in 30s>
  Fallback path:    <what runs when the LLM is unavailable, slow, or returns garbage>
  Cost ceiling:     <tokens per request, requests per user/day, daily cap>
  Kill switch:      <feature flag or config that disables the AI path without redeploy>
  SLO target:       <P95 latency, success rate, fallback rate>
```

Each item in this profile becomes a **test** in Phase 3.

---

## Phase 3 — mandatory failure-mode tests

Every feature that calls an LLM at runtime must have tests for at least these failure modes:

| Failure mode                                   | Test name pattern                                              |
|------------------------------------------------|-----------------------------------------------------------------|
| LLM call times out                             | `test_should_invoke_fallback_when_llm_times_out`               |
| LLM returns malformed / unparseable output     | `test_should_return_domain_error_when_llm_output_is_invalid`   |
| Circuit breaker is open                        | `test_should_skip_llm_call_when_circuit_breaker_is_open`       |
| Rate limit / safety filter triggered           | `test_should_map_safety_filter_to_user_facing_message`         |
| Kill switch toggled off                        | `test_should_bypass_llm_when_kill_switch_disabled`             |
| Cost ceiling reached                           | `test_should_reject_request_when_user_daily_quota_exceeded`    |

These tests run against a **fake LLM client** that you control — they never hit a real LLM provider. The fake must be capable of simulating each failure mode independently.

---

## Phase 5 — GREEN implementation rules

When making the AI failure-mode tests pass:

- **Explicit timeout on every AI call.** Never rely on SDK defaults — they can be infinity, or hours.
- **Circuit breaker around the call.** If you don't have a library, implement a minimal one — a counter + a state machine is enough.
- **Fallback runs without invoking the LLM.** The fallback is deterministic code that produces a degraded-but-correct result.
- **Kill switch checked first.** Before any other AI logic, check the flag. If off, the fallback runs immediately.
- **Cost accounting is deterministic.** Token count is observable from the response; persist the running total per user per day.
- **All AI errors map to domain errors.** Never let a raw `openai.APIError` or vendor exception reach the caller.

---

## Observability for AI runtime

Each AI call emits a structured log with at minimum:

| Field            | Example                                |
|------------------|----------------------------------------|
| `operation`      | `support_bot.respond`                  |
| `request_id`     | correlation ID from the inbound request |
| `model`          | `claude-sonnet-4-6`                    |
| `input_tokens`   | 1247                                   |
| `output_tokens`  | 312                                    |
| `duration_ms`    | 1843                                   |
| `outcome`        | `success` / `timeout` / `circuit_open` / `safety_filter` / `kill_switch` / `cost_limit` |
| `fallback_used`  | `true` / `false`                       |

Emit a metric (counter + histogram) with the same fields for SLO tracking.

---

## Phase 8 — VALIDATE checklist additions

If the feature calls an LLM:

- [ ] Explicit timeout on every AI call (no SDK defaults).
- [ ] Circuit breaker configured with explicit thresholds.
- [ ] Fallback path implemented **and tested** by simulating an LLM outage.
- [ ] Cost ceiling defined (per request, per user/day, daily cap) with alerts.
- [ ] Kill switch implemented **and verified** (toggling disables AI path with no redeploy).
- [ ] SLO documented (latency, success rate, fallback rate).
- [ ] Observability emits all mandatory fields above.
- [ ] No secrets, no PII, no full prompts/responses in logs (redact before logging).
- [ ] Input validation runs **before** the prompt is constructed (no naive concatenation).
- [ ] Output validation runs **before** the response is returned (parse → verify schema → use).

---

## Anti-patterns

### Default SDK timeout

```python
# Wrong — could hang for minutes
client.messages.create(...)

# Right — explicit, short timeout in the request path
client.with_options(timeout=5.0).messages.create(...)
```

### Fallback that calls the LLM differently

```python
# Wrong — the "fallback" still depends on the LLM
def respond(question):
    try:
        return llm.complete(question, model="claude-opus-4-7")
    except TimeoutError:
        return llm.complete(question, model="claude-haiku-4-5")  # still LLM
```

The fallback must be deterministic, non-LLM code. Acceptable fallbacks: a templated response, a database lookup, a cached previous response, a polite error message.

### Kill switch as an environment variable that requires redeploy

A kill switch you can't toggle in <60 seconds is not a kill switch. Use a feature-flag service, a config row in the DB, or a Redis key — something the on-call engineer can flip without a deploy.

### Cost accounting "in the next sprint"

Daily cost caps are easy to add upfront. They're nearly impossible to retrofit once a feature has shipped and a runaway prompt has cost the company $40k overnight. Always include cost ceilings in the first GREEN.

### Treating LLM output as trusted

```python
# Wrong — output may contain SQL, shell, prompt-injection-targeting-the-next-call
return llm_response

# Right
parsed = json.loads(llm_response)            # may raise
validate_schema(parsed, ResponseSchema)       # validate shape
return parsed
```

If the LLM is asked for JSON, the code must reject non-JSON. If it's asked for an enum value, the code must reject values outside the enum.
