import { describe, expect, it } from "vitest";
import {
  detectLang,
  isBinaryPath,
  isImageFile,
} from "./fileContentClassification";

describe("detectLang", () => {
  it.each([
    ["app.py", "python"],
    ["mod.rs", "rust"],
    ["main.go", "go"],
    ["index.ts", "typescript"],
    ["component.tsx", "tsx"],
    ["script.js", "javascript"],
    ["widget.jsx", "jsx"],
    ["config.json", "json"],
    ["values.yaml", "yaml"],
    ["values.yml", "yaml"],
    ["pyproject.toml", "toml"],
    ["README.md", "markdown"],
    ["run.sh", "bash"],
    ["profile.bash", "bash"],
    ["aliases.zsh", "bash"],
    ["query.sql", "sql"],
    ["page.html", "html"],
    ["styles.css", "css"],
  ])("maps %s to %s", (path, expected) => {
    expect(detectLang(path)).toBe(expected);
  });

  it("matches extensions case-insensitively", () => {
    expect(detectLang("Main.PY")).toBe("python");
    expect(detectLang("NOTES.MD")).toBe("markdown");
  });

  it("falls back to text", () => {
    expect(detectLang("data.unknownext")).toBe("text");
    expect(detectLang("LICENSE")).toBe("text");
  });

  it("maps Scala source files", () => {
    expect(detectLang("Service.scala")).toBe("scala");
    expect(detectLang("build.sc")).toBe("scala");
  });

  it("maps language-specific filenames", () => {
    expect(detectLang("Dockerfile")).toBe("dockerfile");
    expect(detectLang("path/to/Makefile")).toBe("make");
    expect(detectLang("CMakeLists.txt")).toBe("cmake");
  });

  it("maps the extended language set", () => {
    expect(detectLang("Main.kt")).toBe("kotlin");
    expect(detectLang("app.rb")).toBe("ruby");
    expect(detectLang("index.php")).toBe("php");
    expect(detectLang("View.swift")).toBe("swift");
    expect(detectLang("styles.scss")).toBe("scss");
    expect(detectLang("App.vue")).toBe("vue");
    expect(detectLang("schema.graphql")).toBe("graphql");
    expect(detectLang("Program.cs")).toBe("csharp");
  });
});

describe("isBinaryPath", () => {
  it.each([
    "logo.png",
    "photo.jpg",
    "scan.jpeg",
    "icon.ico",
    "archive.zip",
    "bundle.tar",
    "data.gz",
    "app.exe",
    "lib.so",
    "font.woff2",
    "clip.mp4",
    "module.pyc",
    "store.sqlite3",
  ])("classifies %s as binary", (path) => {
    expect(isBinaryPath(path)).toBe(true);
  });

  it.each(["app.py", "index.ts", "README.md", "config.json", "notes.txt"])(
    "classifies %s as text",
    (path) => expect(isBinaryPath(path)).toBe(false),
  );

  it("matches binary extensions case-insensitively", () => {
    expect(isBinaryPath("LOGO.PNG")).toBe(true);
  });

  it("treats extension-less paths as text", () => {
    expect(isBinaryPath("Dockerfile")).toBe(false);
  });
});

describe("isImageFile", () => {
  it.each([
    "logo.png",
    "photo.jpg",
    "scan.jpeg",
    "anim.gif",
    "icon.ico",
    "hero.webp",
    "next.avif",
    "diagram.svg",
  ])("classifies %s as an image", (path) => {
    expect(isImageFile(path)).toBe(true);
  });

  it.each(["app.py", "archive.zip", "clip.mp4", "font.woff2", "notes.txt"])(
    "classifies %s as a non-image",
    (path) => expect(isImageFile(path)).toBe(false),
  );

  it("matches image extensions case-insensitively", () => {
    expect(isImageFile("LOGO.PNG")).toBe(true);
  });

  it("uses content type as authoritative", () => {
    expect(isImageFile("blob", "image/png")).toBe(true);
    expect(isImageFile("data.txt", "image/jpeg")).toBe(true);
    expect(isImageFile("logo.png", "text/plain")).toBe(false);
    expect(isImageFile("photo.jpg", "application/octet-stream")).toBe(false);
  });

  it("falls back to the extension without content type", () => {
    expect(isImageFile("logo.png", null)).toBe(true);
    expect(isImageFile("logo.png", undefined)).toBe(true);
    expect(isImageFile("notes.txt", null)).toBe(false);
  });

  it("treats extension-less paths as non-images", () => {
    expect(isImageFile("Dockerfile")).toBe(false);
  });
});
