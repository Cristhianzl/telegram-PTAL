# SOLID principles

The five principles, each with the test that detects a violation, a wrong example, a right example, and when to apply or skip.

Examples below are in Python. The same shapes apply in TypeScript, Java, Go, C#, etc.

---

## SRP — Single Responsibility Principle

> A class should have one, and only one, reason to change.

Each class, function, or module encapsulates a single responsibility. Two reasons to change = it's doing too much.

**The test:** Can you describe it in **one sentence WITHOUT "and" or "or"?** If not, split it.

**Wrong — multiple responsibilities:**
```python
class OrderProcessor:
    def validate_order(self, order): ...    # Validation
    def calculate_total(self, order): ...   # Business logic
    def save_to_database(self, order): ...  # Persistence
    def send_confirmation(self, order): ... # Communication
```

**Right — each class has one reason to change:**
```python
class OrderValidator:
    def validate(self, order): ...

class OrderCalculator:
    def calculate_total(self, order): ...

class OrderRepository:
    def save(self, order): ...

class OrderNotifier:
    def send_confirmation(self, order): ...
```

---

## OCP — Open/Closed Principle

> Software entities should be open for extension, but closed for modification.

When a new requirement appears, add behavior by creating new code — not by changing existing, tested code.

**The test:** When a new variant is requested, do you need to modify existing functions with a new `if/elif/else` branch? If yes, OCP is violated.

**Wrong — every new discount type requires modifying this function:**
```python
def apply_discount(customer_type, amount):
    if customer_type == "regular":
        return amount * 0.05
    elif customer_type == "premium":
        return amount * 0.10
    elif customer_type == "vip":         # Added later — modifying existing code
        return amount * 0.15
```

**Right — new types are added by extension, not modification:**
```python
class DiscountStrategy(ABC):
    @abstractmethod
    def calculate(self, amount): ...

class RegularDiscount(DiscountStrategy):
    def calculate(self, amount): return amount * 0.05

class PremiumDiscount(DiscountStrategy):
    def calculate(self, amount): return amount * 0.10

# Adding VIP = new class, zero changes to existing code
class VIPDiscount(DiscountStrategy):
    def calculate(self, amount): return amount * 0.15

def apply_discount(strategy: DiscountStrategy, amount):
    return strategy.calculate(amount)
```

**When to apply:**
- Function has 3+ branches for type/category and new types are expected.
- An `if/elif` chain grows with every feature request.
- Different behaviors share the same interface but vary in implementation.

**When NOT to apply (see YAGNI):**
- Only 2 branches and no evidence more will come.
- The branching logic is genuinely stable.

---

## LSP — Liskov Substitution Principle

> Subtypes must be substitutable for their base types without altering the correctness of the program.

If you inherit from a base class or implement an interface, the derived class must honor the contract: no surprising restrictions, no unexpected errors, no silent behavior changes.

**The test:** Replace a parent instance with a child instance in any function that uses the parent. Does everything still work? If not, LSP is violated.

**Wrong — `Square` changes the behavior contract of `Rectangle`:**
```python
class Rectangle:
    def __init__(self, width, height):
        self._width = width
        self._height = height

    def set_width(self, value):  self._width = value
    def set_height(self, value): self._height = value
    def area(self):              return self._width * self._height

class Square(Rectangle):    # VIOLATES LSP
    def set_width(self, value):
        self._width = value
        self._height = value  # Unexpected side effect

    def set_height(self, value):
        self._width = value
        self._height = value

# Code that works with Rectangle breaks with Square:
def resize(shape: Rectangle):
    shape.set_width(5)
    shape.set_height(4)
    assert shape.area() == 20  # FAILS with Square — area is 16
```

**Right — separate types, no broken contracts:**
```python
class Shape(ABC):
    @abstractmethod
    def area(self): ...

class Rectangle(Shape):
    def __init__(self, width, height):
        self.width = width
        self.height = height
    def area(self): return self.width * self.height

class Square(Shape):
    def __init__(self, side):
        self.side = side
    def area(self): return self.side ** 2
```

**Common LSP violations to watch for:**
- A subclass that throws `NotImplementedError` on an inherited method.
- A subclass that silently ignores a method call.
- A subclass that narrows accepted inputs (parent accepts any number; child only accepts positive).
- A subclass with side effects the parent doesn't have.

---

## ISP — Interface Segregation Principle

> Clients should not be forced to depend on interfaces they do not use.

Keep interfaces small and focused. If a class must implement methods it doesn't need, the interface is too broad.

**The test:** Does the implementing class have methods that are empty, raise `NotImplementedError`, or return dummy values just to satisfy the interface? If yes, ISP is violated.

**Wrong — bloated interface forces useless implementations:**
```python
class Worker(ABC):
    @abstractmethod
    def work(self): ...
    @abstractmethod
    def eat(self): ...
    @abstractmethod
    def sleep(self): ...

class Robot(Worker):           # Robot doesn't eat or sleep!
    def work(self): ...
    def eat(self): pass        # Useless — forced by interface
    def sleep(self): pass      # Useless — forced by interface
```

**Right — segregated interfaces, each class uses only what it needs:**
```python
class Workable(ABC):
    @abstractmethod
    def work(self): ...

class Feedable(ABC):
    @abstractmethod
    def eat(self): ...

class Restable(ABC):
    @abstractmethod
    def sleep(self): ...

class Human(Workable, Feedable, Restable):
    def work(self): ...
    def eat(self): ...
    def sleep(self): ...

class Robot(Workable):         # Only what it actually does
    def work(self): ...
```

**Practical signs of ISP violation:**
- Methods with `pass`, `return None`, or `raise NotImplementedError`.
- Interfaces with 7+ methods where most implementors only use 2-3.
- A change in one interface method triggers changes in classes that don't use that method.

---

## DIP — Dependency Inversion Principle

> High-level modules should not depend on low-level modules. Both should depend on abstractions.

The business logic (high-level) never directly instantiates or references infrastructure (database, API client, file system). Define an interface; inject the implementation.

**Wrong — high-level directly depends on low-level:**
```python
class UserService:
    def __init__(self):
        self.db = MySQLDatabase()  # Hard dependency
```

**Right — both depend on abstraction:**
```python
class UserRepository(ABC):
    @abstractmethod
    def find_by_id(self, user_id): ...

class MySQLUserRepository(UserRepository):
    def find_by_id(self, user_id): ...

class UserService:
    def __init__(self, repo: UserRepository):  # Injected
        self.repo = repo
```

Injection makes the high-level code:

- **Testable** without spinning up the real DB.
- **Swappable** between implementations (MySQL, Postgres, in-memory).
- **Decoupled** from the framework: the domain layer doesn't import the DB driver.
