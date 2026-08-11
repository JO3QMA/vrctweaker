import { describe, it, expect, vi, beforeEach } from "vitest";
import { ElMessage } from "element-plus";
import { showToast } from "./showToast";

describe("showToast", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("delegates success to ElMessage.success", () => {
    const spy = vi.spyOn(ElMessage, "success").mockReturnValue({} as never);
    showToast.success("Saved");
    expect(spy).toHaveBeenCalledWith("Saved");
  });

  it("delegates error to ElMessage.error", () => {
    const spy = vi.spyOn(ElMessage, "error").mockReturnValue({} as never);
    showToast.error("Failed");
    expect(spy).toHaveBeenCalledWith("Failed");
  });

  it("delegates warning to ElMessage.warning", () => {
    const spy = vi.spyOn(ElMessage, "warning").mockReturnValue({} as never);
    showToast.warning("Careful");
    expect(spy).toHaveBeenCalledWith("Careful");
  });

  it("delegates info to ElMessage.info", () => {
    const spy = vi.spyOn(ElMessage, "info").mockReturnValue({} as never);
    showToast.info("Note");
    expect(spy).toHaveBeenCalledWith("Note");
  });
});
