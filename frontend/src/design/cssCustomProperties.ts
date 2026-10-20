export type CssCustomProperty = {
  selector: string;
  name: string;
  value: string;
  group: string;
  description: string;
};

export type CssDeclaration = {
  selector: string;
  property: string;
  value: string;
  group: string;
  description: string;
};

export type ParsedStyleSheet = {
  customProperties: CssCustomProperty[];
  declarations: CssDeclaration[];
};

export type ResolvedCustomProperty = CssCustomProperty & {
  referencesVar: boolean;
  resolved: string;
};

type ReadDeclaration = {
  end: number;
  property: string;
  value: string;
  description: string;
};

function collapseWs(value: string): string {
  let out = "";
  let quote: '"' | "'" | null = null;
  let pendingSpace = false;
  for (let i = 0; i < value.length; i++) {
    const ch = value[i] ?? "";
    if (quote) {
      out += ch;
      if (ch === "\\" && i + 1 < value.length) {
        out += value[i + 1] ?? "";
        i++;
        continue;
      }
      if (ch === quote) quote = null;
      continue;
    }
    if (ch === '"' || ch === "'") {
      if (pendingSpace && out.length > 0) out += " ";
      pendingSpace = false;
      quote = ch;
      out += ch;
      continue;
    }
    if (/\s/.test(ch)) {
      if (out.length > 0) pendingSpace = true;
      continue;
    }
    if (pendingSpace) out += " ";
    pendingSpace = false;
    out += ch;
  }
  return out.trim();
}

/** Drop box-drawing rules and the whitespace they leave in section comments. */
function cleanComment(value: string): string {
  return collapseWs(value.replace(/─+/g, " "));
}

/** Collapse line-break padding inside functions, e.g. `calc( var(--a) * 1.4 )`. */
function tightenCssValue(value: string): string {
  return collapseWs(value).replace(/\(\s+/g, "(").replace(/\s+\)/g, ")");
}

function isOwnLineComment(
  css: string,
  commentStart: number,
  commentEnd: number,
): boolean {
  const lineStart = css.lastIndexOf("\n", commentStart - 1) + 1;
  const lineEndIdx = css.indexOf("\n", commentEnd);
  const lineEnd = lineEndIdx === -1 ? css.length : lineEndIdx;
  const before = css.slice(lineStart, commentStart);
  const after = css.slice(commentEnd, lineEnd);
  return before.trim() === "" && after.trim() === "";
}

function readDeclaration(css: string, start: number): ReadDeclaration {
  const len = css.length;
  let j = start;
  let property = "";
  while (j < len && /[A-Za-z0-9_-]/.test(css[j] ?? "")) {
    property += css[j];
    j++;
  }
  while (j < len && /\s/.test(css[j] ?? "")) j++;
  if ((css[j] ?? "") !== ":") {
    return { end: start + 1, property: "", value: "", description: "" };
  }
  j++;
  let value = "";
  let description = "";
  let paren = 0;
  while (j < len) {
    if (css.startsWith("/*", j)) {
      const end = css.indexOf("*/", j + 2);
      const stop = end === -1 ? len : end;
      description = collapseWs(css.slice(j + 2, stop));
      j = end === -1 ? len : end + 2;
      continue;
    }
    const ch = css[j] ?? "";
    if (ch === '"' || ch === "'") {
      value += ch;
      j++;
      while (j < len && css[j] !== ch) {
        value += css[j];
        if (css[j] === "\\") {
          j++;
          if (j < len) value += css[j];
        }
        j++;
      }
      if (j < len) {
        value += css[j];
        j++;
      }
      continue;
    }
    if (ch === "(") paren++;
    if (ch === ")" && paren > 0) paren--;
    if ((ch === ";" || ch === "}") && paren === 0) {
      if (ch === ";") {
        let k = j + 1;
        while (k < len && (css[k] === " " || css[k] === "\t")) k++;
        if (css.startsWith("/*", k)) {
          const commentEnd = css.indexOf("*/", k + 2);
          const stop = commentEnd === -1 ? len : commentEnd;
          const after = commentEnd === -1 ? len : commentEnd + 2;
          const lineEndIdx = css.indexOf("\n", after);
          const lineEnd = lineEndIdx === -1 ? len : lineEndIdx;
          if (css.slice(after, lineEnd).trim() === "") {
            description = collapseWs(css.slice(k + 2, stop));
            j = after;
            break;
          }
        }
        j++;
      }
      break;
    }
    value += ch;
    j++;
  }
  return {
    end: j,
    property: property.trim(),
    value: tightenCssValue(value),
    description,
  };
}

