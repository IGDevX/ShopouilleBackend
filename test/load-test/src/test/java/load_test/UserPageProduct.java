package load_test;

import static io.gatling.javaapi.core.CoreDsl.*;
import static io.gatling.javaapi.http.HttpDsl.*;

import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;
import java.time.Duration;
import java.util.Random;
/*
 * The goal of this test is to simulate a fix number of users browser random product pages to check the performance of the product listing endpoint. 
 * Especially focusing on P95 response time.
 */
public class UserPageProduct extends Simulation {
  private static final HttpProtocolBuilder httpProtocol = http.baseUrl("http://catalog-test.mathiaspuyfages.fr")
      .acceptHeader("application/json")
      .userAgentHeader(
          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36");

  private static final Random RNG = new Random();
  private static final ScenarioBuilder scenario = 
  scenario("Random Product Browsing")
    .repeat(5).on(
      exec(session -> session.set("randomPage", 1 + RNG.nextInt(20)))
      .exec(
        http("Get random product page")
          .get("/product?pageIndex=#{randomPage}&pageSize=3")
      )
      .pause(Duration.ofSeconds(3), Duration.ofSeconds(8))
    );

  {
    setUp(
      scenario.injectOpen(
        rampUsers(500).during(Duration.ofMinutes(30))
      )
    ).protocols(httpProtocol);
  }
}