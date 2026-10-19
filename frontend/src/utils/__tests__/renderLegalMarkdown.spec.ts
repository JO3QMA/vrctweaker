import { describe, it, expect } from "vitest";
import {
  EXTERNAL_LINK_CLASS,
  renderLegalMarkdown,
} from "../renderLegalMarkdown";

describe("renderLegalMarkdown", () => {
  it("renders headings and strips raw HTML", () => {
    const html = renderLegalMarkdown(
      "# Title\n\nHello **world**.\n\n<script>alert(1)</script>",
    );
    expect(html).toContain("<h1>");
    expect(html).toContain("Title");
    expect(html).toContain("<strong>world</strong>");
    expect(html).not.toContain("<script");
    expect(html).not.toContain("alert");
  });

  it("replaces safe links with external-link spans", () => {
    const html = renderLegalMarkdown(
      "See [Issues](https://github.com/org/repo/issues).",
    );
    expect(html).toContain(EXTERNAL_LINK_CLASS);
    expect(html).toContain("data-external-href=");
    expect(html).toContain(
      encodeURIComponent("https://github.com/org/repo/issues"),
    );
    expect(html).not.toContain("<a ");
    expect(html).not.toContain('href="https://');
  });

  it("does not emit clickable markup for javascript URLs", () => {
    const html = renderLegalMarkdown("[bad](javascript:alert(1))");
    expect(html).not.toContain("data-external-href");
    expect(html).not.toContain("javascript:");
  });

  it("sanitizes event handlers injected via markdown HTML", () => {
    const html = renderLegalMarkdown(
      '<img src=x onerror="alert(1)">\n\n[ok](https://example.com)',
    );
    expect(html).not.toContain("onerror");
    expect(html).not.toContain("<img");
    expect(html).toContain(EXTERNAL_LINK_CLASS);
  });
});
