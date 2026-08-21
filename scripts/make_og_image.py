#!/usr/bin/env python3
"""Render the Open Graph preview card served at /v2/og-image.png.

Social scrapers do not execute JavaScript, so the /v2 SPA shell needs a
pre-rendered card referenced from static <meta> tags. Re-run this after
changing the copy or palette:

    python3 scripts/make_og_image.py

Output is committed so container builds do not need PIL.
"""

from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

# Facebook/Twitter/Slack all downscale from 1200x630; smaller sources get
# upscaled and look soft.
WIDTH, HEIGHT = 1200, 630

PARCHMENT = (242, 236, 225)
INK = (37, 28, 26)
DEEP_RED = (122, 24, 24)
MUTED = (106, 90, 84)
RULE = (214, 201, 189)

SERIF_BOLD = "/usr/share/fonts/truetype/liberation/LiberationSerif-Bold.ttf"
SERIF_REGULAR = "/usr/share/fonts/truetype/liberation/LiberationSerif-Regular.ttf"

TITLE = "Bible API"
SUBTITLE = "Full-text search across the King James Bible"
TRANSLATIONS = "KJV  ·  ASV  ·  NET  ·  GENEVA  ·  TYNDALE  ·  COVERDALE  ·  WEB"
FOOTER = "prsmusa.com/bible/v2"

ACCENT_BAR_W = 18
MARGIN_X = 96


def font(path: str, size: int) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(path, size)


def draw_tracked(draw, xy, text, fnt, fill, tracking: int) -> None:
    """PIL has no letter-spacing, so advance glyph by glyph."""
    x, y = xy
    for ch in text:
        draw.text((x, y), ch, font=fnt, fill=fill)
        x += draw.textlength(ch, font=fnt) + tracking


def main() -> None:
    img = Image.new("RGB", (WIDTH, HEIGHT), PARCHMENT)
    draw = ImageDraw.Draw(img)

    draw.rectangle([0, 0, ACCENT_BAR_W, HEIGHT], fill=DEEP_RED)
    draw.rectangle(
        [ACCENT_BAR_W + 40, 40, WIDTH - 40, HEIGHT - 40], outline=RULE, width=2
    )

    title_font = font(SERIF_BOLD, 132)
    subtitle_font = font(SERIF_REGULAR, 44)
    translations_font = font(SERIF_BOLD, 24)
    footer_font = font(SERIF_REGULAR, 30)

    x = ACCENT_BAR_W + MARGIN_X

    # anchor="ls" puts the baseline at y, which keeps the stack predictable
    # regardless of ascender height.
    draw.text((x, 300), TITLE, font=title_font, fill=INK, anchor="ls")
    draw.line([x, 340, x + 220, 340], fill=DEEP_RED, width=6)
    draw.text((x, 420), SUBTITLE, font=subtitle_font, fill=MUTED, anchor="ls")
    draw_tracked(draw, (x, 470), TRANSLATIONS, translations_font, DEEP_RED, 2)
    draw.text((x, HEIGHT - 72), FOOTER, font=footer_font, fill=MUTED, anchor="ls")

    out = Path(__file__).resolve().parent.parent / "web" / "public" / "og-image.png"
    out.parent.mkdir(parents=True, exist_ok=True)
    img.save(out, "PNG", optimize=True)
    print(f"wrote {out} ({out.stat().st_size} bytes, {WIDTH}x{HEIGHT})")


if __name__ == "__main__":
    main()
