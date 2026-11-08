import { describe, it, expect, vi, beforeEach } from "vitest";
import { ElMessage } from "element-plus";
import { showToast } from "./showToast";
import { Z_INDEX_TOAST } from "../constants/elementPlusStack";

describe("showToast", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    document.body.innerHTML = "";
  });

  it("delegates success to ElMessage.success with toast zIndex option", () => {
    const spy = vi.spyOn(ElMessage, "success").mockReturnValue({} as never);
    showToast.success("Saved");
    expect(spy).toHaveBeenCalledWith({
      message: "Saved",
      zIndex: Z_INDEX_TOAST,
    });
  });

  it("delegates error to ElMessage.error with toast zIndex option", () => {
    const spy = vi.spyOn(ElMessage, "error").mockReturnValue({} as never);
    showToast.error("Failed");
    expect(spy).toHaveBeenCalledWith({
      message: "Failed",
      zIndex: Z_INDEX_TOAST,
    });
  });

  it("delegates warning to ElMessage.warning with toast zIndex option", () => {
    const spy = vi.spyOn(ElMessage, "warning").mockReturnValue({} as never);
    showToast.warning("Careful");
    expect(spy).toHaveBeenCalledWith({
      message: "Careful",
      zIndex: Z_INDEX_TOAST,
    });
  });

  it("delegates info to ElMessage.info with toast zIndex option", () => {
    const spy = vi.spyOn(ElMessage, "info").mockReturnValue({} as never);
    showToast.info("Note");
    expect(spy).toHaveBeenCalledWith({
      message: "Note",
      zIndex: Z_INDEX_TOAST,
    });
  });
});
