import { test, expect } from "@playwright/test";
for (const outcome of ["success", "conflict", "invalid", "network"]) {
  test(`booking ${outcome}`, async ({ page }) => {
    await page.route("http://localhost:8080/**", async (route) => {
      const path = new URL(route.request().url()).pathname;
      if (path.endsWith("/facilities"))
        return route.fulfill({
          json: [{ id: 1, name: "CourtFlow Demo Badminton" }],
        });
      if (path.endsWith("/courts"))
        return route.fulfill({
          json: [{ id: 1, name: "Court A", price_sen: 2500 }],
        });
      if (path.endsWith("/availability"))
        return route.fulfill({
          json: [
            {
              start_at: "2026-11-02T02:00:00Z",
              end_at: "2026-11-02T03:00:00Z",
              available: true,
              price_sen: 2500,
            },
          ],
        });
      if (path.endsWith("/bookings")) {
        await new Promise((r) => setTimeout(r, 300));
        if (outcome === "network") return route.abort();
        return route.fulfill({
          status:
            outcome === "success" ? 201 : outcome === "conflict" ? 409 : 400,
          json:
            outcome === "success"
              ? { id: "demo-test", amount_sen: 2500 }
              : {
                  error:
                    outcome === "conflict"
                      ? "slot already booked"
                      : "invalid booking request",
                },
        });
      }
    });
    await page.goto("/");
    await page
      .getByRole("button", { name: "10:00 am Available", exact: false })
      .click();
    await page.getByRole("button", { name: "Confirm demo booking" }).click();
    await expect(
      page.getByRole("button", { name: "Confirming…" }),
    ).toBeDisabled();
    if (outcome === "success")
      await expect(
        page.getByRole("status").filter({ hasText: "Demo confirmed" }),
      ).toContainText("Demo confirmed");
    else await expect(page.locator("p[role=alert]")).toBeVisible();
  });
}
test("mobile layout and loading", async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.route("http://localhost:8080/**", (route) => route.abort());
  await page.goto("/");
  await expect(
    page.getByRole("heading", { name: "Your court. Your next game." }),
  ).toBeVisible();
  await expect(page.locator("p[role=alert]")).toBeVisible();
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
  await page.screenshot({ path: "test-results/mobile.png", fullPage: true });
});
