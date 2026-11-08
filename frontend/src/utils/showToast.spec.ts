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

  async function flushToastZIndexPatches(): Promise<void> {
    await new Promise<void>((resolve) => queueMicrotask(resolve));
    await new Promise<void>((resolve) => {
      requestAnimationFrame(() => resolve());
    });
    await new Promise<void>((resolve) => {
      requestAnimationFrame(() => resolve());
    });
  }

  function stubElMessageAddToast(
    spy: ReturnType<typeof vi.spyOn>,
    seedZIndex: number,
  ): void {
    spy.mockImplementation(() => {
      const el = document.createElement("div");
      el.className = "el-message";
      el.style.zIndex = String(seedZIndex);
      document.body.appendChild(el);
      return {} as never;
    });
  }

  it("patches z-index on every .el-message after multiple toasts", async () => {
    const successSpy = vi.spyOn(ElMessage, "success");
    const warningSpy = vi.spyOn(ElMessage, "warning");
    const errorSpy = vi.spyOn(ElMessage, "error");
    stubElMessageAddToast(successSpy, 2001);
    stubElMessageAddToast(warningSpy, 2002);
    stubElMessageAddToast(errorSpy, 2003);

    showToast.success("a");
    showToast.warning("b");
    showToast.error("c");
    await flushToastZIndexPatches();

    const zIndexes = [...document.querySelectorAll(".el-message")].map((el) =>
      el instanceof HTMLElement ? el.style.zIndex : "",
    );
    expect(zIndexes).toEqual(["4000", "4000", "4000"]);
  });
});
