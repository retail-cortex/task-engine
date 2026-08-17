#!/usr/bin/env python3
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
Single-Page Documentation Generator & Standalone HTML Studio.

Traverses all Hugo markdown documentation files, parses frontmatter weights,
resolves Hugo {{< ref >}} cross-references into local anchor links, pre-renders
all Markdown content into static HTML, and outputs:
1. A standalone, responsive, interactive single-page HTML documentation studio (docs_single_page.html).
2. A consolidated single-page Markdown specification manual (docs_bundle.md).
3. A clean LLM context text bundle (llms-full.txt).
"""

from dataclasses import dataclass
from pathlib import Path
import argparse
import html
import json
import re
import sys
from typing import List, Dict, Optional, Tuple


@dataclass
class DocPage:
    filename: str
    path: Path
    title: str
    weight: int
    content: str
    slug: str
    chapter_num: int = 0


def parse_frontmatter(file_path: Path) -> Tuple[Dict[str, str], str]:
    """Extracts frontmatter key-values and raw markdown body from a document."""
    raw_text: str = file_path.read_text(encoding="utf-8")
    
    fm_match = re.match(r"^---\s*\n(.*?)\n---\s*\n(.*)$", raw_text, flags=re.DOTALL)
    if not fm_match:
        return {}, raw_text

    fm_raw: str = fm_match.group(1)
    body: str = fm_match.group(2)
    meta: Dict[str, str] = {}

    for line in fm_raw.splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if ":" in line:
            key, val = line.split(":", 1)
            key = key.strip()
            val = val.strip().strip('"').strip("'")
            meta[key] = val

    return meta, body


def slugify(text: str) -> str:
    """Converts a title or header string into a clean URL/anchor slug."""
    slug: str = re.sub(r"[^\w\s-]", "", text.lower())
    slug = re.sub(r"[\s_-]+", "-", slug).strip("-")
    return slug


def load_doc_pages(content_dir: Path) -> List[DocPage]:
    """Discovers and parses all markdown pages in content_dir sorted by frontmatter weight."""
    docs_dir: Path = content_dir / "docs"
    if not docs_dir.exists():
        docs_dir = content_dir

    pages: List[DocPage] = []
    
    for file_path in docs_dir.glob("*.md"):
        if file_path.name.startswith("_"):
            continue
        
        meta, body = parse_frontmatter(file_path)
        title: str = meta.get("title", file_path.stem.replace("_", " ").title())
        weight_str: str = meta.get("weight", "999")
        try:
            weight = int(weight_str)
        except ValueError:
            weight = 999

        slug: str = slugify(f"chapter-{file_path.stem}")
        pages.append(
            DocPage(
                filename=file_path.name,
                path=file_path,
                title=title,
                weight=weight,
                content=body,
                slug=slug,
            )
        )

    # Sort pages stably by weight, then by filename
    pages.sort(key=lambda p: (p.weight, p.filename))

    for idx, page in enumerate(pages, start=1):
        page.chapter_num = idx
        page.slug = slugify(f"chapter-{page.chapter_num}-{page.title}")

    return pages


def resolve_cross_references(
    body: str, 
    page_slug_map: Dict[str, str]
) -> str:
    """Transforms Hugo ref shortcodes and relative doc links into internal anchor jumps."""
    
    def replace_hugo_ref(match: re.Match[str]) -> str:
        ref_target: str = match.group(1).strip().strip('"').strip("'")
        target_stem: str = Path(ref_target).stem
        target_slug: Optional[str] = page_slug_map.get(target_stem) or page_slug_map.get(ref_target)
        if target_slug:
            return f"#{target_slug}"
        return f"#{slugify(target_stem)}"

    body = re.sub(r'\{\{<\s*ref\s+["\']?([^"\'>\s]+)["\']?\s*>\}\}', replace_hugo_ref, body)
    body = re.sub(r'\{\{<\s*relref\s+["\']?([^"\'>\s]+)["\']?\s*>\}\}', replace_hugo_ref, body)

    # Demote top-level H1 headings inside the chapter to H2 to preserve single-page hierarchy
    lines: List[str] = []
    for line in body.splitlines():
        if line.startswith("# "):
            lines.append("## " + line[2:])
        else:
            lines.append(line)
    
    return "\n".join(lines)


def generate_single_page_markdown(pages: List[DocPage]) -> str:
    """Combines all pages into a unified single-page master specification manual."""
    page_slug_map: Dict[str, str] = {p.filename: p.slug for p in pages}
    for p in pages:
        page_slug_map[p.path.stem] = p.slug

    doc_parts: List[str] = []

    # Document Header
    doc_parts.append("# Enterprise Task Engine: Master Architectural Specification")
    doc_parts.append("> **Unified Single-Page Technical Reference & Specification Manual**")
    doc_parts.append("> *Copyright 2026 Google LLC — Distributed under the Apache License, Version 2.0*")
    doc_parts.append("")
    doc_parts.append("---")
    doc_parts.append("")

    # Master Table of Contents
    doc_parts.append("## Master Table of Contents")
    doc_parts.append("")
    for page in pages:
        doc_parts.append(f"- [Chapter {page.chapter_num}: {page.title}](#{page.slug})")
    doc_parts.append("")
    doc_parts.append("---")
    doc_parts.append("")

    # Document Chapters
    for page in pages:
        resolved_content: str = resolve_cross_references(page.content, page_slug_map)
        
        doc_parts.append(f"<a id=\"{page.slug}\"></a>")
        doc_parts.append(f"# Chapter {page.chapter_num}: {page.title}")
        doc_parts.append("")
        doc_parts.append(resolved_content.strip())
        doc_parts.append("")
        doc_parts.append("---")
        doc_parts.append("")

    # Footer
    doc_parts.append("## End of Specification")
    doc_parts.append("*Enterprise Task Engine monorepo master specification compiled automatically.*")

    return "\n".join(doc_parts)


def format_inline_markdown(text: str) -> str:
    """Renders bold, italics, inline code, links, and escapes safely."""
    # First protect inline code spans
    code_spans: List[str] = []
    def save_code(m: re.Match[str]) -> str:
        code_spans.append(m.group(1))
        return f"___CODE_SPAN_{len(code_spans)-1}___"

    text = re.sub(r"`([^`]+)`", save_code, text)

    # HTML escape rest of text
    text = html.escape(text)

    # Links: [text](url)
    text = re.sub(r'\[(.*?)\]\((.*?)\)', r'<a href="\2">\1</a>', text)

    # Bold: **text** or __text__
    text = re.sub(r'\*\*(.*?)\*\*', r'<strong>\1</strong>', text)
    text = re.sub(r'__(.*?)__', r'<strong>\1</strong>', text)

    # Italics: *text* or _text_
    text = re.sub(r'\*(.*?)\*', r'<em>\1</em>', text)
    text = re.sub(r'_(.*?)_', r'<em>\1</em>', text)

    # Restore code spans with HTML escaping
    for idx, c in enumerate(code_spans):
        text = text.replace(f"___CODE_SPAN_{idx}___", f"<code>{html.escape(c)}</code>")

    return text


def markdown_to_html_blocks(md_text: str) -> str:
    """Pre-renders markdown document into clean static HTML elements."""
    lines: List[str] = md_text.splitlines()
    html_out: List[str] = []
    
    i: int = 0
    n: int = len(lines)
    
    while i < n:
        line: str = lines[i]
        stripped: str = line.strip()

        if not stripped:
            i += 1
            continue

        # 1. Raw HTML anchor
        if stripped.startswith("<a id=") and stripped.endswith("</a>"):
            html_out.append(stripped)
            i += 1
            continue

        # 2. Fenced Code Blocks (``` or ````)
        if stripped.startswith("```"):
            lang: str = stripped.lstrip("`").strip().lower()
            code_lines: List[str] = []
            i += 1
            while i < n and not lines[i].strip().startswith("```"):
                code_lines.append(lines[i])
                i += 1
            if i < n:
                i += 1  # Skip closing fence
            
            raw_code: str = "\n".join(code_lines)
            if lang == "mermaid":
                html_out.append(f'<div class="mermaid-container"><div class="mermaid">{html.escape(raw_code)}</div></div>')
            else:
                css_lang = lang if lang else "plaintext"
                html_out.append(
                    f'<pre><button class="copy-btn" onclick="copyCode(this)">Copy</button><code class="language-{css_lang}">{html.escape(raw_code)}</code></pre>'
                )
            continue

        # 3. Headings (# H1, ## H2, ### H3, #### H4)
        heading_match = re.match(r"^(#{1,6})\s+(.*)$", stripped)
        if heading_match:
            level: int = len(heading_match.group(1))
            h_title: str = heading_match.group(2).strip()
            h_slug: str = slugify(h_title)
            formatted_title: str = format_inline_markdown(h_title)
            
            if level <= 3:
                html_out.append(f'<h{level} id="{h_slug}">{formatted_title}<a class="anchor" href="#{h_slug}">#</a></h{level}>')
            else:
                html_out.append(f'<h{level} id="{h_slug}">{formatted_title}</h{level}>')
            i += 1
            continue

        # 4. Horizontal Rules
        if re.match(r"^(\-{3,}|\*{3,}|_{3,})$", stripped):
            html_out.append("<hr>")
            i += 1
            continue

        # 5. Tables
        if stripped.startswith("|") and stripped.endswith("|"):
            table_rows: List[str] = []
            while i < n and lines[i].strip().startswith("|") and lines[i].strip().endswith("|"):
                table_rows.append(lines[i].strip())
                i += 1
            
            if len(table_rows) >= 2:
                html_out.append('<table>')
                # Header
                header_cols = [c.strip() for c in table_rows[0].strip("|").split("|")]
                html_out.append('<thead><tr>')
                for col in header_cols:
                    html_out.append(f'<th>{format_inline_markdown(col)}</th>')
                html_out.append('</tr></thead>')
                
                # Check if row 1 is separator (---)
                start_row = 1
                if re.match(r"^\|[\s:\-\|]+\|$", table_rows[1]):
                    start_row = 2

                # Body
                html_out.append('<tbody>')
                for r in table_rows[start_row:]:
                    cols = [c.strip() for c in r.strip("|").split("|")]
                    html_out.append('<tr>')
                    for col in cols:
                        html_out.append(f'<td>{format_inline_markdown(col)}</td>')
                    html_out.append('</tr>')
                html_out.append('</tbody></table>')
            continue

        # 6. GitHub-style Alert Callouts & Blockquotes
        if stripped.startswith(">"):
            quote_lines: List[str] = []
            while i < n and lines[i].strip().startswith(">"):
                quote_lines.append(re.sub(r"^>\s?", "", lines[i].strip()))
                i += 1
            
            full_quote: str = "\n".join(quote_lines).strip()
            alert_match = re.match(r"^\[!(NOTE|WARNING|IMPORTANT|CAUTION|TIP)\]\s*(.*)$", full_quote, flags=re.DOTALL)
            if alert_match:
                alert_type = alert_match.group(1).upper()
                alert_body = alert_match.group(2).strip()
                html_out.append(f'<div class="alert-callout alert-{alert_type.lower()}">')
                html_out.append(f'<div class="alert-callout-title">{alert_type}</div>')
                html_out.append(f'<div>{format_inline_markdown(alert_body)}</div>')
                html_out.append('</div>')
            else:
                html_out.append(f'<blockquote>{format_inline_markdown(full_quote)}</blockquote>')
            continue

        # 7. Unordered / Ordered Lists
        if re.match(r"^(\*|-|\+|\d+\.)\s+", stripped):
            is_ordered = bool(re.match(r"^\d+\.\s+", stripped))
            tag = "ol" if is_ordered else "ul"
            html_out.append(f'<{tag}>')
            
            while i < n and re.match(r"^(\*|-|\+|\d+\.)\s+", lines[i].strip()):
                item_text = re.sub(r"^(\*|-|\+|\d+\.)\s+", "", lines[i].strip())
                html_out.append(f'<li>{format_inline_markdown(item_text)}</li>')
                i += 1
            
            html_out.append(f'</{tag}>')
            continue

        # 8. Standard Paragraph
        para_lines: List[str] = []
        while i < n and lines[i].strip() and not lines[i].strip().startswith(("#", "```", "|", ">", "* ", "- ", "+ ", "1. ", "<a ")):
            para_lines.append(lines[i].strip())
            i += 1
        
        if para_lines:
            html_out.append(f'<p>{format_inline_markdown(" ".join(para_lines))}</p>')

    return "\n".join(html_out)


def generate_single_page_html(pages: List[DocPage], markdown_content: str) -> str:
    """Generates a standalone, rich, pre-rendered single-page HTML documentation studio."""
    
    # Pre-render the markdown content into full static HTML
    pre_rendered_html: str = markdown_to_html_blocks(markdown_content)

    # Generate JSON TOC metadata for the sidebar
    toc_entries: List[Dict[str, str]] = []
    for page in pages:
        toc_entries.append({
            "id": page.slug,
            "title": f"Chapter {page.chapter_num}: {page.title}",
            "shortTitle": page.title,
            "chapterNum": str(page.chapter_num)
        })

    toc_json: str = json.dumps(toc_entries)

    html_template = f"""<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Enterprise Task Engine — Master Specification Manual</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Google+Sans:wght@400;500;700&family=Roboto+Mono:wght@400;500;700&family=Roboto:wght@300;400;500;700&display=swap" rel="stylesheet">
  
  <!-- Syntax Highlighting Styles -->
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css" id="hljs-theme">

  <style>
    :root {{
      --bg-primary: #0f172a;
      --bg-surface: #1e293b;
      --bg-card: rgba(30, 41, 59, 0.7);
      --bg-code: #0b1120;
      --border-color: rgba(255, 255, 255, 0.1);
      --border-focus: #38bdf8;
      --text-primary: #f8fafc;
      --text-secondary: #94a3b8;
      --text-muted: #64748b;
      --color-primary: #38bdf8;
      --color-accent: #818cf8;
      --color-success: #34d399;
      --color-warning: #fbbf24;
      --color-danger: #f87171;
      --sidebar-width: 320px;
      --header-height: 64px;
      --shadow-elevation: 0 10px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.5);
    }}

    [data-theme="light"] {{
      --bg-primary: #f8fafc;
      --bg-surface: #ffffff;
      --bg-card: rgba(255, 255, 255, 0.9);
      --bg-code: #f1f5f9;
      --border-color: rgba(0, 0, 0, 0.1);
      --border-focus: #0284c7;
      --text-primary: #0f172a;
      --text-secondary: #475569;
      --text-muted: #94a3b8;
      --color-primary: #0284c7;
      --color-accent: #6366f1;
      --color-success: #059669;
      --color-warning: #d97706;
      --color-danger: #dc2626;
      --shadow-elevation: 0 10px 25px -5px rgba(0, 0, 0, 0.05), 0 8px 10px -6px rgba(0, 0, 0, 0.05);
    }}

    * {{
      box-sizing: border-box;
      margin: 0;
      padding: 0;
    }}

    body {{
      font-family: 'Roboto', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      background-color: var(--bg-primary);
      color: var(--text-primary);
      line-height: 1.65;
      font-size: 15px;
      overflow-x: hidden;
      transition: background-color 0.25s ease, color 0.25s ease;
    }}

    /* Scroll Progress Bar */
    #progress-bar {{
      position: fixed;
      top: 0;
      left: 0;
      height: 3px;
      background: linear-gradient(90deg, var(--color-primary), var(--color-accent));
      width: 0%;
      z-index: 1000;
      transition: width 0.1s ease-out;
    }}

    /* Top Navigation Header */
    header.app-header {{
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      height: var(--header-height);
      background-color: var(--bg-surface);
      border-bottom: 1px solid var(--border-color);
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 24px;
      z-index: 900;
      backdrop-filter: blur(12px);
    }}

    .brand-logo {{
      display: flex;
      align-items: center;
      gap: 12px;
      font-family: 'Google Sans', sans-serif;
      font-weight: 700;
      font-size: 18px;
      color: var(--text-primary);
      text-decoration: none;
    }}

    .brand-logo .badge {{
      font-size: 11px;
      padding: 2px 8px;
      border-radius: 9999px;
      background-color: rgba(56, 189, 248, 0.15);
      color: var(--color-primary);
      border: 1px solid rgba(56, 189, 248, 0.3);
      font-weight: 500;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }}

    .header-actions {{
      display: flex;
      align-items: center;
      gap: 12px;
    }}

    .btn {{
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 8px 14px;
      border-radius: 8px;
      font-size: 13px;
      font-weight: 500;
      cursor: pointer;
      border: 1px solid var(--border-color);
      background-color: var(--bg-surface);
      color: var(--text-primary);
      transition: all 0.2s ease;
    }}

    .btn:hover {{
      background-color: var(--bg-primary);
      border-color: var(--color-primary);
      color: var(--color-primary);
    }}

    .btn-primary {{
      background-color: var(--color-primary);
      color: #0f172a;
      border-color: var(--color-primary);
      font-weight: 700;
    }}

    .btn-primary:hover {{
      background-color: #7dd3fc;
      color: #0f172a;
    }}

    /* Main Container Layout */
    .app-container {{
      display: flex;
      margin-top: var(--header-height);
      min-height: calc(100vh - var(--header-height));
    }}

    /* Sidebar Navigation */
    aside.sidebar {{
      width: var(--sidebar-width);
      position: fixed;
      top: var(--header-height);
      bottom: 0;
      left: 0;
      background-color: var(--bg-surface);
      border-right: 1px solid var(--border-color);
      overflow-y: auto;
      padding: 20px 16px;
      display: flex;
      flex-direction: column;
      gap: 16px;
      z-index: 800;
    }}

    .search-box {{
      position: relative;
    }}

    .search-box input {{
      width: 100%;
      padding: 10px 12px 10px 36px;
      border-radius: 8px;
      border: 1px solid var(--border-color);
      background-color: var(--bg-primary);
      color: var(--text-primary);
      font-size: 13px;
      outline: none;
      transition: border-color 0.2s;
    }}

    .search-box input:focus {{
      border-color: var(--border-focus);
    }}

    .search-box svg {{
      position: absolute;
      left: 12px;
      top: 50%;
      transform: translateY(-50%);
      width: 16px;
      height: 16px;
      color: var(--text-muted);
    }}

    .toc-tree {{
      list-style: none;
      display: flex;
      flex-direction: column;
      gap: 4px;
    }}

    .toc-item a {{
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 12px;
      border-radius: 6px;
      color: var(--text-secondary);
      text-decoration: none;
      font-size: 13px;
      font-weight: 500;
      transition: all 0.15s ease;
      line-height: 1.4;
    }}

    .toc-item a:hover {{
      background-color: rgba(56, 189, 248, 0.08);
      color: var(--color-primary);
    }}

    .toc-item.active a {{
      background-color: rgba(56, 189, 248, 0.15);
      color: var(--color-primary);
      font-weight: 700;
      border-left: 3px solid var(--color-primary);
    }}

    .toc-chapter-num {{
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 20px;
      height: 20px;
      border-radius: 4px;
      background-color: var(--bg-primary);
      font-size: 11px;
      font-weight: 700;
      color: var(--text-muted);
      flex-shrink: 0;
    }}

    .toc-item.active .toc-chapter-num {{
      background-color: var(--color-primary);
      color: #0f172a;
    }}

    /* Main Article Content */
    main.content {{
      margin-left: var(--sidebar-width);
      flex: 1;
      max-width: calc(100vw - var(--sidebar-width));
      padding: 40px 64px 120px 64px;
    }}

    .markdown-body {{
      max-width: 960px;
      margin: 0 auto;
    }}

    /* Typography & Hierarchy */
    h1, h2, h3, h4, h5, h6 {{
      font-family: 'Google Sans', sans-serif;
      font-weight: 700;
      color: var(--text-primary);
      margin-top: 1.8em;
      margin-bottom: 0.6em;
      line-height: 1.3;
      scroll-margin-top: calc(var(--header-height) + 24px);
    }}

    h1 {{
      font-size: 2.2rem;
      border-bottom: 1px solid var(--border-color);
      padding-bottom: 12px;
      margin-top: 2.5em;
      color: var(--color-primary);
    }}

    .markdown-body > h1:first-of-type {{
      margin-top: 0.5em;
    }}

    h2 {{
      font-size: 1.5rem;
      border-bottom: 1px solid var(--border-color);
      padding-bottom: 8px;
    }}

    h3 {{
      font-size: 1.25rem;
    }}

    p, ul, ol {{
      margin-bottom: 1.2em;
      color: var(--text-secondary);
    }}

    ul, ol {{
      padding-left: 24px;
    }}

    li {{
      margin-bottom: 0.4em;
    }}

    strong {{
      color: var(--text-primary);
    }}

    a {{
      color: var(--color-primary);
      text-decoration: none;
      transition: color 0.15s;
    }}

    a:hover {{
      text-decoration: underline;
    }}

    .anchor {{
      opacity: 0;
      margin-left: 8px;
      color: var(--text-muted);
      text-decoration: none;
      font-weight: 400;
      transition: opacity 0.2s;
    }}

    h1:hover .anchor, h2:hover .anchor, h3:hover .anchor {{
      opacity: 1;
    }}

    hr {{
      border: 0;
      height: 1px;
      background-color: var(--border-color);
      margin: 2.5em 0;
    }}

    /* Code Blocks */
    pre {{
      background-color: var(--bg-code);
      border: 1px solid var(--border-color);
      border-radius: 10px;
      padding: 16px 20px;
      overflow-x: auto;
      margin-bottom: 1.5em;
      position: relative;
      font-family: 'Roboto Mono', monospace;
      font-size: 13.5px;
      box-shadow: var(--shadow-elevation);
    }}

    code:not(pre code) {{
      background-color: var(--bg-code);
      color: var(--color-primary);
      padding: 2px 6px;
      border-radius: 4px;
      font-family: 'Roboto Mono', monospace;
      font-size: 0.9em;
      border: 1px solid var(--border-color);
    }}

    .copy-btn {{
      position: absolute;
      top: 10px;
      right: 10px;
      background-color: var(--bg-surface);
      border: 1px solid var(--border-color);
      color: var(--text-muted);
      border-radius: 6px;
      padding: 4px 8px;
      font-size: 11px;
      cursor: pointer;
      opacity: 0;
      transition: opacity 0.2s, color 0.2s;
    }}

    pre:hover .copy-btn {{
      opacity: 1;
    }}

    .copy-btn:hover {{
      color: var(--text-primary);
      border-color: var(--color-primary);
    }}

    /* Tables */
    table {{
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 1.8em;
      background-color: var(--bg-surface);
      border-radius: 8px;
      overflow: hidden;
      border: 1px solid var(--border-color);
      font-size: 14px;
    }}

    th, td {{
      padding: 12px 16px;
      text-align: left;
      border-bottom: 1px solid var(--border-color);
    }}

    th {{
      background-color: var(--bg-surface);
      color: var(--text-primary);
      font-weight: 700;
      font-size: 13px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }}

    tr:nth-child(even) td {{
      background-color: rgba(255, 255, 255, 0.02);
    }}

    /* GitHub-style Alerts / Callouts */
    .alert-callout {{
      border-left: 4px solid var(--color-primary);
      background-color: rgba(56, 189, 248, 0.08);
      border-radius: 0 8px 8px 0;
      padding: 16px 20px;
      margin: 1.5em 0;
    }}

    .alert-callout.alert-note {{
      border-left-color: var(--color-primary);
      background-color: rgba(56, 189, 248, 0.08);
    }}

    .alert-callout.alert-important {{
      border-left-color: var(--color-accent);
      background-color: rgba(129, 140, 248, 0.08);
    }}

    .alert-callout.alert-warning {{
      border-left-color: var(--color-warning);
      background-color: rgba(251, 191, 36, 0.08);
    }}

    .alert-callout.alert-caution {{
      border-left-color: var(--color-danger);
      background-color: rgba(248, 113, 113, 0.08);
    }}

    .alert-callout.alert-tip {{
      border-left-color: var(--color-success);
      background-color: rgba(52, 211, 153, 0.08);
    }}

    .alert-callout-title {{
      font-weight: 700;
      margin-bottom: 4px;
      color: var(--text-primary);
      text-transform: uppercase;
      font-size: 12px;
      letter-spacing: 0.5px;
    }}

    /* Mermaid Container */
    .mermaid-container {{
      background-color: var(--bg-surface);
      border: 1px solid var(--border-color);
      border-radius: 12px;
      padding: 24px;
      margin: 2em 0;
      overflow-x: auto;
      text-align: center;
      box-shadow: var(--shadow-elevation);
    }}

    .mermaid-container svg {{
      max-width: 100%;
      height: auto;
    }}

    /* Floating Back to Top Button */
    #back-to-top {{
      position: fixed;
      bottom: 32px;
      right: 32px;
      width: 44px;
      height: 44px;
      border-radius: 50%;
      background-color: var(--color-primary);
      color: #0f172a;
      border: none;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      box-shadow: 0 4px 14px rgba(0,0,0,0.3);
      opacity: 0;
      visibility: hidden;
      transition: all 0.25s ease;
      z-index: 950;
    }}

    #back-to-top.visible {{
      opacity: 1;
      visibility: visible;
    }}

    #back-to-top:hover {{
      transform: translateY(-2px);
      box-shadow: 0 6px 20px rgba(56, 189, 248, 0.4);
    }}

    /* Print / PDF Stylesheet */
    @media print {{
      header.app-header,
      aside.sidebar,
      #progress-bar,
      #back-to-top,
      .header-actions {{
        display: none !important;
      }}

      body {{
        background: #ffffff !important;
        color: #000000 !important;
      }}

      main.content {{
        margin: 0 !important;
        padding: 0 !important;
        max-width: 100% !important;
      }}

      pre, table, .mermaid-container {{
        page-break-inside: avoid;
        box-shadow: none !important;
      }}

      h1 {{
        page-break-before: always;
      }}
    }}

    /* Responsive Mobile Layout */
    @media (max-width: 900px) {{
      aside.sidebar {{
        transform: translateX(-100%);
        transition: transform 0.3s ease;
      }}

      aside.sidebar.open {{
        transform: translateX(0);
      }}

      main.content {{
        margin-left: 0;
        max-width: 100vw;
        padding: 24px 20px;
      }}

      .menu-toggle {{
        display: inline-flex !important;
      }}
    }}

    .menu-toggle {{
      display: none;
      background: none;
      border: none;
      color: var(--text-primary);
      cursor: pointer;
      padding: 6px;
    }}
  </style>