/** Parse custom properties and declarations, keeping source order and comments. */
export function parseStyleSheet(css: string): ParsedStyleSheet {
  const customProperties: CssCustomProperty[] = [];
  const declarations: CssDeclaration[] = [];
  const len = css.length;
  let i = 0;
  let depth = 0;
  let selectorBuf = "";
  let selector = "";
  let group = "";
  let pendingDescription = "";
  const groupStack: string[] = [];

  while (i < len) {
    if (css.startsWith("/*", i)) {
      const end = css.indexOf("*/", i + 2);
      const stop = end === -1 ? len : end;
      const text = cleanComment(css.slice(i + 2, stop));
      const commentEnd = end === -1 ? len : end + 2;
      if (isOwnLineComment(css, i, commentEnd)) {
        group = text;
        pendingDescription = "";
      } else {
        pendingDescription = text;
      }
      i = commentEnd;
      continue;
    }

    const ch = css[i] ?? "";
    if (ch === '"' || ch === "'") {
      const quote = ch;
      const chunk = [ch];
      i++;
      while (i < len && css[i] !== quote) {
        chunk.push(css[i] ?? "");
        if (css[i] === "\\") {
          i++;
          if (i < len) chunk.push(css[i] ?? "");
        }
        i++;
      }
      if (i < len) {
        chunk.push(css[i] ?? "");
        i++;
      }
      if (depth === 0) selectorBuf += chunk.join("");
      continue;
    }

    if (ch === "{") {
      if (depth === 0) {
        selector = collapseWs(selectorBuf);
        selectorBuf = "";
        groupStack.push(group);
      }
      depth++;
      i++;
      continue;
    }

    if (ch === "}") {
      if (depth > 0) depth--;
      i++;
      if (depth === 0) {
        const restored = groupStack.pop();
        if (restored !== undefined) group = restored;
        selector = "";
        pendingDescription = "";
      }
      continue;
    }

    if (depth === 0) {
      selectorBuf += ch;
      i++;
      continue;
    }

    if (depth === 1 && /[A-Za-z_-]/.test(ch)) {
      const decl = readDeclaration(css, i);
      i = decl.end;
      if (decl.property === "") continue;
      const description = decl.description || pendingDescription;
      pendingDescription = "";
      const row: CssDeclaration = {
        selector,
        property: decl.property,
        value: decl.value,
        group,
        description,
      };
      declarations.push(row);
      if (decl.property.startsWith("--")) {
        customProperties.push({
          selector,
          name: decl.property,
          value: decl.value,
          group,
          description,
        });
      }
      continue;
    }

    i++;
  }

  return { customProperties, declarations };
}

export function valueReferencesVar(value: string): boolean {
  return /var\(\s*--/.test(value);
}

/** Element whose cascade matches the declaration selector (`:root`, `html.dark`, …). */
export function elementForCssSelector(
  doc: Document,
  selector: string,
): Element | null {
  try {
    return doc.querySelector(selector);
  } catch {
    return null;
  }
}

export function readComputedCustomProperty(
  doc: Document,
  selector: string,
  name: string,
): string {
  const el = elementForCssSelector(doc, selector);
  if (!el) return "";
  return getComputedStyle(el).getPropertyValue(name).trim();
}

/**
 * Used length of a custom property. getComputedStyle().getPropertyValue does not
 * evaluate calc(); applying the variable as font-size does.
 */
export function readUsedLength(
  doc: Document,
  selector: string,
  name: string,
): string {
  const host = elementForCssSelector(doc, selector);
  if (!host) return "";
  const probe = doc.createElement("span");
  probe.style.fontSize = `var(${name})`;
  host.appendChild(probe);
  const used = getComputedStyle(probe).fontSize.trim();
  probe.remove();
  return used;
}

/** Prefer a used px length when the specified value is still a calc(). */
export function usedLengthIfComputed(specified: string, used: string): string {
  if (!specified.includes("calc(")) return specified;
  return /^-?\d+(\.\d+)?px$/.test(used) ? used : specified;
}

export function resolveCustomProperties(
  doc: Document,
  props: readonly CssCustomProperty[],
): ResolvedCustomProperty[] {
  return props.map((prop) => {
    const referencesVar = valueReferencesVar(prop.value);
    if (!referencesVar) {
      return { ...prop, referencesVar, resolved: "" };
    }
    const specified = readComputedCustomProperty(doc, prop.selector, prop.name);
    const resolved = usedLengthIfComputed(
      specified,
      readUsedLength(doc, prop.selector, prop.name),
    );
    return { ...prop, referencesVar, resolved };
  });
}

export function readUsedFont(
  doc: Document,
  className: string,
): { fontSize: string; lineHeight: string; fontWeight: string } {
  const el = doc.createElement("p");
  el.className = className;
  doc.body.appendChild(el);
  const cs = getComputedStyle(el);
  const used = {
    fontSize: cs.fontSize,
    lineHeight: cs.lineHeight,
    fontWeight: cs.fontWeight,
  };
  el.remove();
  return used;
}
