package load_test;

import static io.gatling.javaapi.core.CoreDsl.*;
import static io.gatling.javaapi.http.HttpDsl.*;

import io.gatling.javaapi.core.*;
import io.gatling.javaapi.http.*;
import java.time.Duration;
import java.util.Random;

/*
 * The goal of this scenario is to simulate a burst of users browsing random product pages to check the performance of the product listing endpoint.
 */
public class BurstUserNavigation extends Simulation{

    private static final HttpProtocolBuilder httpProtocol = http.baseUrl("http://catalog-test.mathiaspuyfages.fr")
    .acceptHeader("application/json")
    .userAgentHeader(
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"
    );

    private static final Random RNG = new Random();
    private static final int PAGE_SIZE = 2;
    
    private static final ScenarioBuilder scenario =
        scenario("Random Product Browsing").repeat(10)
            .on(
                // ensure we have a pageSize available
                exec(session -> session.set("pageSize", PAGE_SIZE)),

                // first hit to obtain X-Total-Count header
                exec(
                    http("Get first product page (for total count)")
                        .get("/product?pageIndex=0&pageSize=#{pageSize}")
                        .check(header("X-Total-Count").saveAs("totalCount"))
                ),


                exec(session -> {
                    String totalCountStr = session.getString("totalCount");
                    int totalCount = 0;
                    if (totalCountStr != null && !totalCountStr.isEmpty()) {
                        try {
                            totalCount = Integer.parseInt(totalCountStr);
                        } catch (NumberFormatException e) {
                            totalCount = 0;
                        }
                    }
                    int totalPages = Math.max(1, (totalCount + PAGE_SIZE - 1) / PAGE_SIZE);
                    int randomPage = 1 + RNG.nextInt(totalPages);
                    return session.set("totalPages", totalPages).set("randomPage", randomPage);
                }),

                exec(
                    http("Get random product page")
                        .get("/product?pageIndex=#{randomPage}&pageSize=#{pageSize}")
                        .check(status().is(200))
                )
                .pause(Duration.ofSeconds(3), Duration.ofSeconds(8))
            );

    {
        setUp(
        scenario.injectOpen(
            stressPeakUsers(1000).during(Duration.ofMinutes(3))
        )
        ).protocols(httpProtocol);
    }
}