</head>
<body>
  <div id="progress-bar"></div>

  <!-- Application Header -->
  <header class="app-header">
    <div style="display: flex; align-items: center; gap: 12px;">
      <button class="menu-toggle" id="sidebar-toggle" aria-label="Toggle Sidebar">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="3" y1="12" x2="21" y2="12"></line><line x1="3" y1="6" x2="21" y2="6"></line><line x1="3" y1="18" x2="21" y2="18"></line></svg>
      </button>
      <a href="#" class="brand-logo">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="color: var(--color-primary);"><path d="M12 2L2 7l10 5 10-5-10-5z"></path><path d="M2 17l10 5 10-5"></path><path d="M2 12l10 5 10-5"></path></svg>
        Enterprise Task Engine
        <span class="badge">v5.0 Specs</span>
      </a>
    </div>

    <div class="header-actions">
      <button class="btn" id="theme-toggle" title="Toggle Dark / Light Mode">
        <svg id="theme-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg>
        <span id="theme-text">Theme</span>
      </button>
      <button class="btn" onclick="window.print()" title="Print or Export to PDF">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 6 2 18 2 18 9"></polyline><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"></path><rect x="6" y="14" width="12" height="8"></rect></svg>
        Print / PDF
      </button>
      <a href="docs_bundle.md" download="Enterprise_Task_Engine_Specs.md" class="btn btn-primary" title="Download Markdown Bundle">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
        Export Markdown
      </a>
    </div>
  </header>

  <!-- Container -->
  <div class="app-container">
    <!-- Sidebar Navigation -->
    <aside class="sidebar" id="app-sidebar">
      <div class="search-box">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
        <input type="text" id="toc-filter" placeholder="Search chapters..." aria-label="Search Chapters">
      </div>
      <ul class="toc-tree" id="toc-list">
        <!-- Dynamically populated via JavaScript -->
      </ul>
    </aside>

    <!-- Main Content Area -->
    <main class="content">
      <article class="markdown-body" id="rendered-content">
{pre_rendered_html}
      </article>
    </main>
  </div>

  <!-- Floating Back to Top Button -->
  <button id="back-to-top" title="Back to Top" aria-label="Back to Top">
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="19" x2="12" y2="5"></line><polyline points="5 12 12 5 19 12"></polyline></svg>
  </button>

  <!-- Highlight.js & Mermaid Module (Progressive Enhancements) -->
  <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
  <script type="module">
    import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
    window.mermaid = mermaid;
    try {{
      mermaid.initialize({{
        startOnLoad: false,
        theme: document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'neutral',
        securityLevel: 'loose',
        flowchart: {{
          useMaxWidth: true,
          htmlLabels: true,
          curve: 'basis'
        }}
      }});
      mermaid.run();
    }} catch(e) {{
      console.warn('Mermaid initialization deferred:', e);
    }}
  </script>

  <!-- Application Controller Script -->
  <script>
    const TOC_DATA = {toc_json};

    document.addEventListener('DOMContentLoaded', () => {{
      const tocList = document.getElementById('toc-list');
      const progressBar = document.getElementById('progress-bar');
      const backToTopBtn = document.getElementById('back-to-top');
      const tocFilter = document.getElementById('toc-filter');
      const themeToggle = document.getElementById('theme-toggle');
      const sidebarToggle = document.getElementById('sidebar-toggle');
      const sidebar = document.getElementById('app-sidebar');

      // 1. Build Table of Contents
      function renderTOC(filterText = '') {{
        tocList.innerHTML = '';
        const lower = filterText.toLowerCase();
        TOC_DATA.forEach(item => {{
          if (filterText && !item.title.toLowerCase().includes(lower)) return;
          const li = document.createElement('li');
          li.className = 'toc-item';
          li.id = 'toc-nav-' + item.id;
          li.innerHTML = `
            <a href="#${{item.id}}">
              <span class="toc-chapter-num">${{item.chapterNum}}</span>
              <span>${{item.shortTitle}}</span>
            </a>
          `;
          tocList.appendChild(li);
        }});
      }}
      renderTOC();

      // 2. Filter TOC
      tocFilter.addEventListener('input', (e) => {{
        renderTOC(e.target.value);
      }});

      // 3. Highlight code blocks
      if (window.hljs) {{
        document.querySelectorAll('pre code').forEach((block) => {{
          hljs.highlightElement(block);
        }});
      }}

      // 4. Scroll Progress & ScrollSpy
      window.addEventListener('scroll', () => {{
        const winScroll = document.documentElement.scrollTop || document.body.scrollTop;
        const height = document.documentElement.scrollHeight - document.documentElement.clientHeight;
        const scrolled = (winScroll / height) * 100;
        progressBar.style.width = scrolled + '%';

        if (winScroll > 400) {{
          backToTopBtn.classList.add('visible');
        }} else {{
          backToTopBtn.classList.remove('visible');
        }}
      }});

      // ScrollSpy via IntersectionObserver
      const observer = new IntersectionObserver((entries) => {{
        entries.forEach(entry => {{
          if (entry.isIntersecting) {{
            const id = entry.target.getAttribute('id');
            if (id) {{
              document.querySelectorAll('.toc-item').forEach(el => el.classList.remove('active'));
              const activeItem = document.getElementById('toc-nav-' + id);
              if (activeItem) {{
                activeItem.classList.add('active');
                activeItem.scrollIntoView({{ block: 'nearest' }});
              }}
            }}
          }}
        }});
      }}, {{ rootMargin: '-80px 0px -70% 0px' }});

      document.querySelectorAll('a[id^="chapter-"]').forEach(anchor => {{
        observer.observe(anchor);
      }});

      // 5. Back to Top Click
      backToTopBtn.addEventListener('click', () => {{
        window.scrollTo({{ top: 0, behavior: 'smooth' }});
      }});

      // 6. Theme Toggle
      themeToggle.addEventListener('click', () => {{
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', newTheme);
        localStorage.setItem('doc-theme', newTheme);
        
        const hljsLink = document.getElementById('hljs-theme');
        if (hljsLink) {{
          hljsLink.href = newTheme === 'dark' 
            ? 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css'
            : 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github.min.css';
        }}
      }});

      const savedTheme = localStorage.getItem('doc-theme');
      if (savedTheme) {{
        document.documentElement.setAttribute('data-theme', savedTheme);
      }}

      // 7. Mobile Sidebar Toggle
      if (sidebarToggle) {{
        sidebarToggle.addEventListener('click', () => {{
          sidebar.classList.toggle('open');
        }});
      }}
    }});

    // Global Code Copy Helper
    window.copyCode = function(button) {{
      const pre = button.parentElement;
      const code = pre.querySelector('code');
      navigator.clipboard.writeText(code.innerText).then(() => {{
        button.innerText = 'Copied!';
        setTimeout(() => {{ button.innerText = 'Copy'; }}, 2000);
      }});
    }};
  </script>
