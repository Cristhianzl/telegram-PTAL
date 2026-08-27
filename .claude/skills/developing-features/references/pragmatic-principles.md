# Pragmatic principles

DRY, KISS, YAGNI, and the Law of Demeter. The four heuristics that catch over-engineering and under-design.

---

## DRY — Don't Repeat Yourself

> Every piece of knowledge must have a single, unambiguous, authoritative representation within a system.

- Extract repeated logic (5+ lines appearing 2+ times) into named functions.
- **But prefer duplication over wrong abstraction.** If two blocks look similar by coincidence but change for different reasons, keep them separate.
- "Wrong abstraction" = premature generalization, unclear purpose, or coupling unrelated concerns. **Not** the same as extracting obvious duplication with clear intent.

The smell that says "wrong abstraction":
- The function has 4+ boolean parameters that switch its behavior.
- Two callers pass nearly opposite arguments to get the behavior they want.
- A comment in the helper says "if X, do A; if Y, do B" — A and B are different concepts being squeezed into one function.

When two pieces of code look the same but evolve independently, leave them duplicated. Re-evaluate after the next change.

---

## KISS — Keep It Simple

> The simplest solution that meets the requirements is the best one.

- Prefer small, composable, single-purpose functions and classes.
- Avoid unnecessary abstractions and indirection.
- Don't create a factory for something a simple function handles.
- Don't use generics when a concrete type is clear and stable.
- If you wrote 200 lines and 50 would do, rewrite it.

The test: "Would a senior engineer skim this and say it's overcomplicated?" If yes, simplify.

---

## YAGNI — You Ain't Gonna Need It

> Don't implement anything until it is actually needed.

Don't write code for speculative future requirements. Build for today, refactor tomorrow when the need is real.

**The test:** Is this code solving a problem that exists **right now**, or a problem I think **might** exist later?

**Wrong — building for imaginary future requirements:**
```python
class NotificationService:
    def __init__(self, strategy: NotificationStrategy):
        self.strategy = strategy
        self.retry_policy = ExponentialBackoffRetry()  # Nobody asked for retry
        self.circuit_breaker = CircuitBreaker()         # Nobody asked for this
        self.rate_limiter = RateLimiter(100)            # Nobody asked for this
        self.dead_letter_queue = DeadLetterQueue()      # Nobody asked for this
```

**Right — solves the current requirement, nothing more:**
```python
class NotificationService:
    def __init__(self, sender: NotificationSender):
        self.sender = sender

    def notify(self, user_id, message):
        self.sender.send(user_id, message)
```

**Common YAGNI violations:**
- Adding configuration parameters nobody requested.
- Creating abstract base classes when only one implementation exists.
- Building plugin systems for a feature with one plugin.
- Adding caching before measuring whether there's a performance problem.
- Creating a generic "event bus" when only one event type exists.

---

## Law of Demeter — Principle of Least Knowledge

> A method should only talk to its immediate friends — not to strangers.

An object should only call methods on:
- itself,
- its own fields,
- its parameters, and
- objects it creates locally.

Long chains of calls (`a.b.c.d.method()`) reveal that the caller knows too much about the internal structure of other objects.

**The test:** Count the dots. If a call chain goes deeper than one dot from your direct dependency, you're likely violating Demeter.

**Wrong — knows the internal structure of order → customer → address:**
```python
def get_shipping_label(order):
    street = order.customer.address.street
    city = order.customer.address.city
    zip_code = order.customer.address.zip_code
    return f"{street}, {city} - {zip_code}"
```

**Right — ask, don't dig:**
```python
def get_shipping_label(order):
    return order.get_shipping_address()

# The Order class delegates internally:
class Order:
    def get_shipping_address(self):
        return self.customer.get_formatted_address()
```

**Common violations:**
```python
# Deep chain — caller knows too much
user.profile.settings.notifications.email_enabled

# Encapsulated — caller asks the right object
user.is_email_notification_enabled()
```

**Grep to detect long chains:**
```bash
grep -rn "\w\.\w\+\.\w\+\.\w\+" --include="*.py" --include="*.ts" --include="*.js" \
  | grep -v "import\|require\|from\|test\|spec\|mock\|node_modules"
```

When **not** to apply Demeter strictly:
- Fluent builders (`query.where(...).order_by(...).limit(...)`) — by design.
- Data-class chains where every step is a plain field accessor with no behavior. Still a smell, but lower priority than the same chain through methods.
