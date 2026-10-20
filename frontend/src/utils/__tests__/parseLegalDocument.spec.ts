import { describe, it, expect } from "vitest";
import {
  formatLegalLastUpdated,
  parseLegalDocument,
} from "../parseLegalDocument";

describe("parseLegalDocument", () => {
  it("reads last updated from a Japanese line and strips it", () => {
    const md = `# Title\n\n最終更新: 2026-10-14\n\nBody text.`;
    const parsed = parseLegalDocument(md);
    expect(parsed.lastUpdated).toBe("2026-10-14");
    expect(parsed.bodyMarkdown).not.toContain("最終更新");
    expect(parsed.bodyMarkdown).toContain("Body text.");
  });

  it("reads last updated from YAML frontmatter", () => {
    const md = `---\nlastUpdated: 2025-01-02\n---\n# Title\n\nBody.`;
    const parsed = parseLegalDocument(md);
    expect(parsed.lastUpdated).toBe("2025-01-02");
    expect(parsed.bodyMarkdown).toBe("Body.");
  });

  it("returns null when no date is present", () => {
    const parsed = parseLegalDocument("# Only\n\nContent.");
    expect(parsed.lastUpdated).toBeNull();
    expect(parsed.bodyMarkdown).toBe("Content.");
  });

  it("strips leading document title heading for in-app page h1", () => {
    const md = `# Privacy Policy\n\n最終更新: 2026-10-14\n\n## Section\n\nBody.`;
    const parsed = parseLegalDocument(md);
    expect(parsed.bodyMarkdown).not.toMatch(/^#\s/);
    expect(parsed.bodyMarkdown).toContain("## Section");
  });
});

describe("formatLegalLastUpdated", () => {
  it("formats for Japanese locale", () => {
    expect(formatLegalLastUpdated("2026-10-14", "ja")).toBe("2026年10月14日");
  });

  it("formats for English locale", () => {
    expect(formatLegalLastUpdated("2026-10-14", "en")).toContain("October");
  });
});