</body>
</html>
"""
    return html_template


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate unified single-page documentation studio.")
    parser.add_argument(
        "--content-dir",
        type=Path,
        default=Path("docs/content"),
        help="Path to Hugo content directory (default: docs/content)",
    )
    parser.add_argument(
        "--output-md",
        type=Path,
        default=Path("docs/docs_bundle.md"),
        help="Destination path for single-page markdown output (default: docs/docs_bundle.md)",
    )
    parser.add_argument(
        "--output-llm",
        type=Path,
        default=Path("docs/llms-full.txt"),
        help="Destination path for LLM context text output (default: docs/llms-full.txt)",
    )
    parser.add_argument(
        "--output-html",
        type=Path,
        default=Path("docs/docs_single_page.html"),
        help="Destination path for single-page HTML output (default: docs/docs_single_page.html)",
    )

    args = parser.parse_args()

    content_dir: Path = args.content_dir.resolve()
    if not content_dir.exists():
        workspace_root: Path = Path(__file__).resolve().parent.parent
        content_dir = workspace_root / "docs" / "content"

    if not content_dir.exists():
        sys.stderr.write(f"Error: Content directory {content_dir} not found.\n")
        sys.exit(1)

    pages: List[DocPage] = load_doc_pages(content_dir)
    if not pages:
        sys.stderr.write(f"Error: No documentation pages found in {content_dir}.\n")
        sys.exit(1)

    markdown_bundle: str = generate_single_page_markdown(pages)
    html_bundle: str = generate_single_page_html(pages, markdown_bundle)

    # Ensure output parent directories exist
    args.output_md.parent.mkdir(parents=True, exist_ok=True)
    args.output_llm.parent.mkdir(parents=True, exist_ok=True)
    args.output_html.parent.mkdir(parents=True, exist_ok=True)

    args.output_md.write_text(markdown_bundle, encoding="utf-8")
    args.output_llm.write_text(markdown_bundle, encoding="utf-8")
    args.output_html.write_text(html_bundle, encoding="utf-8")

    print(f"Successfully generated single-page documentation ({len(pages)} chapters):")
    print(f"  -> Standalone HTML Studio: {args.output_html}")
    print(f"  -> Markdown Bundle:       {args.output_md}")
    print(f"  -> LLM Full Context:      {args.output_llm}")


if __name__ == "__main__":
    main()
