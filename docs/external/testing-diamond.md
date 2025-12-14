# Testing Diamond

external sources may refer to this file as "the testing diamond" (as opposed to the traditional testing pyramid) due to the emphasis on integration tests.

## code4it

source: https://www.code4it.dev/architecture-notes/testing-pyramid-vs-testing-diamond/

copywrite original author [Davide Bellone ](https://github.com/bellons91)

content captured here for transparency and ease of reference within this repository.

### content begins here

Every software project requires some sort of testing. You can execute them manually or put in place a whole suite of automatic tests that run timely, for example when you close a PR or daily, every day at the same time.

Still, not every project requires the same amount of tests, nor the same kind of tests.

Every test has a purpose to exist and has some intrinsic pros and cons. Talking about Unit Tests, the more tests you write, the more difficult it will be to change them if you update the whole architecture of the project.

Therefore, we should carefully choose the right strategy to test the most critical parts of our system. Meaningless tests can cause more trouble than benefits.

In this article, we’re gonna learn about the types of tests, their purposes, and the best testing strategies we can apply to our projects.

#### Different types of testing: Unit Tests vs Integration Tests vs End-to-end Tests

As we introduced before, there are lots of types of tests that you can write. Each test has a purpose, some pros, and some cons.

##### Unit Tests: easy to write, fast to run, but coupled with implementation details

Usually, when we think about tests, we implicitly think about Unit Tests.

As we all know, a Unit Test is a test that refers to a unit. Ok, but what the heck is a unit? There is no official definition: in fact, everything can be a unit. Or, better, everything that can be tested independently is a unit.

You can write Unit Tests about a class. For example, when you create a Priority Queue - like the one recently released in .NET 🔗 - you can write tests on the initialization and on the methods it exposes. For example, when you enqueue an item you want to test that the total count is increased. When you dequeue an item, you want to test that the retrieved item is the one with the highest priority. In this case, the whole class is the Unit.

Or you can write a test on a specific method; an example is the application of a Regular Expression: you might want to have a set of Unit Tests that verify that the RegEx works correctly with some input test strings and fails correctly with invalid strings. You are testing the correctness of the application of a single method. In this case, a single method is the Unit.

In the end, Unit Tests are the most common type of tests.

They have some advantages:

- they are fast to write;
- they are fast to execute;
- they don’t require access to external dependencies, such as a database;

But they also have some disadvantages:

- they are tightly coupled with the production code: if you add a new parameter to - the method under test, you have to update all the tests that use the newly-modified - method;
- they don’t test the interaction between real components: you must use fakes, stubs, - and mocks to test the correctness of your tests;
- they don’t help in discovering side effects of a change: say that you have a - fictious method Sanitize(string text), exposed by the class StringUtilities. Sanitize returns an empty string if the text variable is invalid. Now you change the implementation to return null instead of an empty string if the input string is invalid. And say that an external class TextHandler injects, using DI, the StringUtilities class in order to use the Sanitize method. When writing tests for the TextHandler method, you don’t create a concrete instance of StringUtilities, but rely on stubs or mocks. Then, if you change the implementation of Sanitize without updating also the tests of TextHandler, you won’t notice the change and the related side effects, and the production code is probably going to break because you are performing operations on a null value instead of an empty string.

#### Integration Tests: focus on the interaction between components, but harder to write

Integration Tests are meant to validate that different parts of the system (for example, two concrete classes) work well together.

One kind of Integration Test that I find particularly effective is the one that tests API functionalities: you spin up an in-memory test server that runs your APIs with all the concrete internal dependencies (no stubs or mocks), and then you call such endpoints and validate the final result.

You can also - as I’d recommend - use an in-memory database instead of a real one. And all the external dependencies, such as external API systems, to ensure that your tests don’t have dependencies upon systems you can’t control.

We can list some advantages of Integration Tests:

- they ensure that the system, with real classes, works correctly;
- they test upon functionalities, not implementations. If you add a new parameter to an internal method, you don’t have to worry about updating the tests: implementation details do not matter;
- you can focus on the expected result of the whole elaboration; say that your system needs to integrate data coming from different sources: with Integration Tests, you can mock such sources with realistic data and ensure that the whole execution brings the expected result.
- they are easy to maintain: since they’re not tightly coupled with the production code, they change only if the functionalities change;
- they can act as real-life documentation: your tests explain how the system behaves instead of how the single parts of the system are built.

Clearly, there are some downsides:

- they are harder to write if compared to Unit Tests;
- a single Integration Test is slower than a single Unit Test, because you have to spin up a whole server and execute complex operations. Clearly, if you expect your system to respond in X milliseconds, you can expect your test to run in around X milliseconds (plus the initialization of the server);

#### E2E Tests: use concrete external dependencies to ensure real-life tests

Finally, End-to-end (E2E) Tests are the most similar to manual tests.

They run against the whole stack of the application, using real instances of frontend, backend, database, and external services, and run in a system almost identical to the one used by real users.

They are the most difficult of the three to write, but they are also incredibly useful to cover the most important execution paths.

An example of an E2E Test could be made of the following steps:

1. the user logs in;
1. the user searches for a product;
1. the user adds a product to the cart;
1. the user confirms the payment;

that would bring to the following assertions:

1. the user can complete those paths successfully;
1. the order is received by the system;
1. the payment is correctly handled by the third-party system;
1. an email is sent to the user

As you can see, we cover the most critical user paths.

Again, we have advantages and disadvantages:

Advantages:

- they cover business-critical paths;
- they use production-like systems and dependencies, to ensure that the system as a whole works correctly;

Disadvantages:

- the hardest to write;
- the slowest to run;
- can be difficult to maintain: if your tests rely on the UI, and it slightly changes (for example, you change the class of the button that sends the Complete Checkout command), the whole E2E fails.

#### Testing Pyramid: focus on Unit Tests

The “Testing Pyramid” is a way to think of a testing strategy where the main focus is on Unit Tests.

Testing Pyramid: the main focus is on Unit Tests

Have a look at the proportions between testing types. Using Testing Pyramid, we have:

- a lot of Unit Tests: run a lot of fast tests to cover most of the functionalities;
- a bunch of Integration Tests: since they are slower, but still important, run just a bunch of them;
- a few E2E Tests: sometimes you can even skip them, but in general it’s a good practice to cover at least the most critical paths.

#### Testing Diamond: focus on Integration Tests

Testing Diamond is another way to think of a testing strategy, opposed as to the Testing Pyramid. Here, our focus is on Integration Tests.

Testing Diamond: the main focus is on Integration Tests

Unit Tests, Integration Tests, and E2E Tests have a different proportions:

- a few Unit Tests: write only strictly necessary Unit Tests: for example, data transformation, text parsing, and complex calculations;
- a lot of Integration Tests: since they cover business-critical functionalities, here is where you should spend most of your time and effort;
- a bunch of E2E Tests: again, cover only the critical operations.

#### What to choose?

As we saw, Testing Pyramid is focused more on fast and easy-to-write tests, while Testing Diamond is more about business-critical tests.

Below you can see a GIF that explains why Integration Tests are more important than Unit Tests: the door alone works correctly, and the lock works correctly, so Unit Tests for both systems pass correctly. But the overall system does not work: the Integration Tests would fail.

Unit Tests vs Integration Tests

So, what to choose? As always, there is no silver bullet: it always depends on your project and your team.

But, in general, I’d prefer Testing Diamond: our tests should verify that the overall system works properly, without worrying about implementation details.

If you rename a class or change the internal data access (eg: you transform a static class into a service), you don’t want to update the tests: as long as the final result is correct, the test is more than enough.

#### Code Coverage and Testing Strategies

As I often say, reaching a high Code Coverage is not a useful goal: a system with 100% test coverage still has bugs - it’s just that you haven’t found them yet.

Also, if you mark every class as to be ignored by code coverage (in C#, using ExcludeFromCodeCoverage), you can have 100% code coverage without a single test.

So, don’t write tests for the sake of reaching 100% code coverage; on the contrary, use code coverage to see which parts of the systems have not been tested.

In my opinion, code coverage works best with Testing Pyramid: write Integration Tests, and make sure that your system works correctly. Then use code coverage indicators to learn which parts have not been tested (for example, you validated an empty string but not a null value). You can use Unit Tests to cover those specific cases.

#### Further readings

There are lots of articles out there about Unit Tests, and they often say the same things. So, I want to highlight an article that focuses on Integration Tests and E2E Tests: this article is well-written and explores such topics in detail:

🔗 [End-to-End Testing vs Integration Testing | Testim](https://www.testim.io/blog/end-to-end-testing-vs-integration-testing/)

You might be wondering: how can we run Integration Tests for .NET APIs? I wrote an article with a simple, basic approach, that can help you understand it:

🔗 [How to run Integration Tests for .NET API | Code4IT](https://www.code4it.dev/blog/integration-tests-for-dotnet-api/)

## steven-giesel

source: https://steven-giesel.com/blogPost/86b6fae7-95a7-44fa-a85a-00ee1b6dd697

copywrite original author Steven Giesel

content captured here for transparency and ease of reference within this repository.

### content begins here

In this article, we will discuss the testing pyramid - what it is and what are some problems with that.

We will also discuss a different approach: The testing diamond.

#### Testing Pyramid

The testing pyramid is a testing strategy that emphasizes a test suite's composition and its coverage across different levels of the application. The pyramid shape implies that there should be fewer high-level end-to-end (E2E) tests and more low-level unit tests, with a moderate number of integration tests in the middle.



- Unit Tests: These are small, fast, and isolated tests that verify the behavior of a single unit of code in isolation, typically a single function or method. Unit tests ensure that each individual unit of code behaves correctly.

- Integration Tests: These are medium-level tests that verify that the interactions between different units of code work correctly. Integration tests are usually slower and more complex than unit tests and may require more setup and teardown.

- End-to-End Tests: These high-level tests verify that the entire application works as expected. E2E tests simulate user behavior and interactions and are typically slower and more brittle than unit or integration tests.


So the biggest foundation is unit tests. And if you are following patterns like TDD (Test-driven development) then this is no news to you. But there are some problems with unit tests:


- False Sense of Security: Unit tests only test individual units of code in isolation and may not capture bugs that arise when units are integrated. Over-reliance on unit tests may give a false sense of security.
- Limited Coverage: Unit tests may not cover all possible scenarios or edge cases, leading to insufficient test coverage. Integration and E2E tests are needed to ensure the entire application works correctly.
- Slow Feedback Loop: Although unit tests are fast, the feedback loop may be slow when codebase changes require many tests. This can slow down development and impede iteration.

\Especially the last point is something I have seen multiple times in the past - a solid test suite with thousands over thousands of tests. Of course, one could argue that those tests were not perfect, so we suffered from this. And this is totally valid, but I am also not the only one encountering those problems.

Another issue is that, especially with microservice-oriented architecture or bigger Web-API projects, you are not really in the position of your client. Your client, in the form of your front-end, is using your API or services. To enable that, we can shift from a big foundation of unit tests to a big foundation of integration tests:

The advantage of this is that you can create tests that cover a whole business requirement in the vertical stack! So you can name your tests accordingly (Hey, BDD!). You are also testing (de)serialization, communication, and your DI-Container. This is especially helpful in a micro-service-oriented architecture (or Web-API's). The best thing is that you write fewer tests overall with some quality outcomes. Please don't get me wrong, there is still plenty of space for unit tests. Imagine scenarios like caching, where it is quite hard to make a useful and stable setup from an API point of view.

Now you might ask yourself if I have to spin up a real DB and backend every time, isn't that slow? Well, let's have a look at [Entity Framework](https://github.com/dotnet/efcore) - they have over 3300 tests and connect to a cosmos database (locally emulated via Azure Cosmos DB emulator) for the majority of tests where a real database would be needed. The overall runtime is something from 4 to 7 minutes. Quite fast. And for the server?

#### ASP.NET Core Web API

For ASP.NET Core Web API you can utilize the WebApplicationFactory. Oversimplified, it is an in-memory server that you can spin up in your integration tests in a matter of milliseconds. I will not go into too much detail here, because I already wrote a whole article about this: ["Introduction to WebApplicationFactory".](https://steven-giesel.com/blogPost/cd62475b-2c7d-4ce2-bd97-9670f91ebac8)

#### Conclusion

With a shift towards diamond-shaped testing, you can write less, but more meaningful tests that are really a living documentation.

## krython

retrieved 2025-12-14
from https://www.krython.com/tutorial/python/testing-strategies-pyramid-and-diamond

all copywrite original author Krython

content captured here for transparency and ease of reference within this repository.

### content begins here

#### 🤔 What is the Testing Pyramid?

The Testing Pyramid is like building a house 🏠 - you need a strong foundation of many small unit tests, a middle layer of integration tests, and a smaller roof of end-to-end tests!

In Python terms, this means organizing your tests in layers:

- ✨ Unit Tests (Base) - Fast, focused, numerous
- 🚀 Integration Tests (Middle) - Test component  interactions
- 🛡️ End-to-End Tests (Top) - Full system validation

#### 💡 What is the Testing Diamond?


The Testing Diamond is a modern evolution that looks like a diamond shape 💎. It emphasizes more integration tests in the middle, recognizing that they often catch the most valuable bugs!

Here’s why developers love these strategies:

- Fast Feedback ⚡: Unit tests run in milliseconds
- Confidence 💪: Multiple layers catch different bugs
- Cost-Effective 💰: Find bugs early when they’re cheap to fix
- Maintainable 🔧: Clear structure for test organization

Real-world example: Imagine building an e-commerce API 🛒. Unit tests check individual functions, integration tests verify database interactions, and E2E tests ensure customers can complete purchases!

#### 🧙‍♂️ Test Strategy Selection

```python
# 🎯 Strategy Decision Framework

class TestStrategySelector:
    """Choose the right testing approach! 🤔"""
    
    @staticmethod
    def analyze_project(project_type, team_size, complexity):
        """Recommend testing strategy based on project"""
        
        # 🏔️ Use Testing Pyramid when:
        if (project_type in ["library", "utility", "algorithm"] or
            team_size < 5 or
            complexity == "low"):
            return {
                "strategy": "pyramid",
                "reason": "Many isolated components, clear boundaries",
                "distribution": {
                    "unit": 70,      # 70% unit tests
                    "integration": 20,  # 20% integration
                    "e2e": 10          # 10% end-to-end
                }
            }
        
        # 💎 Use Testing Diamond when:
        elif (project_type in ["microservice", "api", "web_app"] or
              complexity == "high" or
              team_size > 10):
            return {
                "strategy": "diamond",
                "reason": "Complex interactions, many integrations",
                "distribution": {
                    "unit": 20,        # 20% unit tests
                    "integration": 60,  # 60% integration (the bulk!)
                    "e2e": 20          # 20% end-to-end
                }
            }
        
        # 🔄 Hybrid approach
        else:
            return {
                "strategy": "hybrid",
                "reason": "Balanced approach for medium complexity",
                "distribution": {
                    "unit": 40,
                    "integration": 40,
                    "e2e": 20
                }
            }

# 🪄 Advanced Testing Patterns
class AdvancedTestPatterns:
    """Level up your testing game! 🚀"""
    
    @pytest.fixture
    def time_machine(self):
        """Control time in tests! ⏰"""
        with freeze_time("2024-01-01") as frozen:
            yield frozen
    
    def test_subscription_expiry(self, time_machine):
        # 📅 Test time-dependent features
        user = User(subscription_end=datetime(2024, 1, 15))
        
        assert user.has_active_subscription() == True
        
        # Fast forward time! 🚀
        time_machine.move_to("2024-01-16")
        assert user.has_active_subscription() == False
    
    @pytest.mark.parametrize("test_input,expected", [
        ({"amount": 100, "country": "US"}, 108.5),
        ({"amount": 100, "country": "UK"}, 120),
        ({"amount": 100, "country": "JP"}, 110),
    ])
    def test_international_pricing(self, test_input, expected):
        # 🌍 Test multiple scenarios efficiently
        result = calculate_total_with_tax(**test_input)
        assert result == expected
```

#### 🏗️ Test Organization Patterns

Structure your tests for maximum maintainability:

```
# 🚀 Advanced Test Organization

# tests/
# ├── unit/              # 🔹 Fast, isolated tests
# │   ├── test_models.py
# │   ├── test_utils.py
# │   └── test_validators.py
# ├── integration/       # 💎 Component interaction tests
# │   ├── test_api_endpoints.py
# │   ├── test_database_operations.py
# │   └── test_external_services.py
# ├── e2e/              # 🌐 Full system tests
# │   ├── test_user_journeys.py
# │   └── test_critical_paths.py
# └── fixtures/         # 🛠️ Shared test utilities
#     ├── factories.py
#     └── mocks.py

# 🎯 Smart test categorization with markers
@pytest.mark.unit
@pytest.mark.fast
def test_calculate_discount():
    """Runs in milliseconds! ⚡"""
    pass

@pytest.mark.integration
@pytest.mark.slow
@pytest.mark.requires_db
def test_order_processing():
    """Needs database connection 🛢️"""
    pass

@pytest.mark.e2e
@pytest.mark.critical
def test_checkout_flow():
    """Critical business path 🎯"""
    pass

# Run specific test categories
# pytest -m "unit"           # Run only unit tests
# pytest -m "not slow"       # Skip slow tests
# pytest -m "critical"       # Run critical tests only
```

#### 🤯 Pitfall 2: Testing Implementation Instead of Behavior

```
# ❌ Testing HOW instead of WHAT
def test_bad_implementation_focused():
    calculator = Calculator()
    
    # 😰 Testing internal details!
    assert calculator._internal_buffer == []
    calculator.add(5)
    assert calculator._internal_buffer == [5]
    assert calculator._operation_count == 1

# ✅ Testing behavior and outcomes!
def test_good_behavior_focused():
    calculator = Calculator()
    
    # 🎯 Test what users care about!
    calculator.add(5)
    calculator.add(3)
    assert calculator.result() == 8
    
    calculator.multiply(2)
    assert calculator.result() == 16
```

#### 🛠️ Best Practices


1. 🎯 Right Tool for the Job: Use pyramid for algorithms, diamond for services
2. ⚡ Fast Feedback Loop: Unit tests should run in seconds, not minutes
3. 🛡️ Test Isolation: Each test should be independent
4. 📊 Meaningful Coverage: Aim for behavior coverage, not line coverage
5. ✨ Maintainable Tests: Tests are code too - keep them clean!
