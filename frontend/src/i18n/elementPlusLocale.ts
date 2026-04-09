import elementEn from "element-plus/es/locale/lang/en";
import elementJa from "element-plus/es/locale/lang/ja";
import elementKo from "element-plus/es/locale/lang/ko";
import elementZhCn from "element-plus/es/locale/lang/zh-cn";
import elementZhTw from "element-plus/es/locale/lang/zh-tw";

/** Maps app UI locale codes to Element Plus locale bundles. */
export function elementPlusLocaleFor(code: string) {
  switch (code) {
    case "ja":
    case "ja-JP":
      return elementJa;
    case "zh-CN":
      return elementZhCn;
    case "zh-TW":
      return elementZhTw;
    case "ko":
      return elementKo;
    default:
      return elementEn;
  }
}
